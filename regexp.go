package goschedule

import (
	"regexp"
)

// These regular expressions are used by the extract methods.
var (
	anchorRe              *regexp.Regexp = regexp.MustCompile(`(?is)<a.+?</a>`)
	parenthesesRe         *regexp.Regexp = regexp.MustCompile(`(?is)\(.*?\)`)
	collegeLinkRe         *regexp.Regexp = regexp.MustCompile(`<a.+?\|`)
	collegeChunkRe        *regexp.Regexp = regexp.MustCompile(`(?i)<a name.+?</a>\n<h2>.+?</h2>`)
	classChunkRe          *regexp.Regexp = regexp.MustCompile(`(?is)<table bgcolor="#ffcccc".*?</table>`)
	classNameRe           *regexp.Regexp = regexp.MustCompile(`(?i)name=.*?>`)
	classAbbreviationRe   *regexp.Regexp = regexp.MustCompile(`[a-z]+`)
	classCodeRe           *regexp.Regexp = regexp.MustCompile(`\d+`)
	classTitleRe          *regexp.Regexp = regexp.MustCompile(`(?i)<a href.*?>.+?</a>`)
	tagRe                 *regexp.Regexp = regexp.MustCompile(`(?i)<.+?>`)
	sectChunkRe           *regexp.Regexp = regexp.MustCompile(`(?s).{7}<A HREF=h.+?</td>`)
	meetingTimeRe         *regexp.Regexp = regexp.MustCompile(`(?i)\w{1,5}\s*\d{3,4}-\d{3,4}`)
	spotsRe               *regexp.Regexp = regexp.MustCompile(`\d+`)
	DeptAbbreviationRe    string         = `(?i)\(.+?\)`
	DeptLinkRe            string         = `(?i)href="?.+\.html?"?`
	BlankLineRe           string         = `^\s*$`
	ClassDescriptionChunk string         = `(?is)<p><b><a name=.*?>.*?\n\n`
	ClassDescriptionTitle string         = `(?i)<a name.*?>.+?</a>`
	ClassDescriptionTagRe string         = `(?i)<[^a].+?[^a]>`
)
