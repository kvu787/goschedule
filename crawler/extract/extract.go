// Package extract provides structs for UW time
// schedule information and methods to extract
// them from a webpage.
package extract

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"

	"github.com/kvu787/go_schedule/crawler/config"
	"github.com/kvu787/go_schedule/crawler/database"
)

// Extracter is implemented by any byteslice from which
// information can be extracted from with Extract.
// p is a (makeshift) optional argument that
// will act as a parent/ancestor to the extracted
// objects.
type Extracter interface {
	Extract(parent database.Queryer, db *sql.DB) interface{}
}

// DeptIndex wraps a webpage that contains an index
// of UW departments.
type DeptIndex []byte

// Extract grabs department information from an index of
// UW departments.
// Parameter p Parent should be nil.
// Returns a slice of Dept structs.
func (body DeptIndex) Extract(parent database.Queryer) []database.Dept {
	root := config.RootIndex
	depts := []database.Dept{}
	cks := matchRegexp(body, config.DeptChunkRe)
	for _, ck := range cks {
		var dept database.Dept
		// grab link
		urlFrag := findRegexp(ck, config.DeptLinkRe, true)
		urlFrag = removeRegexp(urlFrag, `"`, `href=`)
		dept.Link = root + string(urlFrag)
		// grab title
		fullTitle := removeRegexp(ck, config.TagRe)
		fullTitle = replaceRegexp(fullTitle, `&nbsp;`, " ")
		fullTitle = replaceRegexp(fullTitle, `&amp;`, "&")
		dept.Title = string(bytes.TrimSpace(removeRegexp(fullTitle, config.DeptAbbreviationRe)))
		// grab abbreviation
		if abbreviation := findRegexp(fullTitle, config.DeptAbbreviationRe, false); abbreviation != nil {
			abbreviation = removeRegexp(abbreviation, `\(`, `\)`)
			abbreviation = replaceRegexp(abbreviation, " ", "")
			dept.Abbreviation = string(bytes.ToUpper(abbreviation))
		} else {
			// skip dept if abbreviation not found
			continue
		}
		// append to slice
		depts = append(depts, dept)
	}
	return depts
}

// ClassIndex wraps a webpage that contains an index
// of UW class headings.
type ClassIndex []byte

// Extract grabs class information from an index of
// UW class headings.
// Parameter p Parent should be a Dept struct.
// Returns a slice of Class structs.
func (body ClassIndex) Extract(parent database.Queryer) []database.Class {
	dept := parent.(database.Dept)
	classes := []database.Class{}
	cks := matchRegexp(body, config.ClassChunkRe)
	indx := matchIndexRegexp(body, config.ClassChunkRe)
	// loop through chunks and extract fields
	for i, ck := range cks {
		var class database.Class
		// set child relation to dept
		class.DeptAbbreviation = dept.PrimaryKey().(string)
		// grab name (abbreviation and code)
		name := findRegexp(ck, config.ClassNameRe, true)
		// grab abbreviation
		class.Abbreviation = string(bytes.ToLower(findRegexp(name, config.ClassAbbreviationRe, true)))
		// grab code
		class.Code = string(findRegexp(name, config.ClassCodeRe, true))
		// grab title
		title := findRegexp(ck, config.ClassTitleRe, true)
		class.Title = string(removeRegexp(title, config.TagRe))
		// set AbbreviationCode key
		class.AbbreviationCode = class.Abbreviation + class.Code
		// set index position
		class.Index = indx[i][0]
		// append to slice
		classes = append(classes, class)
	}
	return classes // already sorted by index
}

// SectIndex wraps a webpage that contains an index
// of UW class headings.
type SectIndex []byte

// Extract grabs class information from an index of
// UW sections.
// Parameter p Parent accepts a slice of Class structs.
// Returns a slice of Sect structss.
func (body SectIndex) Extract(parent []database.Class) []database.Sect {
	classes := parent
	sects := []database.Sect{}
	// get sec chunks and indices
	cks := matchRegexp(body, config.SectChunkRe)
	indx := matchIndexRegexp(body, config.SectChunkRe)
	// loop through chunks and extract fields
	for i, ck := range cks {
		var sect database.Sect
		ck = removeRegexp(ck, config.TagRe)
		// assign class information to sect
		if len(classes) > 1 && classes[1].Index < indx[i][0] { // check if next class in queue has lower index than current sect
			classes = classes[1:] // pop queue
		}
		sect.ClassDeptAbbreviation = classes[0].PrimaryKey().(string)
		// split chunks into lines
		var ckLns [][]byte = bytes.Split(ck, []byte("\n"))
		// collect meeting times in slice, to be converted to JSON later
		var meetingTimes []database.MeetingTime
		// check first line for meeting time info
		if mt, err := checkMeetingTime(ckLns[0]); err == nil {
			meetingTimes = append(meetingTimes, mt)
		}
		// check first line for rest of fields
		sect.Restriction = string(bytes.TrimSpace(ckLns[0][0:7]))
		sect.SLN = string(bytes.TrimSpace(ckLns[0][7:13]))
		sect.Section = string(bytes.TrimSpace(ckLns[0][13:16]))
		sect.Credit = string(bytes.TrimSpace(ckLns[0][16:24]))
		sect.Instructor = string(bytes.TrimSpace(ckLns[0][56:83]))
		sect.Status = string(bytes.TrimSpace(ckLns[0][83:89]))
		spots := bytes.TrimSpace(ckLns[0][89:101])
		if m := matchRegexp(spots, config.SpotsRe); m != nil {
			sect.TakenSpots, _ = strconv.Atoi(string(m[0]))
			sect.TotalSpots, _ = strconv.Atoi(string(m[1]))
		}
		sect.Grades = string(bytes.TrimSpace(ckLns[0][101:108]))
		sect.Fee = string(bytes.TrimSpace(ckLns[0][108:115]))
		sect.Other = string(bytes.TrimSpace(ckLns[0][115:]))
		// crawl through other lines
		ckLns = ckLns[1:]
		for _, ln := range ckLns {
			// check if MeetingTime
			if mt, err := checkMeetingTime(ln); err == nil {
				meetingTimes = append(meetingTimes, mt)
			} else if re := regexp.MustCompile(config.BlankLine); re.Match(ln) {
				// skip if blank line
			} else { // else append to sect.Info
				sect.Info += string(bytes.TrimSpace(ln)) + "\n"
			}
		}
		// store JSON representation of MeetingTimes
		mtJSON, err := json.Marshal(meetingTimes)
		if err != nil {
			continue // skip if meeting time doesn't store
		}
		sect.MeetingTimes = string(mtJSON)
		// append finished sect
		sects = append(sects, sect)
	}
	return sects
}

// checkMeetingTime checks if a byteslice contains
// information for a MeetingTime struct.
// If a MeetingTime is found, it is returned with nil
// error. Else, nil is returned for MeetingTime and non
// nil for error.
func checkMeetingTime(ln []byte) (database.MeetingTime, error) {
	var mt database.MeetingTime
	re := regexp.MustCompile(config.MeetingTimeRe)
	if re.FindAllIndex(ln, -1) != nil {
		mt.Days = string(bytes.TrimSpace(ln[24:31]))
		mt.Time = string(bytes.TrimSpace(ln[31:42]))
		mt.Building = string(bytes.TrimSpace(ln[42:47]))
		mt.Room = string(bytes.TrimSpace(ln[47:56]))
		return mt, nil
	}
	return database.MeetingTime{}, errors.New("Meeting time pattern not found.")
}

// replaceRegexp replaces all matches by the regexp r with
// the string repl in the byteslice b.
func replaceRegexp(b []byte, r string, repl string) []byte {
	re := regexp.MustCompile(r)
	return re.ReplaceAll(b, []byte(repl))
}

// removeRegexp uses each given regxp in sequence to
// match and remove patterns in the string.
func removeRegexp(b []byte, r ...string) []byte {
	for _, v := range r {
		re := regexp.MustCompile(v)
		b = re.ReplaceAll(b, []byte{})
	}
	return b
}

// matchRegex uses the given regexp to return an array
// of all matches in the byteslice.
// Returns nil if there are no matches.
func matchRegexp(b []byte, r string) [][]byte {
	re := regexp.MustCompile(r)
	if match := re.FindAll(b, -1); match == nil {
		return nil
	} else {
		return match
	}
}

// matchIndexRegex uses the given regexp to return an array
// of arrays indicating the start and stop indices of
// each match in the byteslice.
func matchIndexRegexp(b []byte, r string) [][]int {
	re := regexp.MustCompile(r)
	return re.FindAllIndex(b, -1)
}

// findRegexp uses the given regexp to return a match
// if found in the byteslice.
// Returns nil if no match found.
func findRegexp(b []byte, r string, first bool) []byte {
	re := regexp.MustCompile(r)
	matches := re.FindAll(b, -1)
	if len(matches) == 0 {
		return nil
	}
	if first {
		return matches[0]
	} else {
		return matches[len(matches)-1]
	}
}
