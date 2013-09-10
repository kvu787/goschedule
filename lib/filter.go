package goschedule

import (
	"html"
	"unicode/utf8"
)

// Filter returns a copy of the input string with UTF-8 invalid characters replaced
// with `?` and unescapes HTML escape sequences.
func Filter(in string) string {
	return html.UnescapeString(filterUtf8(in, "?"))
}

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
