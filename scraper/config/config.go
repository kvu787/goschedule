package config

import "time"

const (
	ScraperTimeout      time.Duration = 5 * time.Minute                                         // how long to wait between each full scrape
	ScraperFetchTimeout time.Duration = 200 * time.Millisecond                                  // delay between launching each fetch during scraping
	RootIndex           string        = "https://www.washington.edu/students/timeschd/AUT2013/" // link to departments index
	AppRoot             string        = "$GOPATH/src/github.com/kvu787/go-schedule"             // path to application root
	SchemaPath          string        = "scraper/utility/sql/schema.sql"                        // relative to AppRoot
	SwitchDBTable       string        = "switch_table"
	SwitchDBCol         string        = "switch_col"
)
