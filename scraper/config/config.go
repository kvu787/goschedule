package config

import "time"

var (
	ScraperTimeout      time.Duration = 5 * time.Minute
	ScraperFetchTimeout time.Duration = 200 * time.Millisecond
)

const (
	RootIndex           string = "https://www.washington.edu/students/timeschd/AUT2013/"
	DeptChunkRe         string = `(?i)<li><a.+?</a>` // remove tags
	DeptAbbreviationRe  string = `(?i)\(.+?\)`
	DeptLinkRe          string = `(?i)href="?.+\.html?"?`
	TagRe               string = `(?i)<.+?>`
	ClassChunkRe        string = `(?is)<table bgcolor="#ffcccc".*?</table>`
	ClassNameRe         string = `(?i)name=.*?>` // includes abbreviation and code
	ClassAbbreviationRe string = `[a-z]+`
	ClassCodeRe         string = `\d+`
	ClassTitleRe        string = `(?i)<a href.*?>.+?</a>`    // remove tags
	SectChunkRe         string = `(?s).{7}<A HREF=h.+?</td>` // remove tags
	SpotsRe             string = `\d+`
	MeetingTimeRe       string = `(?i)\w{1,5}\s*\d{3,4}-\d{3,4}`
	BlankLine           string = `^\s*$`
)
