# Go Schedule

Go Schedule is 
- a library that provides functions for extracting data from UW time schedule pages
- and a web application that uses the library to serve a better time schedule. 

![Screen shot of Go Schedule](https://raw.githubusercontent.com/kvu787/goschedule/master/goschedule.png)

# Usage

## Library

The library is in `lib/`. Install it with `go get github.com/kvu787/goschedule/lib`.

## Web application

The web application is in `goschedule/`. Install it with `go get github.com/kvu787/goschedule/goschedule`.

Run `goschedule help` for more details.

Deployment: 

- Copy the configuration file at `goschedule/config.sample.json` to `config.json` and edit as necessary.
- Setup the databases with  `goschedule setup create --config=<path to config>`.
- Scrape the UW time schedule with `goschedule scrape --config=<path to config>`.
- Run the web application locally with `goschedule web --config=<path to config> --schedule=<name of schedule in config> --local=8080`. 