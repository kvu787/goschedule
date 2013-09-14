package frontend

import (
	"strings"
)

// route is a string that indicates a URL routing pattern.
type route string

// match assumes that route r is well-formed.
// Note that requesting a domain without a path will return a *url.URL where url.Path == "/".
// Ex. "example.com" and "example.com/" both have a path of "/".
// However, "example.com/asfd" and "example.com/asdf/" have paths of "/asdf" and "/asdf/", respectively.
func (ro route) match(path string) bool {
	route := string(ro)
	switch {
	case route == "/" && path == "/":
		return true
	case route == "/" && path != "/":
		return false
	case route != "/" && path == "/":
		return false
	}
	routeSlice := strings.Split(route, "/")[1:]
	pathSlice := strings.Split(path, "/")[1:]
	if len(routeSlice) != len(pathSlice) {
		return false
	}
	return _match(routeSlice, pathSlice)
}

func _match(routeSlice []string, pathSlice []string) bool {
	r := routeSlice[0]
	p := pathSlice[0]
	// terminating condition
	if len(routeSlice) == 1 {
		if r[0] != ':' {
			if r == p {
				return true
			}
			return false
		}
		return true
	}
	// recursion
	if r[0] != ':' {
		if r == p {
			return true && _match(routeSlice[1:], pathSlice[1:])
		}
		return false
	}
	return true && _match(routeSlice[1:], pathSlice[1:])
}
