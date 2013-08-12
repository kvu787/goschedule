package extract

// These regular expressions are used by the extract methods.
const (
	DeptChunkRe           string = `(?is)<li><a.+?</a>` // remove tags
	DeptAbbreviationRe    string = `(?i)\(.+?\)`
	DeptLinkRe            string = `(?i)href="?.+\.html?"?`
	TagRe                 string = `(?i)<.+?>`
	ClassChunkRe          string = `(?is)<table bgcolor="#ffcccc".*?</table>`
	ClassNameRe           string = `(?i)name=.*?>` // includes abbreviation and code
	ClassAbbreviationRe   string = `[a-z]+`
	ClassCodeRe           string = `\d+`
	ClassTitleRe          string = `(?i)<a href.*?>.+?</a>`    // remove tags
	SectChunkRe           string = `(?s).{7}<A HREF=h.+?</td>` // remove tags
	SpotsRe               string = `\d+`
	MeetingTimeRe         string = `(?i)\w{1,5}\s*\d{3,4}-\d{3,4}`
	BlankLineRe           string = `^\s*$`
	ClassDescriptionChunk string = `(?is)<p><b><a name=.*?>.*?\n\n`
	ClassDescriptionTitle string = `(?i)<a name.*?>.+?</a>`
	ClassDescriptionTagRe string = `(?i)<[^a].+?[^a]>`
)
