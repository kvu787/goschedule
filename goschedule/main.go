package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kvu787/goschedule"
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
			fmt.Println(usage)
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

var flagSet = flag.NewFlagSet("flags", flag.ExitOnError)

func init() {
	flagSet.String("config", "", "Path to a JSON formatted config file.")
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
	var config map[string]interface{}
	parseConfig(args[1:], &config)
	// setup superuser db connection
	dbLogin := config["dbLogin"].(map[string]interface{})
	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=require",
		dbLogin["user"].(string),
		dbLogin["dbname"].(string),
		dbLogin["password"].(string),
	))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// setup databases for each schedule
	var dbSets []map[string]interface{}
	for _, v := range config["schedules"].([]interface{}) {
		dbSets = append(dbSets, v.(map[string]interface{}))
	}
	for _, dbSet := range dbSets {
		_, err := db.Exec(fmt.Sprintf("%s DATABASE goschedule_%s_switch", command, dbSet["name"].(string)))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = db.Exec(
			fmt.Sprintf("%s DATABASE goschedule_%s_app1", command, dbSet["name"].(string)))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = db.Exec(
			fmt.Sprintf("%s DATABASE goschedule_%s_app2", command, dbSet["name"].(string)))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// load db schemas if in create mode
		if command == "CREATE" {
			// load switch schema
			dbSwitch, err := sql.Open("postgres", fmt.Sprintf(
				"user=%s dbname=%s password=%s sslmode=require",
				dbLogin["user"].(string),
				fmt.Sprintf("goschedule_%s_switch", dbSet["name"].(string)),
				dbLogin["password"].(string),
			))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if _, err := dbSwitch.Exec("CREATE TABLE switch_table ( switch_col int)"); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if _, err := dbSwitch.Exec("INSERT INTO switch_table VALUES (1)"); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// load app db schemas
			for i := 1; i < 3; i++ {
				dbApp, err := sql.Open("postgres", fmt.Sprintf(
					"user=%s dbname=%s password=%s sslmode=require",
					dbLogin["user"].(string),
					fmt.Sprintf("goschedule_%s_app%d", dbSet["name"].(string), i),
					dbLogin["password"].(string),
				))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				objects := []interface{}{goschedule.College{}, goschedule.Dept{}, goschedule.Class{}, goschedule.Sect{}}
				for _, object := range objects {
					if _, err := dbApp.Exec(goschedule.GenerateSchema(object)); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		}
	}
}

func handleScrape(flags []string) {

}

func handleWeb(flags []string) {
	fcgi := flagSet.Bool("fcgi", false, "")
	flagSet.Parse(flags)
	fmt.Println(*fcgi)
}

type jsonConfig struct {
}

func parseConfig(flags []string, target interface{}) {
	flagSet.Parse(flags)
	configPath := flagSet.Lookup("config").Value.String()
	if configPath == "" {
		fmt.Println("missing `--config` flag")
		os.Exit(1)
	}
	config, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("error reading config at %s\n", configPath)
		os.Exit(1)
	}
	json.Unmarshal(config, target)
}
