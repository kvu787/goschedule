package goschedule

import (
	"testing"
)

func TestExtractColleges(t *testing.T) {
	content := `<a href="#AUP">Architecture and Urban Planning</a> |
                <a href="#AS">Arts &amp; Sciences</a> |
                <a href="#AUP">Built Environments</a> |
                <a href="#B">Business School</a> |
                <!-- <a href="#CCS">Center for Career Serv</a> | -->
                <a href="#D">Dentistry</a> |
                <a href="#ED">Education</a> |


                <a name="ED"></a>
                <h2>college name</h2>
                some department content...

                <a name="AS"></a>
                <h2>college name</h2>
                some department content...`
	colleges, err := ExtractColleges(content, "")
	if len(colleges) != 2 || err != nil {
		t.Errorf("case: %q", content)
	}
}

func TestExtractDepts(t *testing.T) {
	testSet := []struct {
		content string
		length  int
		err     bool
	}{
		{
			``, 0, false,
		},
		{
			`
            <a href="cse.html">CS (CSE)</a>
            <a href="#cse">CS (CSE)</a>
            <a href="math.html">Math (math)</a>
            <a href="biol.html">Biology</a>`, 2, false,
		},
	}
	for _, test := range testSet {
		depts, err := ExtractDepts(test.content, "a college", "uw.edu/")
		if (len(depts) != test.length) || ((err != nil) != test.err) {
			t.Errorf("case: %q", test.content)
		}
	}
}

func TestValidateDept(t *testing.T) {
	hrefs := map[string]int{
		"math.html": 1,
	}
	testSet := []struct {
		href     string
		content  string
		expected bool
	}{
		{`cse.html`, `Computer Science and Engineering (CSE)`, true},
		{`cse.html`, `Computer Science and Engineering (Comp Sci) (CSE)`, true},
		{`#CSE`, `Computer Science and Engineering (CSE)`, false},
		{`cse.html`, `Computer Science and Engineering`, false},
		{`math.html`, `Computer Science and Engineering (CSE)`, false},
	}
	for _, test := range testSet {
		if validateDept(test.href, test.content, hrefs) != test.expected {
			t.Errorf("case %v", test)
		}
	}
}
