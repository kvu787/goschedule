package main

import (
	"flag"
	"fmt"
	"os"
)

var usage string = `

go-schedule is a tool for running the Go-Schedule application.

Usage:

    go-schedule command [flags]

Commands:

    setup    Setup the databases used by scrape to store records.
    scrape   Scrape the UW time schedule.
    web      Start the web application to view Go-Schedule.
    help     Use "go-schedule help [command] for more information about a command.

Built by Kevin Vu
`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}
	var flags []string
	if len(os.Args) < 3 {
		flags = []string{}
	} else {
		flags = os.Args[2:]
	}
	switch os.Args[1] {
	case "setup":
		handleSetup(flags)
	case "scrape":
		handleScrape(flags)
	case "web":
		handleWeb(flags)
	case "help":
	default:
		fmt.Println(usage)
		os.Exit(1)
	}

}

var flagSet = flag.NewFlagSet("", flag.PanicOnError)

func handleSetup(flags []string) {
	var user string
	var password string
	flagSet.StringVar(&user, "", "", "")
	flagSet.StringVar(&password, "", "", "")
	flagSet.Parse(flags)
	// setupDb(user, password)
}

func handleScrape(flags []string) {

}

func handleWeb(flags []string) {
	fcgi := flagSet.Bool("fcgi", false, "")
	flagSet.Parse(flags)
	fmt.Println(*fcgi)
}
