package main

import (
	"database/sql"
	"fmt"
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
	if err := os.Chdir(os.ExpandEnv(config.AppRoot)); err != nil { // is this idiomatic and safe?
		fmt.Println(err)
		return
	}
	for {
		_main()
	}
}

// This inner main function is used so that the deferred
// functions will be called.
func _main() {
	fmt.Println("Setting up database")
	start := time.Now()
	// open switch db
	switchDB, err := sql.Open(config.DbConnSwitch.Driver(), config.DbConnSwitch.Conn())
	if err != nil {
		log.Fatalln("Failed to open switch db")
		log.Fatalln(err)
		return
	}
	defer switchDB.Close()
	// determine which app db to use
	db, err := database.GetAppDB(switchDB, true)
	if err != nil {
		log.Fatalln("Failed to determine app db")
		log.Fatalln(err)
		return
	}
	defer db.Close()
	// run setup sql against db
	if err = database.SetupDB(db); err != nil {
		log.Fatalln("Failed to setup app db")
		log.Fatalln(err)
		return
	}
	// scrape
	client := http.DefaultClient
	if err = scrapeConcurrentLoadBalance(client, db); err != nil {
		log.Fatalln("Scraping failed")
		log.Fatalln(err)
		return
	}
	// flip switch
	if err = database.FlipSwitch(switchDB); err != nil {
		log.Fatalln("Failed to flip switch db")
		log.Fatalln(err)
		return
	}
	fmt.Println("Scrape done")
	fmt.Println("Time taken:", time.Since(start))
	time.Sleep(config.ScraperTimeout)
}

func scrapeConcurrentLoadBalance(c *http.Client, db *sql.DB) error {
	fmt.Println("Scraper started in concurrent, load balancing mode")
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		log.Fatalln("Failed to fetch RootIndex (department page)")
		return err
	}
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	fetchBuffer := 2
	fetchBufC := make(chan int, fetchBuffer)
	fetchCountC := make(chan int)
	for _, dept := range depts {
		go func(dept database.Dept) {
			fetchBufC <- 1
			scrapeDept(dept, c, db)
			fetchCountC <- 1
			<-fetchBufC
		}(dept)
	}
	for i := 0; i < len(depts); i++ {
		<-fetchCountC
	}
	return nil
}

func scrapeConcurrent(c *http.Client, db *sql.DB, concurrent bool) error {
	fmt.Println("Scraper started in concurrent mode")
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		log.Fatalln("Failed to fetch RootIndex (department page)")
		return err
	}
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	quitc := make(chan int)
	for _, dept := range depts {
		go func(dept database.Dept) {
			scrapeDept(dept, c, db)
			quitc <- 1
		}(dept)
		time.Sleep(config.ScraperFetchTimeout)
	}
	for i := 0; i < len(depts); i++ {
		<-quitc
	}
	return nil
}

func scrape(c *http.Client, db *sql.DB, concurrent bool) error {
	fmt.Println("Scraper started in non-concurrent mode")
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		log.Fatalln("Failed to fetch RootIndex (department page)")
		return err
	}
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	for _, dept := range depts {
		scrapeDept(dept, c, db)
	}
	return nil
}

func scrapeDept(dept database.Dept, c *http.Client, db *sql.DB) {
	classSectIndex, err := fetch.Get(c, dept.Link)
	if err != nil {
		return // skip if dept link is bad
	}
	database.Insert(db, dept)
	classes := extract.ClassIndex(classSectIndex).Extract(dept)
	for _, class := range classes {
		database.Insert(db, class)
	}
	sects := extract.SectIndex(classSectIndex).Extract(classes)
	for _, sect := range sects {
		database.Insert(db, sect)
	}
}
