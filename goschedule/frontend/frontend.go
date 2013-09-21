package frontend

import (
	"database/sql"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"sort"
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
		http.HandleFunc("/", router)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return err
		}
		http.HandleFunc("/", router)
		if err := fcgi.Serve(listener, nil); err != nil {
			return err
		}
	}
	return nil
}

var routing = [][]interface{}{
	{"/", indexHandler},
	{"/test_ajax", ajaxHandler},
	{"/schedule", deptsHandler},
	{"/schedule/:dept", classesHandler},
	{"/schedule/:dept/:class", sectsHandler},
	{"/assets/:type/:file", assetHandler},
}

type routeHandler func(http.ResponseWriter, *http.Request, map[string]string)

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var matched bool
	for _, tuple := range routing {
		handler := tuple[1].(func(http.ResponseWriter, *http.Request, map[string]string))
		if ro := route(tuple[0].(string)); ro.match(path) {
			handler(w, r, ro.parse(path))
			matched = true
		}
	}
	if !matched {
		fmt.Fprintf(w, "No route matched for:\n%q", r.URL.Path)
	}
}

func ajaxHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	fmt.Fprintf(w, `alert('works!');`)
}

// CREDIT: http://stackoverflow.com/questions/11467731/is-it-possible-to-have-nested-templates-in-go-using-the-standard-library-googl
func indexHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	t := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/base.html",
	))
	t.ExecuteTemplate(w, "base", nil)
}

func deptsHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var data = make(map[string][]goschedule.Dept)
	// get colleges
	collegeRecords, err := goschedule.Select(appDb, goschedule.College{}, "ORDER BY abbreviation")
	if err != nil {
		panic(err)
	}
	var collegeNames []string
	var collegesNamesToAbbreviations = make(map[string]string)
	for _, v := range collegeRecords {
		college := v.(goschedule.College)
		// create list of college names
		collegeNames = append(collegeNames, college.Name)
		// create map of college names to abbreviations
		collegesNamesToAbbreviations[college.Name] = college.Abbreviation
	}
	for _, collegeName := range collegeNames {
		// get depts
		deptRecords, err := goschedule.Select(appDb, goschedule.Dept{}, fmt.Sprintf("WHERE collegekey = '%s'", collegesNamesToAbbreviations[collegeName]))
		if err != nil {
			panic(err)
		}
		// create map of college names to depts
		for _, v := range deptRecords {
			data[collegeName] = append(data[collegeName], v.(goschedule.Dept))
		}
	}
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"title": strings.Title,
		"upper": strings.ToUpper,
	}).ParseFiles(
		"templates/depts.html",
		"templates/base.html",
	))
	// sort slice of college names
	sort.Strings(collegeNames)
	viewBag := map[string]interface{}{
		"collegeNames":         collegeNames,
		"collegeAbbreviations": collegesNamesToAbbreviations,
		"collegesMap":          data,
	}
	t.ExecuteTemplate(w, "base", viewBag)
}

func classesHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	classRecords, err := goschedule.Select(appDb, goschedule.Class{}, fmt.Sprintf("WHERE deptkey = '%s' ORDER BY code", params["dept"]))
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
		"templates/classes.html",
		"templates/base.html",
	))
	viewBag := map[string]interface{}{
		"classes": classes,
		"dept":    params["dept"],
	}
	t.ExecuteTemplate(w, "base", viewBag)
}

func sectsHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
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
		"templates/sects.html",
		"templates/base.html",
	))
	viewBag := make(map[string]interface{})
	viewBag["dept"] = dept
	viewBag["class"] = class
	viewBag["sects"] = sects
	t.ExecuteTemplate(w, "base", viewBag)
}

func assetHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	filePath := fmt.Sprintf("assets/%s/%s", params["type"], params["file"])
	staticFile, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(w, "404, file not found error: %v", err.Error())
	} else {
		http.ServeContent(w, r, params["file"], time.Now(), staticFile)
	}
}
