package main

import (
	"net/url"
	"testing"
)

func TestMatchRoute(t *testing.T) {
	if !route("/chicken").Match(&url.URL{Path: "/chicken"}) {
		t.Errorf("fail1")
	}
	if !route("/chicken/:coop").Match(&url.URL{Path: "/chicken/one"}) {
		t.Errorf("fail2")
	}
	if !route("/schedule/:dept/:class").Match(&url.URL{Path: "/schedule/cse/cse102"}) {
		t.Errorf("fail3")
	}
	if route("/schedule/:dept/info").Match(&url.URL{Path: "/schedule/cse/cse102"}) {
		t.Errorf("fail4")
	}
	if route("/chicken/").Match(&url.URL{Path: "/chicken"}) {
		t.Errorf("fail5")
	}
	if !route("/").Match(&url.URL{Path: "/"}) {
		t.Errorf("fail6")
	}
	if route("/chicken").Match(&url.URL{Path: "/"}) {
		t.Errorf("fail7")
	}
	if route("/schedule/:dept").Match(&url.URL{Path: "/schedule"}) {
		t.Errorf("fail8")
	}
	if !route("/assets/:type/:file").Match(&url.URL{Path: "/assets/css/bootstrap.min.css"}) {
		t.Errorf("fail9")
	}

}
