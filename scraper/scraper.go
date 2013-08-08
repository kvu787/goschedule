package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kvu787/go-schedule/scraper/config"
	"github.com/kvu787/go-schedule/scraper/database"
	"github.com/kvu787/go-schedule/scraper/extract"
	"github.com/kvu787/go-schedule/scraper/fetch"
	_ "github.com/lib/pq"
)

func main() {
	if err := os.Chdir(os.ExpandEnv(config.AppRoot)); err != nil {
		fmt.Println(err)
		return
	}
	dbSwitch, err := database.GetSwitch()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		start := time.Now()
		log.Println("Scraper started at:", start.String())
		client := http.DefaultClient
		var err error
		var db *sql.DB
		if dbSwitch == 2 {
			fmt.Println("Setting up database on DbConn1...")
			db, err = sql.Open(config.DbConn1.Driver(), config.DbConn1.Conn())

		} else {
			fmt.Println("Setting up database on DbConn2...")
			db, err = sql.Open(config.DbConn2.Driver(), config.DbConn2.Conn())
		}
		if err != nil {
			log.Fatalln("Error in opening database")
			fmt.Println(err)
			return
		}
		defer db.Close()
		if err = setupDb(db); err != nil {
			log.Fatalln("Error in setting up database, stopping scraper")
			fmt.Println(err)
			return
		}
		Scrape(client, db, true)
		fmt.Println("Time taken:", time.Since(start))
		if dbSwitch == 2 {
			database.UpdateSwitch(2, 1)
		} else {
			database.UpdateSwitch(1, 2)
		}
		dbSwitch, _ = database.GetSwitch()
		time.Sleep(config.ScraperTimeout)
	}
}

func setupDb(db *sql.DB) error {
	statements, err := database.ParseSqlFile(config.SchemaPath)
	if err != nil {
		return err
	}
	for _, s := range statements {
		_, err := db.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func Scrape(c *http.Client, db *sql.DB, concurrent bool) error {
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		return err
	}
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	if concurrent {
		fmt.Println("Scraper started in concurrent mode")
		quitc := make(chan int)
		for _, dept := range depts {
			go func(dept database.Dept) {
				fmt.Println("Scraping:", dept.Title)
				scrapeDept(dept, c, db)
				quitc <- 1
			}(dept)
			time.Sleep(config.ScraperFetchTimeout)
		}
		for i := 0; i < len(depts); i++ {
			<-quitc
		}
	} else {
		fmt.Println("Scraper started in non-concurrent mode")
		for _, dept := range depts {
			fmt.Println("Scraping:", dept.Title)
			fetch.Get(c, dept.Link)
			scrapeDept(dept, c, db)
		}
	}
	fmt.Println("Scraper done")
	return nil
}

func scrapeDept(dept database.Dept, c *http.Client, db *sql.DB) {
	classSectIndex, err := fetch.Get(c, dept.Link)
	if err != nil {
		fmt.Println(err)
		return // skip if dept link is bad
	}
	if database.Insert(db, dept); err != nil {
		fmt.Println("Error inserting dept")
	}
	classes := extract.ClassIndex(classSectIndex).Extract(dept)
	for _, class := range classes {
		if err := database.Insert(db, class); err != nil {
			fmt.Println("Error inserting class")
		}
	}
	sects := extract.SectIndex(classSectIndex).Extract(classes)
	for _, sect := range sects {
		if err := database.Insert(db, sect); err != nil {
			fmt.Println("Error inserting sect")
		}
	}
}

// This is probably currently broken
func offlineScrape(db *sql.DB) {
	fmt.Println("Scraper started")
	math, _ := ioutil.ReadFile("utility/sample/math.html")
	engl, _ := ioutil.ReadFile("utility/sample/engl.html")
	cse, _ := ioutil.ReadFile("utility/sample/cse.html")
	classSectIndices := [][]byte{
		math,
		engl,
		cse,
	}
	for i, classSectIndex := range classSectIndices {
		dept := database.Dept{
			fmt.Sprintf("Department Number %d", i+1),
			fmt.Sprintf("DN%d", i+1),
			fmt.Sprintf("http://uw.edu/dept%d", i+1),
		}
		if err := database.Insert(db, dept); err != nil {
			fmt.Println("Error inserting dept")
			fmt.Println(dept.Abbreviation)
		}
		classes := extract.ClassIndex(classSectIndex).Extract(dept)
		for _, class := range classes {
			if err := database.Insert(db, class); err != nil {
				fmt.Println("Error inserting class")
				fmt.Println("Dept", dept.Abbreviation)
				fmt.Println("Class key", class.AbbreviationCode)
				fmt.Println(err)
			}
		}
		sects := extract.SectIndex(classSectIndex).Extract(classes)
		for _, sect := range sects {
			if err := database.Insert(db, sect); err != nil {
				fmt.Println("Error inserting sect")
				fmt.Println(dept.Abbreviation)
				fmt.Println(sect.SLN)
				fmt.Println(err)
			}
		}
	}
}
