package web

import (
	"github.com/kvu787/go_schedule/backend/extract"
	"html/template"
	"net/http"
	// "net/url"
	"strings"

	"appengine"
	"appengine/datastore"
)

func init() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/depts/", handleDepts)
	http.HandleFunc("/dept/", handleClasses)
	http.HandleFunc("/class/", handleSects)
}

// NOTE: nested template pattern from http://stackoverflow.com/a/11468132/1559886

func handleIndex(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(
		"web/templates/index.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", nil)
}

func handleDepts(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	depts := []*extract.Dept{}
	datastore.NewQuery("Dept").
		GetAll(c, &depts)

	t := template.Must(template.ParseFiles(
		"web/templates/depts.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", depts)
}

func handleClasses(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// grab dept abbreviation
	path := strings.Split(r.URL.String(), "/")
	deptAbbr := path[len(path)-1]

	// set dept key
	deptKey := datastore.NewKey(c, "Dept", deptAbbr, 0, nil)

	// grab classes with dept key
	classes := []*extract.Class{}
	datastore.NewQuery("Class").
		Ancestor(deptKey).
		GetAll(c, &classes)

	t := template.Must(template.ParseFiles(
		"web/templates/classes.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", classes)
}

func handleSects(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	// grab dept key, class key
	path := strings.Split(r.URL.String(), "/")
	deptStringID := path[len(path)-2]
	classStringID := path[len(path)-1]

	// set class key
	deptKey := datastore.NewKey(c, "Dept", deptStringID, 0, nil)
	classKey := datastore.NewKey(c, "Class", classStringID, 0, deptKey)

	// grab sects with class key
	sects := []*extract.Sect{}
	datastore.NewQuery("Sect").
		Ancestor(classKey).
		GetAll(c, &sects)

	t := template.Must(template.ParseFiles(
		"web/templates/sects.html",
		"web/templates/base.html",
	))
	t.ExecuteTemplate(w, "base", sects)
}
