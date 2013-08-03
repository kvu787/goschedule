package main

import (
	"net/url"
	"strings"
)

type route string

func (r route) Match(u *url.URL) bool {
	if string(r) == "/" || u.String() == "/" {
		return string(r) == u.String()
	}
	rSlice := strings.Split(string(r), "/")[1:]
	uSlice := strings.Split(u.String(), "/")[1:]
	if len(rSlice) != len(uSlice) {
		return false
	}
	return _match(rSlice, uSlice)
}

func _match(rSlice []string, uSlice []string) bool {
	if len(rSlice) == 1 {
		if string(rSlice[0][0]) != ":" {
			return rSlice[0] == uSlice[0]
		}
		return true
	}
	if string(rSlice[0][0]) != ":" {
		if rSlice[0] == uSlice[0] {
			return true && _match(rSlice[1:], uSlice[1:])
		}
		return false
	}
	return true && _match(rSlice[1:], uSlice[1:])
}
