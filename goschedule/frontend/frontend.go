package frontend

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

	"github.com/kvu787/goschedule/lib"
)

var appDb *sql.DB

func Serve(db *sql.DB, local bool, frontendRoot string, port int) error {
	appDb = db
	if err := os.Chdir(os.ExpandEnv(frontendRoot)); err != nil {
		fmt.Println(err)
		return err
	}
	if local {
		fmt.Println("Go Schedule frontend started locally on port 8080")
		http.HandleFunc("/", routing)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		fmt.Println("Go Schedule frontend started through fcgi on port 9000")
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return err
		}
		http.HandleFunc("/", routing)
		if err := fcgi.Serve(listener, nil); err != nil {
			return err
		}
	}
	return nil
}

func routing(w http.ResponseWriter, r *http.Request) {
	switch {
	case route("/").match(r.URL.Path):
		indexHandler(w, r)
		return
	// properly serve files in the assets directory
	case route("/assets/:type/:file").match(r.URL.Path):
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
	case route("/schedule").match(r.URL.Path):
		deptsHandler(w, r)
		return
	case route("/schedule/:dept").match(r.URL.Path):
		classesHandler(w, r)
		return
	case route("/schedule/:dept/:class").match(r.URL.Path):
		sectsHandler(w, r)
		return
	default:
		fmt.Fprintf(w, "No route matched for:\n%q", r.URL.Path)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(
		"web/templates/index.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", nil)
}

// CREDIT STACKOVERFLOW FOR TEMPLATING PATTERN
func deptsHandler(w http.ResponseWriter, r *http.Request) {
	deptRecords, err := goschedule.Select(appDb, goschedule.Dept{}, "")
	if err != nil {
		panic(err)
	}
	var depts []goschedule.Dept
	for _, v := range deptRecords {
		depts = append(depts, v.(goschedule.Dept))
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
	classRecords, err := goschedule.Select(appDb, goschedule.Class{}, fmt.Sprintf("WHERE deptkey = '%s' ORDER BY code", dept))
	if err != nil {
		panic(err)
	}
	var classes []goschedule.Class
	for _, v := range classRecords {
		classes = append(classes, v.(goschedule.Class))
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
	sectRecords, err := goschedule.Select(appDb, goschedule.Sect{}, fmt.Sprintf("WHERE classkey = '%s' ORDER BY section", class))
	if err != nil {
		panic(err)
	}
	var sects []goschedule.Sect
	for _, v := range sectRecords {
		sects = append(sects, v.(goschedule.Sect))
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
