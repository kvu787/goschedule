// Package config provides various configurable variables used in
// both the web and scraper programs.
package config

import "time"

const (
	// how long to wait between each full scrape
	ScraperTimeout time.Duration = 5 * time.Minute
	// delay between launching each fetch during scraping
	ScraperFetchTimeout time.Duration = 200 * time.Millisecond
	// link to departments index
	RootIndex string = "https://www.washington.edu/students/timeschd/AUT2013/"
	// path to application root
	AppRoot string = "$GOPATH/src/github.com/kvu787/go-schedule"
	// relative to AppRoot
	SchemaPath string = "scraper/utility/sql/schema.sql"
	// name of the table in the 'switch' db (should only be one)
	SwitchDBTable string = "switch_table"
	// name of the column in the 'switch' db (should only be one)
	SwitchDBCol string = "switch_col"
)
