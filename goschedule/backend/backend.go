package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/kvu787/goschedule/lib"
)

// Scrape will begin a full time schedule scrape and store results in a database.
// Parameter link must be a the time schedule page listing departments and colleges.
func Scrape(link string, db *sql.DB) {
	if err := db.Ping(); err != nil {
		panic("Bad db connection")
	}
	body, err := get(link)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch time schedule root at %s: %s", link, err))
	}
	colleges, err := goschedule.ExtractColleges(body)
	if err != nil {
		log.Println(err)
	}
	uniqueDepts := make(map[string]int)
	fmt.Println("starting scrape")
	// scrape colleges
	for _, college := range colleges {
		if err := goschedule.Insert(db, college); err != nil {
			log.Println(err)
		}
		// scrape department for each college
		depts, err := goschedule.ExtractDepts(body[college.Start:college.End], college.Abbreviation, link, &uniqueDepts)
		if err != nil {
			log.Println(err)
		}
		for _, dept := range depts {
			classIndex, err := get(dept.Link)
			if err != nil {
				log.Printf("department %q SKIPPED: %v\n", dept.Name, err)
				continue
			}
			if err := dept.ScrapeAbbreviation(classIndex); err != nil {
				log.Printf("department %q SKIPPED: %v\n", dept.Name, err)
				continue
			}
			if err := goschedule.Insert(db, dept); err != nil {
				fmt.Println("LINK:", dept.Link)
				log.Println(err)
			}
			// scrape classes for each department
			classes := goschedule.ExtractClasses(classIndex, dept.Abbreviation)
			var sections []goschedule.Sect
			// scrape sections for each class
			for _, class := range classes {
				if err := goschedule.Insert(db, class); err != nil {
					log.Println(err)
				}
				sects, err := goschedule.ExtractSects(classIndex[class.Start:class.End], class.AbbreviationCode)
				if err != nil {
					log.Println(err)
				}
				sections = append(sections, sects...)
			}
			for _, sect := range sections {
				if err := goschedule.Insert(db, sect); err != nil {
					log.Println(err)
				}
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

// parseConfig reads a JSON format byte slice into a map.
func parseConfig(config []byte) (result map[string]interface{}) {
	json.Unmarshal(config, &result)
	return
}

// get requests a link with the given client and returns the bytes
// of the response body if successful.
// A response with a non-2XX/3XX status code is considered an error.
func get(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Print(err)
		if urlError, ok := err.(*url.Error); ok && urlError.Err.Error() == "EOF" {
			return "", getEofError(fmt.Sprintf("get EOF error: %+v, Link: %q", err, link))
		}
		return "", fmt.Errorf("get error: %+v, Link: %q", err, link)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 || resp.StatusCode < 200 {
		return "", fmt.Errorf("get: returned with non-2XX/3XX status code: %d, link: %q", resp.StatusCode, link)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("get: error in reading response body: %v", err)
	}
	return string(body), nil
}

type getEofError string

func (err getEofError) Error() string {
	return string(err)
}
