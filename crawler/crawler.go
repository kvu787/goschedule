// Main class for the Go Schedule web app.
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kvu787/go_schedule/crawler/config"
	"github.com/kvu787/go_schedule/crawler/database"
	"github.com/kvu787/go_schedule/crawler/extract"
	"github.com/kvu787/go_schedule/crawler/fetch"
	_ "github.com/lib/pq"
)

var offline bool = false

func main() {
	if !offline {
		client := http.DefaultClient
		db, err := sql.Open(config.Db, config.DbConn)
		defer db.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		Crawl(client, db)
	} else {
		db, err := sql.Open(config.TestDb, config.TestDbConn)
		defer db.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		offlineCrawl(db)
	}
}

func Crawl(c *http.Client, db *sql.DB) {
	deptIndex, err := fetch.Get(c, config.RootIndex)
	if err != nil {
		fmt.Println(err)
		return
	}
	depts := extract.DeptIndex(deptIndex).Extract(nil)
	for _, dept := range depts {
		classSectIndex, err := fetch.Get(c, dept.Link)
		if err != nil {
			fmt.Println(err)
			continue // skip if dept link is bad
		}
		if database.Insert(db, dept); err != nil {
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

func offlineCrawl(db *sql.DB) {
	fmt.Println("Crawler started")
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
