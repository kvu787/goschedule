// Package goschedule is a library for extracting data from the
// UW time schedule.
package goschedule

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// htmlDecoder that creates an XML decoder that can handle typical HTML.
// See http://golang.org/pkg/encoding/xml/#Decoder.
func newHtmlDecoder(s string) *xml.Decoder {
	decoder := xml.NewDecoder(strings.NewReader(s))
	decoder.Strict = false
	decoder.Entity = xml.HTMLEntity
	decoder.AutoClose = xml.HTMLAutoClose
	return decoder
}

// cleanHref surrounds the value of an href tag with quotes if it is unquoted.
func cleanHref(s string) string {
	return regexp.MustCompile(`(?i)href=(?:"|')?(.+)[.](html?)(?:"|')?`).ReplaceAllString(s, `href="$1.$2"`)
}

type errorsSlice []error

func (e errorsSlice) Error() string {
	var errStrings string
	for _, err := range e {
		errStrings += err.Error() + "; "
	}
	return errStrings
}

// ExtractColleges grabs College structs from a string
func ExtractColleges(content string) ([]College, error) {
	var colleges []College
	var errs errorsSlice
	// process hash links
	for _, match := range collegeLinkRe.FindAllString(content, -1) {
		var college College
		// parse links from xml
		tag := struct {
			Href    string `xml:"href,attr"`
			Content string `xml:",innerxml"`
		}{}
		decoder := newHtmlDecoder(match)
		if err := decoder.Decode(&tag); err != nil && err != io.EOF {
			errs = append(errs, fmt.Errorf("skipped a college: error unmarshalling xml (%s): %v", string(match), err))
			continue
		}
		abbreviation := strings.TrimPrefix(strings.TrimSpace(tag.Href), "#")
		// set attributes
		college.Abbreviation = abbreviation
		college.Name = tag.Content
		// setup regex to get positions
		collegeRe := regexp.MustCompile(fmt.Sprintf(`(?i)<a name="%s.+?</a>\n<h2>.+?</h2>((?s).*?)<a name=".+?</a>\n<h2>.+?</h2>`, abbreviation))
		if position := collegeRe.FindStringIndex(content); position != nil {
			college.Start = position[0]
			college.End = position[1]
		} else {
			startPosition := regexp.MustCompile(
				fmt.Sprintf(`(?i)<a name="%s.+?</a>\n<h2>.+?</h2>`, abbreviation)).
				FindStringIndex(content) // will panic if doesn't find a match
			if startPosition == nil {
				errs = append(errs, fmt.Errorf(`skipped college: could not find abbreviation in main body: "%s"`, college.Abbreviation))
				continue
			}
			college.Start = startPosition[0]
			college.End = len(content)
		}
		colleges = append(colleges, college)
	}
	if len(errs) > 0 {
		return colleges, errs
	} else {
		return colleges, nil
	}
}

// Extract grabs Dept structs from a string.
//
// All Dept structs in the returned slice will use collegeKey as their collegeKey attribute.
// processed is a map of Dept.Abbreviation's that have already been processed. The int values
// are not used.
// ExtractDepts will skip a Dept if its abbreviation is in processed. Else, it will add the
// abbreviation to processed.
//
// Note that the department's abbreviation (primary key) cannot be scraped from the department index.
// The abbreviation from the class listing by visiting the department page. Use
// Dept.ScrapeAbbreviation with a class index page (Dept.Link).
func ExtractDepts(content, collegeKey, url string, processed *map[string]int) ([]Dept, error) {
	var depts []Dept
	var errs errorsSlice
	for _, match := range anchorRe.FindAllString(content, -1) {
		// check validity
		tag := struct {
			Href    string `xml:"href,attr"`
			Content string `xml:",innerxml"`
		}{}
		decoder := newHtmlDecoder(cleanHref(match))
		if err := decoder.Decode(&tag); err != nil && err != io.EOF {
			errs = append(errs, fmt.Errorf("skipped a department: error unmarshalling xml (%s): %v", string(match), err))
			continue
		}
		tag.Href = strings.TrimSpace(tag.Href)
		tag.Content = strings.TrimSpace(tag.Content)
		if valid := validateDept(tag.Href, tag.Content); !valid {
			continue
		}
		// create Dept
		var dept Dept
		// grab link
		dept.Link = url + string(tag.Href)
		// grab title
		dept.Name = strings.TrimSpace(parenthesesRe.ReplaceAllString(tag.Content, ""))
		// grab href
		var href string
		if temp := strings.Split(tag.Href, "."); len(temp) > 0 {
			href = temp[0]
		} else {
			errs = append(errs, fmt.Errorf(`skipped department: invalid href format: "%s"`, tag.Href))
			continue
		}
		// check department for uniqueness
		if _, exists := (*processed)[href]; exists {
			continue
		} else { // add to map if unique
			(*processed)[href] = 1
		}
		// add college
		dept.CollegeKey = collegeKey
		// add href to map
		depts = append(depts, dept)
	}
	if len(errs) > 0 {
		return depts, errs
	} else {
		return depts, nil
	}
}

// validateDept checks the elements of a Dept for validity.
func validateDept(href, content string) bool {
	if len(href) < 1 {
		return false
	}
	if href[0] == '#' {
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
	var classes []Class
	matchIndices := classChunkRe.FindAllStringIndex(content, -1)
	for _, match := range classChunkRe.FindAllString(content, -1) {
		var class Class
		class.DeptKey = deptKey
		// grab name (abbreviation and code)
		name := classNameRe.FindString(match)
		// grab abbreviation
		class.Abbreviation = strings.ToLower(classAbbreviationRe.FindString(name))
		// grab code
		class.Code = strings.ToLower(classCodeRe.FindString(name))
		// grab title
		class.Title = strings.ToLower(tagRe.ReplaceAllString(classTitleRe.FindString(match), ""))
		// set AbbreviationCode key
		class.AbbreviationCode = class.Abbreviation + class.Code
		// append to slice
		classes = append(classes, class)
	}
	// set class positions
	for i := 0; i < len(classes); i++ {
		classes[i].Start = matchIndices[i][0]
		if i == len(classes)-1 {
			classes[i].End = len(content)
			break
		}
		classes[i].End = matchIndices[i+1][0]
	}
	return classes
}

// ExtractSects grabs Sect structs from a string. All Sect structs
// in the returned slice will use classKey as their ClassKey attribute.
func ExtractSects(content, classKey string) ([]Sect, error) {
	var sects []Sect
	var errs errorsSlice
	for _, match := range sectChunkRe.FindAllString(content, -1) {
		var sect Sect
		// remove html tags
		match = tagRe.ReplaceAllString(match, "")
		lines := strings.Split(match, "\n")
		var meetingTimes []MeetingTime
		// check first line for meeting time
		if mt, err := checkMeetingTime(lines[0]); err == nil {
			meetingTimes = append(meetingTimes, mt)
		}
		// extract sect attributes from rest of first line
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
		// crawl through other lines
		lines = lines[1:]
		for _, line := range lines {
			// check if MeetingTime
			if mt, err := checkMeetingTime(line); err == nil {
				meetingTimes = append(meetingTimes, mt)
			} else if blankLineRe.MatchString(line) {
				// skip if blank line
			} else { // else append to sect.Info
				sect.Info += strings.TrimSpace(line) + "\n"
			}
		}
		// store JSON representation of MeetingTime's
		mtJson, err := json.Marshal(meetingTimes)
		if err != nil {
			errs = append(errs, err)
			sect.MeetingTimes = "error"
		} else {
			sect.MeetingTimes = string(mtJson)
		}
		sect.ClassKey = classKey
		sects = append(sects, sect)
	}
	if len(errs) < 1 {
		return sects, nil
	} else {
		return sects, errs
	}
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

// ExtractClassDescriptions extracts class descriptions from content. It returns a
// map of class abbreviationCode's (primary key) to class descriptions.
func ExtractClassDescriptions(content string) (map[string]string, error) {
	descriptions := make(map[string]string)
	for _, match := range classDescriptionRe.FindAllStringSubmatch(content, -1) {
		if len(match) < 3 {
			return nil, fmt.Errorf("less than 3 submatches found: %q", match)
		}
		// grab class AbbreviationCode
		abbreviationCode := strings.ToLower(strings.TrimSpace(match[1]))
		// store abbreviationCode and description as key-value pair
		descriptions[abbreviationCode] = strings.TrimSpace(match[2])
	}
	return descriptions, nil
}
