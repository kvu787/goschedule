package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/kvu787/goschedule/goschedule/backend"
	"github.com/kvu787/goschedule/goschedule/frontend"
	"github.com/kvu787/goschedule/goschedule/shared"
	"github.com/kvu787/goschedule/lib"
	_ "github.com/lib/pq"
)

var usage string = `goschedule is a tool for running the Go Schedule application.

Usage:

    goschedule <command> [flags]

Commands:

    setup    Setup the databases used by scrape to store records.
    scrape   Scrape the UW time schedule.
    web      Start the web application to view Go Schedule.
    help     Use "goschedule help [command] for more information about a command.

Built by Kevin Vu, 2013.`

var setupHelp string = `Usage:

	goschedule setup <create|teardown> --config=<path to config>

Examples:

	'goschedule setup create --config=./config.json': Reads the config and creates several databases for each defined schedule.
	'goschedule setup teardown --config=./config.json': Drops databases according to each defined schedule's name.

Note that 'goschedule setup teardown' will not work properly if you change the schedules in the JSON config after running 'goschedule setup create'.`

var scrapeHelp string = `Usage:

	goschedule scrape --config=<path to config>

Scrapes each schedule defined in the config and stores results in databases.
Expects that 'goschedule setup create' has been run to setup the databases.`

var webHelp string = `Usage:

	goschedule web --config=<path to config> --schedule=<schedule name> --fcgi=<port number>|--local=<port number>

Examples:
	
	'goschedule web --config=./config.json --schedule=aut2013 --local=8080': Starts Go Schedule web app that can be viewed in a browser at localhost:8080.
	'goschedule web --config=./config.json --schedule=aut2014 --fcgi=9000': Starts Go Schedule web app serving through fcgi on port 9000 (Used with an nginx server).

Note that the flags need to be in the order shown in 'Usage'.`

var dbSetupStatements = make([]string, 7)

func init() {
	for i, object := range []interface{}{goschedule.College{}, goschedule.Dept{}, goschedule.Class{}, goschedule.Sect{}} {
		dbSetupStatements[i] = goschedule.GenerateSchema(object)
	}
	dbSetupStatements[4] = "CREATE EXTENSION plpythonu"
	dbSetupStatements[5] = wordScoreSqlFunc
	dbSetupStatements[6] = letterScoreSqlFunc
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}
	var flags []string
	if len(os.Args) > 2 {
		flags = os.Args[2:]
	}
	switch os.Args[1] {
	case "help":
		if len(os.Args) > 2 {
			switch command := os.Args[2]; command {
			case "setup":
				fmt.Println(setupHelp)
				os.Exit(0)
			case "scrape":
				fmt.Println(scrapeHelp)
				os.Exit(0)
			case "web":
				fmt.Println(webHelp)
				os.Exit(0)
			default:
				fmt.Printf("ERROR: help not implemented for %q\n", command)
				os.Exit(1)
			}
		}
		fmt.Println(usage)
		os.Exit(0)
	case "setup":
		handleSetup(flags)
	case "scrape":
		handleScrape(flags)
	case "web":
		handleWeb(flags)
	default:
		fmt.Println("ERROR: unrecognized arguments")
		fmt.Println(usage)
		os.Exit(1)
	}
}

func handleSetup(args []string) {
	if len(args) < 1 {
		fmt.Println("ERROR: not enough arguments")
		os.Exit(1)
	}
	// create or drop databases
	var command string
	switch args[0] {
	case "create":
		command = "CREATE"
	case "teardown":
		command = "DROP"
	default:
		fmt.Println("unrecognized argument")
		os.Exit(1)
	}
	// load config
	conf := parseConfig(args[1:])
	user := conf.DbLogin["user"]
	password := conf.DbLogin["user"]
	dbname := conf.DbLogin["dbname"]
	// connect to superuser db
	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=require",
		user,
		dbname,
		password,
	))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// setup databases for each schedule
	for _, schedule := range conf.Schedules {
		for _, statement := range []string{
			fmt.Sprintf("%s DATABASE goschedule_%s_switch", command, schedule["name"]),
			fmt.Sprintf("%s DATABASE goschedule_%s_app1", command, schedule["name"]),
			fmt.Sprintf("%s DATABASE goschedule_%s_app2", command, schedule["name"]),
		} {
			if _, err := db.Exec(statement); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if command == "CREATE" {
			// load switch schema
			runSql("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				user,
				fmt.Sprintf("goschedule_%s_switch", schedule["name"]),
				password,
			), "CREATE TABLE switch_table ( switch_col int)", "INSERT INTO switch_table VALUES (1)")
			// load app db schemas and python functions
			for i := 1; i < 3; i++ {
				runSql("postgres", fmt.Sprintf(
					"user=%s dbname=%s password=%s sslmode=require",
					user,
					fmt.Sprintf("goschedule_%s_app%d", schedule["name"], i),
					password,
				), dbSetupStatements...)
			}
		}
	}
}

func handleScrape(args []string) {
	conf := parseConfig(args)
	user := conf.DbLogin["user"]
	password := conf.DbLogin["password"]
	dbname := conf.DbLogin["dbname"]
	for {
		// scrape for each schedule specified in config
		for _, schedule := range conf.Schedules {
			// connect to switch db
			switchDb, err := sql.Open("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				user,
				fmt.Sprintf("goschedule_%s_switch", schedule["name"]),
				password,
			))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			appNum, err := shared.GetSwitch(switchDb)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// reset app db
			runSql("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				user,
				dbname,
				password),
				fmt.Sprintf("DROP DATABASE goschedule_%s_app%d", schedule["name"], appNum),
				fmt.Sprintf("CREATE DATABASE goschedule_%s_app%d", schedule["name"], appNum))
			// connect to app db
			appDb, err := sql.Open("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				conf.DbLogin["user"],
				fmt.Sprintf("goschedule_%s_app%d", schedule["name"], appNum),
				conf.DbLogin["password"],
			))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// load app db with schemas
			runSql("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				user,
				fmt.Sprintf("goschedule_%s_app%d", schedule["name"], appNum),
				password,
			), dbSetupStatements...)
			// start scrape
			start := time.Now()
			fmt.Printf("Scraping %q using application database %d\n", schedule["url"], appNum)
			backend.Scrape(schedule["url"], conf.DepartmentDescriptionIndex, appDb)
			fmt.Println("Time taken:", time.Since(start))
			// flip db switch
			if err := flipSwitch(switchDb); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			switchDb.Close()
			fmt.Printf("Scrape for %q done\n", schedule["url"])
			// close connection to app db
			appDb.Close()
		}
		if !conf.LoopScraper {
			break
		}
		time.Sleep(time.Duration(conf.ScraperTimeout) * time.Minute)
	}
}

func handleWeb(flags []string) {
	if len(flags) < 3 {
		fmt.Println("ERROR: not enough arguments")
		os.Exit(1)
	}
	conf := parseConfig(flags[0:1])
	webFlags := flag.NewFlagSet("flags", flag.ContinueOnError)
	var local int
	var fcgi int
	var schedule string
	webFlags.IntVar(&local, "local", 0, "Local port number to serve and listen on.")
	webFlags.IntVar(&fcgi, "fcgi", 0, "Fcgi port number to serve and listen on.")
	webFlags.StringVar(&schedule, "schedule", "", "Name of the schedule (from config) to serve.")
	webFlags.Parse(flags[1:])
	var scheduleName string
	for _, s := range conf.Schedules {
		if schedule == s["name"] {
			scheduleName = schedule
		}
	}
	if scheduleName == "" {
		fmt.Printf("ERROR: cannot find schedule name %q in config\n", schedule)
		os.Exit(1)
	}
	if fcgi != 0 && local != 0 {
		fmt.Println("ERROR: cannot set both --fcgi and --local flags")
		os.Exit(1)
	}
	user := conf.DbLogin["user"]
	password := conf.DbLogin["password"]
	dbSwitch, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=require",
		user,
		fmt.Sprintf("goschedule_%s_switch", schedule),
		password,
	))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	appDbConnString := fmt.Sprintf("user=%s dbname=goschedule_%s_app%%d password=%s sslmode=require",
		user,
		schedule,
		password,
	)
	if local != 0 {
		fmt.Printf("Go Schedule frontend started locally on port %d\n", local)
		if err := frontend.Serve(appDbConnString, dbSwitch, true, conf.FrontendRoot, local); err != nil {
			fmt.Printf("ERROR in handleWeb: %v\n", err)
			os.Exit(1)
		}
	}
	if fcgi != 0 {
		fmt.Printf("Go Schedule frontend serving through fcgi on port %d\n", fcgi)
		if err := frontend.Serve(appDbConnString, dbSwitch, false, conf.FrontendRoot, fcgi); err != nil {
			fmt.Printf("ERROR in handleWeb: %v\n", err)
			os.Exit(1)
		}
	}
}

// config represents a JSON config file marshalled into a struct.
type config struct {
	FrontendRoot               string
	DepartmentDescriptionIndex string
	ScraperTimeout             int
	LoopScraper                bool
	DbLogin                    map[string]string
	Schedules                  []map[string]string
}

func findScheduleInConfig(name string, schedules []map[string]string) bool {
	for _, s := range schedules {
		if name == s["name"] {
			return true
		}
	}
	return false
}

// parseConfig will use the given args to try to load a file from the `--config` flag.
// If the config flag is not set, or if it cannot read the file path, or it encounters
// an error when unmarshalling the config from JSON, it will call os.Exit(1).
// Else, it will return a config struct.
func parseConfig(args []string) config {
	flagSet := flag.NewFlagSet("flags", flag.ContinueOnError)
	configPathPtr := flagSet.String("config", "", "Path to a JSON formatted config file.")
	flagSet.Parse(args)
	configPath := *configPathPtr
	if configPath == "" {
		fmt.Println("missing `--config` flag")
		os.Exit(1)
	}
	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("error reading config at %s\n", configPath)
		os.Exit(1)
	}
	parsedConfig := config{}
	if err := json.Unmarshal(rawConfig, &parsedConfig); err != nil {
		fmt.Printf("error parsing config file at %s to JSON\n", configPath)
		os.Exit(1)
	}
	return parsedConfig
}

// runSql is a convenience method that  connects to a database with the given
// driver and connection string and executes SQL statements. If it encounters
// an error, it prints the error and exits with status code 1.
func runSql(driver, connection string, statements ...string) {
	db, err := sql.Open(driver, connection)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// flipSwitch changes the value stored in the 'switch db' from 1 to 2
// or from 2 to 1.
func flipSwitch(db *sql.DB) error {
	currentSwitch, err := shared.GetSwitch(db)
	if err != nil {
		return err
	}
	var newSwitch int
	if currentSwitch == 1 {
		newSwitch = 2
	} else {
		newSwitch = 1
	}
	query := fmt.Sprintf(
		"UPDATE switch_table SET switch_col = %d WHERE switch_col = %d",
		newSwitch,
		currentSwitch,
	)
	_, err = db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

var wordScoreSqlFunc string = `CREATE OR REPLACE FUNCTION word_score (search text, phrase text)
  RETURNS integer
AS $$
    import string

    def word_similarity(search_term, word):
        search_term = string.lower(search_term)
        word = string.lower(word)
        if len(search_term) > len(word):
            return 0
        for i in range(len(search_term)):
            if search_term[i] != word[i]:
                return 0
        return len(search_term)

    search_terms = search.split(' ')
    score = 0
    for search_term in search_terms:
        for word in phrase.split(' '):
            if word_similarity(search_term, word) > 0:
                score += 1
    return score
$$ LANGUAGE plpythonu;`

var letterScoreSqlFunc string = `CREATE OR REPLACE FUNCTION letter_score (search text, phrase text)
  RETURNS integer
AS $$
    import string

    def word_similarity(search_term, word):
        search_term = string.lower(search_term)
        word = string.lower(word)
        if len(search_term) > len(word):
            return 0
        for i in range(len(search_term)):
            if search_term[i] != word[i]:
                return 0
        return len(search_term)

    search_terms = search.split(' ')
    score = 0
    for search_term in search_terms:
        for word in phrase.split(' '):
            score += word_similarity(search_term, word)
    return score
$$ LANGUAGE plpythonu;`
