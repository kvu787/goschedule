package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kvu787/go_schedule/crawler/config"
	"github.com/kvu787/go_schedule/crawler/database"
	_ "github.com/lib/pq"
)

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case route("/").Match(r.URL):
		indexHandler(w, r)
		return
	// serve files in the assets directory properly
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
	case r.URL.String() == "/schedule":
		deptsHandler(w, r)
		return
	case route("/schedule/:dept").Match(r.URL):
		classesHandler(w, r)
		return
	case route("/schedule/:dept/:class").Match(r.URL):
		sectsHandler(w, r)
		return
	default:
		fmt.Fprintf(w, "no route matched")
		fmt.Fprintln(w, r.URL.String())
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
	// fmt.Fprintln(w, "depts handler called")
	// fmt.Fprintln(w, r.URL.String())

	db, _ := sql.Open(config.Db, config.DbConn)
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
	dept := strings.Split(r.URL.String(), "/")[2]

	db, _ := sql.Open(config.Db, config.DbConn)
	defer db.Close()
	queryers, _ := database.Select(db, database.Class{}, fmt.Sprintf("WHERE dept_abbreviation = '%s'", dept))
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
	dept := strings.Split(r.URL.String(), "/")[2]
	class := strings.Split(r.URL.String(), "/")[3]
	// fmt.Fprintln(w, "sects handler called")
	// fmt.Fprintln(w, r.URL.String())
	// fmt.Fprintln(w, "dept:", dept)
	// fmt.Fprintln(w, "class:", class)

	db, _ := sql.Open(config.Db, config.DbConn)
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

func main() {
	http.HandleFunc(`/`, router)
	fmt.Println("Go Schedule frontend started")
	http.ListenAndServe(":8080", nil)
}
