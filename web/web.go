package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
	"time"

	"github.com/kvu787/go_schedule/crawler/config"
	"github.com/kvu787/go_schedule/crawler/database"
	_ "github.com/lib/pq"
)

type router struct{}

func (ro router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case route("/").Match(r.URL):
		indexHandler(w, r)
		return
	// properly serve files in the assets directory
	case route("/assets/:type/:file").Match(r.URL):
		var filePath string = string(r.URL.Path[1:])
		filePathSlice := strings.Split(filePath, "/")
		fileName := filePathSlice[len(filePathSlice)-1]
		staticFile, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(w, "404, file not found error: "+err.Error())
		} else {
			http.ServeContent(w, r, fileName, time.Now(), staticFile)
		}
		return
	case route("/schedule").Match(r.URL):
		deptsHandler(w, r)
		return
	case route("/schedule/:dept").Match(r.URL):
		classesHandler(w, r)
		return
	case route("/schedule/:dept/:class").Match(r.URL):
		sectsHandler(w, r)
		return
	default:
		fmt.Fprintf(w, "No route matched for:\n%s", r.URL.String())
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/base.html",
	))
	t.ExecuteTemplate(w, "base", nil)
}

func deptsHandler(w http.ResponseWriter, r *http.Request) {
	db := determineDb()
	defer db.Close()
	queryers, _ := database.Select(db, database.Dept{}, "")
	var depts []database.Dept
	for _, v := range queryers {
		depts = append(depts, v.(database.Dept))
	}
	t := template.Must(template.ParseFiles(
		"templates/depts.html",
		"templates/base.html",
	))
	t.ExecuteTemplate(w, "base", depts)
}

func classesHandler(w http.ResponseWriter, r *http.Request) {
	dept := strings.Split(r.URL.Path, "/")[2]
	db := determineDb()
	defer db.Close()
	queryers, _ := database.Select(db, database.Class{}, fmt.Sprintf("WHERE dept_abbreviation = '%s' ORDER BY code", dept))
	var classes []database.Class
	for _, v := range queryers {
		classes = append(classes, v.(database.Class))
	}
	t := template.Must(template.ParseFiles(
		"templates/classes.html",
		"templates/base.html",
	))

	viewBag := make(map[string]interface{})
	viewBag["classes"] = classes
	viewBag["dept"] = dept

	t.ExecuteTemplate(w, "base", viewBag)
}

func sectsHandler(w http.ResponseWriter, r *http.Request) {
	dept := strings.Split(r.URL.Path, "/")[2]
	class := strings.Split(r.URL.Path, "/")[3]

	db := determineDb()
	defer db.Close()
	queryers, _ := database.Select(db, database.Sect{}, fmt.Sprintf("WHERE class_dept_abbreviation = '%s' ORDER BY section", class))
	var sects []database.Sect
	for _, v := range queryers {
		sects = append(sects, v.(database.Sect))
	}
	t := template.Must(template.ParseFiles(
		"templates/sects.html",
		"templates/base.html",
	))

	viewBag := make(map[string]interface{})
	viewBag["dept"] = dept
	viewBag["class"] = class
	viewBag["sects"] = sects

	t.ExecuteTemplate(w, "base", viewBag)
}

func determineDb() *sql.DB {
	i, _ := database.GetSwitch()
	if i == 1 {
		db, _ := sql.Open(config.DbConn1.Driver(), config.DbConn1.Conn())
		return db
	} else {
		db, _ := sql.Open(config.DbConn2.Driver(), config.DbConn2.Conn())
		return db
	}
}

func main() {
	fmt.Println("Go Schedule frontend started")
	listener, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		fmt.Println(err)
	}
	if err := fcgi.Serve(listener, router{}); err != nil {
		fmt.Println(err)
	}
}
