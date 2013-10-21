package frontend

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kvu787/goschedule/goschedule/shared"
	"github.com/kvu787/goschedule/lib"
)

var appDb *sql.DB
var switchDatabase *sql.DB
var conn string

func Serve(connString string, switchDb *sql.DB, local bool, frontendRoot string, port int) error {
	conn = connString
	switchDatabase = switchDb
	if err := os.Chdir(os.ExpandEnv(frontendRoot)); err != nil {
		return err
	}
	if local {
		http.HandleFunc("/", router)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
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
	{"/search", searchHandler},
	{"/schedule", deptsHandler},
	{"/schedule/:dept", classesHandler},
	{"/schedule/:dept/:class", sectsHandler},
	{"/assets/:type/:file", assetHandler},
}

type routeHandler func(http.ResponseWriter, *http.Request, map[string]string)

func router(w http.ResponseWriter, r *http.Request) {
	// determine application db
	appNum, err := shared.GetSwitch(switchDatabase)
	if err != nil {
		panic(fmt.Sprintf("Failed to query switch database for app db number in frontend.router: %v", err))
	}
	if appNum == 1 {
		appNum = 2
	} else {
		appNum = 1
	}
	appDb, err = sql.Open("postgres", fmt.Sprintf(conn, appNum))
	if err != nil {
		panic(err)
	}
	defer appDb.Close()
	// process request
	path := strings.ToLower(r.URL.Path)
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

func searchHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	search := strings.TrimSpace(r.FormValue("search"))
	colleges, err := searchColleges(search)
	if err != nil {
		panic(err)
	}
	depts, err := searchDepts(search)
	if err != nil {
		panic(err)
	}
	classes, err := searchClasses(search)
	if err != nil {
		panic(err)
	}

	var htmlBuffer = &bytes.Buffer{}
	viewBag := map[string]interface{}{
		"colleges": colleges,
		"depts":    depts,
		"classes":  classes,
		"query":    search,
	}
	searchTemplate, err := ioutil.ReadFile("templates/search.html")
	if err != nil {
		panic(err)
	}
	if err := template.Must(template.New("").Funcs(template.FuncMap{
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"boldWords": boldWords,
		"toHTML":    toHTML,
	}).Parse(string(searchTemplate))).Execute(htmlBuffer, viewBag); err != nil {
		panic(err)
	}
	htmlStr := htmlBuffer.String()
	htmlStr = strings.Replace(htmlStr, "\n", "", -1)

	t := template.Must(template.ParseFiles("templates/search.js"))
	// sort slice of college names
	t.ExecuteTemplate(w, "searchjs", template.HTML(htmlStr))
}

func toHTML(in string) template.HTML {
	return template.HTML(in)
}

func boldWords(search, in string) string {
	inSlice := strings.Split(strings.TrimSpace(in), " ")
	searchSlice := strings.Split(strings.TrimSpace(search), " ")
	for i := range inSlice {
		var outWord string
		var longestSearchTerm string
		for _, searchTerm := range searchSlice {
			if checkedWord := boldPrefix(inSlice[i], searchTerm); len(searchTerm) > len(longestSearchTerm) && len(checkedWord) > len(outWord) {
				longestSearchTerm = searchTerm
				outWord = checkedWord
			}
		}
		if len(outWord) > 0 {
			inSlice[i] = outWord
		}
	}
	return strings.Join(inSlice, " ")
}

func boldPrefix(word, searchTerm string) string {
	if strings.HasPrefix(strings.ToLower(word), strings.ToLower(searchTerm)) {
		word = word[:len(searchTerm)] + "</strong>" + word[len(searchTerm):]
		word = "<strong>" + word
		return word
	}
	return ""
}

func searchColleges(search string) ([]goschedule.College, error) {
	records, err := goschedule.Select(appDb, goschedule.College{},
		fmt.Sprintf("ORDER BY word_score('%s', name) + word_score('%s', abbreviation) DESC, letter_score('%s', name) + letter_score('%s', abbreviation) DESC LIMIT 5", search, search, search, search))
	if err != nil {
		return nil, err
	}
	var colleges []goschedule.College
	for _, record := range records {
		colleges = append(colleges, record.(goschedule.College))
	}
	return colleges, nil
}

func searchDepts(search string) ([]goschedule.Dept, error) {
	records, err := goschedule.Select(appDb, goschedule.Dept{},
		fmt.Sprintf("ORDER BY word_score('%s', name) + word_score('%s', abbreviation) DESC, letter_score('%s', name) + letter_score('%s', abbreviation) DESC LIMIT 5", search, search, search, search))
	if err != nil {
		return nil, err
	}
	var depts []goschedule.Dept
	for _, record := range records {
		depts = append(depts, record.(goschedule.Dept))
	}
	return depts, nil
}

func searchClasses(search string) ([]goschedule.Class, error) {
	records, err := goschedule.Select(appDb, goschedule.Class{}, fmt.Sprintf("ORDER BY word_score('%s', abbreviation || ' ' || code || ' ' || name) DESC, letter_score('%s', abbreviation) + letter_score('%s', code) + letter_score('%s', name) DESC LIMIT 5", search, search, search, search))
	if err != nil {
		return nil, err
	}
	var classes []goschedule.Class
	for _, record := range records {
		classes = append(classes, record.(goschedule.Class))
	}
	return classes, nil
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
