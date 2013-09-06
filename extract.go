// Package goschedule is a library for extracting data from the
// UW time schedule.
package goschedule

import (
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// filterUtf8 is replaces, "?" all invalid characters (per the UTF-8 encoding
// of Unicode) with the repl.
func filterUtf8(in, repl string) (out string) {
	for _, r := range []rune(in) {
		if r != utf8.RuneError {
			out += string(r)
		} else {
			out += repl
		}
	}
	return out
}

// ExtractColleges grabs College structs from a string
func ExtractColleges(content string) ([]College, error) {
	content = filterUtf8(content, "?")
	var colleges []College
	var errString string

	// process hash links
	matches := collegeLinkRe.FindAllString(content, -1)
	for _, match := range matches {
		var college College // Name, Abbreviation, position

		// parse links from xml
		tag := struct {
			Href    string `xml:"href,attr"`
			Content string `xml:",innerxml"`
		}{}
		if err := xml.Unmarshal([]byte(match), &tag); err != nil {
			errString += err.Error() + ", "
			continue
		}
		abbreviation := strings.TrimPrefix(strings.TrimSpace(tag.Href), "#")

		// set attributes
		college.Abbreviation = abbreviation
		college.Name = html.UnescapeString(tag.Content)

		// setup regex to get positions
		collegeRe := regexp.MustCompile(fmt.Sprintf(`(?i)<a name="%s.+?</a>\n<h2>.+?</h2>((?s).*?)<a name=".+?</a>\n<h2>.+?</h2>`, abbreviation))
		if position := collegeRe.FindStringIndex(content); position != nil {
			college.start = position[0]
			college.end = position[1]
		} else {
			startPosition := regexp.MustCompile(
				fmt.Sprintf(`(?i)<a name="%s.+?</a>\n<h2>.+?</h2>`, abbreviation)).
				FindStringIndex(content) // will panic if doesn't find a match
			if startPosition == nil {
				continue
			}
			college.start = startPosition[0]
			college.end = len(content)
		}
		colleges = append(colleges, college)
	}
	if len(errString) > 0 {
		return colleges, fmt.Errorf(errString)
	} else {
		return colleges, nil
	}
}

// Extract grabs Dept structs from a string. All Dept structs in the
// returned slice will use collegeKey as their collegeKey attribute.
func ExtractDepts(content, collegeKey, url string) ([]Dept, error) {
	content = filterUtf8(content, "?")
	var depts []Dept
	var errString string
	var hrefs = map[string]int{}

	matches := anchorRe.FindAllString(content, -1)
	for _, match := range matches {
		// check validity
		tag := struct {
			Href    string `xml:"href,attr"`
			Content string `xml:",innerxml"`
		}{}
		if err := xml.Unmarshal([]byte(match), &tag); err != nil {
			errString += err.Error() + ", "
			continue
		}
		tag.Href = strings.TrimSpace(tag.Href)
		tag.Content = strings.TrimSpace(tag.Content)
		if valid := validateDept(tag.Href, tag.Content, hrefs); !valid {
			continue
		}

		// create Dept
		var dept Dept
		// grab link
		dept.Link = url + string(tag.Href)
		// grab title
		dept.Name = html.UnescapeString(
			strings.TrimSpace(
				parenthesesRe.ReplaceAllString(tag.Content, "")))
		// grab abbreviation
		if temp := strings.Split(tag.Href, "."); len(temp) > 0 {
			dept.Abbreviation = temp[0]
		}
		// add college
		dept.CollegeKey = collegeKey
		// add href to map
		hrefs[tag.Href] = 0
		depts = append(depts, dept)
	}
	if len(errString) > 0 {
		return depts, fmt.Errorf(errString)
	} else {
		return depts, nil
	}
}

// validateDept checks the elements of a Dept for validiting. Also
// checks that the Dept has not already been processed in hrefs.
func validateDept(href, content string, hrefs map[string]int) bool {
	if len(href) < 1 {
		return false
	}
	if href[0] == '#' {
		return false
	}
	if _, exists := hrefs[string(href)]; exists {
		return false
	}
	if parenthesesRe.FindString(content) == "" {
		return false
	}
	return true
}

// ExtractClasses grabs Class structs from a string. All Class structs
// in the returned slice will use deptKey as their DeptKey attribute.
func ExtractClasses(content, deptKey string) []Class {
	content = filterUtf8(content, "?")
	var classes []Class

	matches := classChunkRe.FindAllString(content, -1)
	matchIndices := classChunkRe.FindAllStringIndex(content, -1)
	for _, match := range matches {
		var class Class
		class.DeptKey = deptKey
		// grab name (abbreviation and code)
		name := classNameRe.FindString(match)
		// grab abbreviation
		class.Abbreviation = strings.ToLower(
			classAbbreviationRe.FindString(name))
		// grab code
		class.Code = strings.ToLower(
			classCodeRe.FindString(name))
		// grab title
		class.Title = strings.ToLower(
			tagRe.ReplaceAllString(
				classTitleRe.FindString(match), ""))
		// set AbbreviationCode key
		class.AbbreviationCode = class.Abbreviation + class.Code
		// append to slice
		classes = append(classes, class)
	}
	// set Class positions
	for i, class := range classes {
		class.start = matchIndices[i][0]
		if i == len(classes)-1 {
			class.end = len(content)
			break
		}
		class.end = matchIndices[i+1][0]
	}
	return classes
}

// ExtractSects grabs Sect structs from a string. All Sect structs
// in the returned slice will use classKey as their ClassKey attribute.
func ExtractSects(content, classKey string) []Sect {
	content = filterUtf8(content, "?")
	var sects []Sect

	matches := sectChunkRe.FindAllString(content, -1)
	for _, match := range matches {
		var sect Sect

		match = tagRe.ReplaceAllString(match, "")
		lines := strings.Split(match, "\n")

		var meetingTimes []MeetingTime
		if mt, err := checkMeetingTime(lines[0]); err == nil {
			meetingTimes = append(meetingTimes, mt)
		}

		sect.Restriction = strings.TrimSpace(lines[0][0:7])
		sect.SLN = strings.ToLower(strings.TrimSpace(lines[0][7:13]))
		sect.Section = strings.TrimSpace(lines[0][13:16])
		sect.Credit = strings.TrimSpace(lines[0][16:24])
		sect.Instructor = strings.TrimSpace(lines[0][56:83])
		sect.Status = strings.TrimSpace(lines[0][83:89])
		spots := strings.TrimSpace(lines[0][89:101])
		if m := spotsRe.FindAllString(spots, -1); len(m) > 1 {
			sect.TakenSpots, _ = strconv.ParseInt(m[0], 10, 64)
			sect.TotalSpots, _ = strconv.ParseInt(m[1], 10, 64)
		}

		sect.Grades = strings.TrimSpace(lines[0][101:108])
		sect.Fee = strings.TrimSpace(lines[0][108:115])
		sect.Other = strings.TrimSpace(lines[0][115:])
	}
	return sects
}

// checkMeetingTime checks if a string contains information for
// a MeetingTime struct.
// If a MeetingTime is found, it is returned with nil error. Else,
// nil is returned for MeetingTime and non-nil for error.
func checkMeetingTime(line string) (MeetingTime, error) {
	var mt MeetingTime
	if meetingTimeRe.FindString(line) != "" {
		mt.Days = strings.TrimSpace(line[24:31])
		mt.Time = strings.TrimSpace(line[31:42])
		mt.Building = strings.TrimSpace(line[42:47])
		mt.Room = strings.TrimSpace(line[47:56])
		return mt, nil
	}
	return mt, fmt.Errorf("")
}
