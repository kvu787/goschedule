# Go Schedule

Go Schedule is 
- a library that provides functions for extracting data from UW time schedule pages
- and a web application that uses the library to serve a better time schedule. 

# Usage

## Library

The library is in `lib/`. Install it with `go get github.com/kvu787/goschedule/lib`.

## Web application

The web application is in `goschedule/`. Install it with `go get github.com/kvu787/goschedule/goschedule`.

Run `goschedule help` for more details.

Deployment: 

- Edit the configuration file at `goschedule/config.json` as necessary.
- Setup the databases with  `goschedule setup create --config=<path to config>`.
- Scrape the UW time schedule with `goschedule scrape --config=<path to config>`.
- Run the web application locally with `goschedule web --config=<path to config> --schedule=<name of schedule in config> --local=8080`. 