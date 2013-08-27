package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
	"time"

	"github.com/kvu787/go-schedule/scraper/config"
	"github.com/kvu787/go-schedule/scraper/database"
	_ "github.com/lib/pq"
)

// router is a one-off type that implements the http.Handler interface.
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
		staticFile, err := os.Open("web/" + filePath)
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
		"web/templates/index.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", nil)
}

func deptsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := determineDb()
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()
	queryers, _ := database.Select(db, database.Dept{}, "ORDER BY title")
	var depts []database.Dept
	for _, v := range queryers {
		depts = append(depts, v.(database.Dept))
	}
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"title": strings.Title,
		"upper": strings.ToUpper,
	}).ParseFiles(
		"web/templates/depts.html",
		"web/templates/base.html",
	))

	t.ExecuteTemplate(w, "base", depts)
}

func classesHandler(w http.ResponseWriter, r *http.Request) {
	dept := strings.Split(strings.ToLower(r.URL.Path), "/")[2]
	db, err := determineDb()
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()
	queryers, _ := database.Select(db, database.Class{}, fmt.Sprintf("WHERE dept_abbreviation = '%s' ORDER BY code", dept))
	var classes []database.Class
	for _, v := range queryers {
		classes = append(classes, v.(database.Class))
	}
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"title": strings.Title,
		"upper": strings.ToUpper,
	}).ParseFiles(
		"web/templates/classes.html",
		"web/templates/base.html",
	))

	viewBag := make(map[string]interface{})
	viewBag["classes"] = classes
	viewBag["dept"] = dept

	t.ExecuteTemplate(w, "base", viewBag)
}

func sectsHandler(w http.ResponseWriter, r *http.Request) {
	dept := strings.Split(strings.ToLower(r.URL.Path), "/")[2]
	class := strings.Split(strings.ToLower(r.URL.Path), "/")[3]
	db, err := determineDb()
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer db.Close()
	queryers, _ := database.Select(db, database.Sect{}, fmt.Sprintf("WHERE class_dept_abbreviation = '%s' ORDER BY section", class))
	var sects []database.Sect
	for _, v := range queryers {
		sects = append(sects, v.(database.Sect))
	}
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}).ParseFiles(
		"web/templates/sects.html",
		"web/templates/base.html",
	))

	viewBag := make(map[string]interface{})
	viewBag["dept"] = dept
	viewBag["class"] = class
	viewBag["sects"] = sects

	t.ExecuteTemplate(w, "base", viewBag)
}

func determineDb() (*sql.DB, error) {
	switchDB, err := sql.Open(config.DbConnSwitch.Driver(), config.DbConnSwitch.Conn())
	if err != nil {
		return nil, err
	}
	db, err := database.GetAppDB(switchDB, false)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	if err := os.Chdir(os.ExpandEnv(config.AppRoot)); err != nil {
		fmt.Println(err)
		return
	}
	switch {
	case len(os.Args) < 2:
		fmt.Println("Go Schedule frontend started on port 8080")
		if err := http.ListenAndServe(":8080", router{}); err != nil {
			fmt.Println(err)
			return
		}
	case os.Args[1] == "fcgi":
		fmt.Println("Go Schedule frontend started, listening on port 9000")
		listener, err := net.Listen("tcp", "127.0.0.1:9000")
		if err != nil {
			fmt.Println(err)
			return
		}
		if err := fcgi.Serve(listener, router{}); err != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Println(`
Start the Go Schedule web app.
	usage: web [listen_and_serve]
	arguments
		fcgi    If present, a web server will listen and serve through fcgi on port 9000.
				If not present, the server will listen and serve regularly through port 8080.`)
		return
	}
}
