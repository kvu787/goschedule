package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kvu787/goschedule"
	"github.com/kvu787/goschedule/goschedule/backend"
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

Built by Kevin Vu`

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
			fmt.Println("command help not implemented")
			os.Exit(1)
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
		fmt.Println("unrecognized arguments")
		fmt.Println(usage)
		os.Exit(1)
	}
}

func handleSetup(args []string) {
	if len(args) < 2 {
		fmt.Println("not enough arguments")
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
	config := parseConfig(args[1:])
	// connect to superuser db
	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=require",
		config.DbLogin["user"],
		config.DbLogin["dbname"],
		config.DbLogin["password"],
	))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// setup databases for each schedule
	for _, schedule := range config.Schedules {
		for _, statement := range []string{
			fmt.Sprintf("%s DATABASE goschedule_%s_switch", command, schedule["name"]),
			fmt.Sprintf("%s DATABASE goschedule_%s_app1", command, schedule["name"]),
			fmt.Sprintf("%s DATABASE goschedule_%s_app2", command, schedule["name"]),
		} {
			if _, err := db.Exec(statement); err != nil {
				fmt.Println(err)
			}
		}
		// load db schemas if in create mode
		if command == "CREATE" {
			// load switch schema
			dbSwitch, err := sql.Open("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				config.DbLogin["user"],
				fmt.Sprintf("goschedule_%s_switch", schedule["name"]),
				config.DbLogin["password"],
			))
			if err != nil {
				fmt.Println(err)
			}
			for _, statement := range []string{
				"CREATE TABLE switch_table ( switch_col int)",
				"INSERT INTO switch_table VALUES (1)",
			} {
				if _, err := dbSwitch.Exec(statement); err != nil {
					fmt.Println(err)
				}
			}
			// load app db schemas
			for i := 1; i < 3; i++ {
				dbApp, err := sql.Open("postgres", fmt.Sprintf(
					"user=%s dbname=%s password=%s sslmode=require",
					config.DbLogin["user"],
					fmt.Sprintf("goschedule_%s_app%d", schedule["name"], i),
					config.DbLogin["password"],
				))
				if err != nil {
					fmt.Println(err)
				}
				objects := []interface{}{goschedule.College{}, goschedule.Dept{}, goschedule.Class{}, goschedule.Sect{}}
				for _, object := range objects {
					if _, err := dbApp.Exec(goschedule.GenerateSchema(object)); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}

func handleScrape(args []string) {
	config := parseConfig(args)
	for _, schedule := range config.Schedules {
		// connect to db
		db, err := sql.Open("postgres", fmt.Sprintf(
			"user=%s dbname=%s password=%s sslmode=require",
			config.DbLogin["user"],
			fmt.Sprintf("goschedule_%s_app1", schedule["name"]),
			config.DbLogin["password"],
		))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Scraping", schedule["url"])
		backend.Scrape(schedule["url"], db)
		break // use only the first schedule
	}
}

func handleWeb(flags []string) {
	// fcgi := flagSet.Bool("fcgi", false, "")
	// flagSet.Parse(flags)
	// fmt.Println(*fcgi)
}

type config struct {
	DepartmentDescriptionIndex string
	LoopScraper                bool
	DbLogin                    map[string]string
	Schedules                  []map[string]string
}

// parseConfig will use the given args to try to load a file from the `--config` flag.
// If the config flag is not set, or if it cannot read the file path, or it encounters
// an error when unmarshalling the config from JSON, it will call os.Exit(1).
// Else, it will return a jsonConfig struct.
func parseConfig(args []string) config {
	var flagSet = flag.NewFlagSet("flags", flag.ExitOnError)
	flagSet.String("config", "", "Path to a JSON formatted config file.")
	flagSet.Parse(args)
	configPath := flagSet.Lookup("config").Value.String()
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
