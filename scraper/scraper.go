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
	// switch to application directory
	if err := os.Chdir(os.ExpandEnv(config.AppRoot)); err != nil { // is this idiomatic and safe?
		fmt.Println(err)
		return
	}
	// run loop calls to _main with a timeout in between each call
	for {
		_main()
		if !config.LoopScraper {
			log.Println("LoopScraper is off, exiting")
			break
		}
		time.Sleep(config.ScraperTimeout)
	}
}

// _main runs one full scrape process.
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
		log.Println(err)
		log.Fatalln("Failed to setup app db")
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
	// done!
	fmt.Println("Scrape done")
	fmt.Println("Time taken:", time.Since(start))
}

// scrapeConcurrentLoadBalance limits how many fetch
// and db insert operations can run concurrently.
func scrapeConcurrentLoadBalance(c *http.Client, db *sql.DB) error {
	fmt.Println("Scraper started in concurrent, load balancing mode")
	// fetch department index
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		log.Fatalln("Failed to fetch RootIndex (department page)")
		return err
	}
	// extract dept structs from dept index
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	// setup channels
	fetchBufc := make(chan int, config.FetchBuffer)
	insertBufc := make(chan int, config.InsertBuffer)
	donec := make(chan int)
	// run 'scrape dept page' operations concurrently
	for _, dept := range depts {
		go func(dept database.Dept) {
			scrapeDeptLoadBalance(dept, c, db, fetchBufc, insertBufc)
			donec <- 1
		}(dept)
	}
	// return when all scrape dept operations finish
	for i := 0; i < len(depts); i++ {
		<-donec
	}
	return nil
}

// scrapeConcurrent runs the scrapeDept operations
// concurrently, staggering calls with a timeout.
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

// scrape fetches pages and inserts into the database
// sequentially.
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

// scrapeDeptLoadBalance fetches a class/section page
// and runs class/section inserts concurrently and load
// balanced.
// Each insert operation runs when there is space in the
// insertc chan.
func scrapeDeptLoadBalance(dept database.Dept, c *http.Client, db *sql.DB, fetchc chan int, insertc chan int) {
	// fetch class/section page if buffer is ready
	fetchc <- 1
	classSectIndex, err := fetch.Get(c, dept.Link)
	if err != nil {
		<-fetchc
		return // skip if dept link is bad
	}
	<-fetchc
	// chan to track number of INSERTs issued
	localInsertc := make(chan int)
	// queue up inserts
	go func(dept database.Dept) {
		insertc <- 1
		database.Insert(db, dept)
		<-insertc
		localInsertc <- 1
	}(dept)
	classes := extract.ClassIndex(classSectIndex).Extract(dept)
	for _, class := range classes {
		go func(class database.Class) {
			insertc <- 1
			database.Insert(db, class)
			<-insertc
			localInsertc <- 1
		}(class)
	}
	sects := extract.SectIndex(classSectIndex).Extract(classes)
	for _, sect := range sects {
		go func(sect database.Sect) {
			insertc <- 1
			database.Insert(db, sect)
			<-insertc
			localInsertc <- 1
		}(sect)
	}
	// count inserts issued in this function
	localInserts := 1 + len(classes) + len(sects)
	// wait to get signals from all inserts
	for i := 0; i < localInserts; i++ {
		<-localInsertc
	}
	// all inserts complete
	return
}

// scrapeDept sequentially fetches a class/section page
// and inserts class and section information into the db.
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
