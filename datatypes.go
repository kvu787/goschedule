package goschedule

// Position provides the start and end indices of a struct
// extracted from a document.
type position struct {
	start int
	end   int
}

// A College is a UW college that has many departments
type College struct {
	QuarterKey   string
	Name         string
	Abbreviation string
	position
}

// A Department is a UW department that has many classes.
type Dept struct {
	CollegeKey   string
	Name         string
	Abbreviation string
	Link         string
}

// A Class is UW class that has many sections.
type Class struct {
	DeptKey          string
	AbbreviationCode string
	Abbreviation     string
	Code             string
	Title            string
	Description      string
	position
}

// A Sect is a UW section.
type Sect struct {
	ClassKey     string // foreign key
	Restriction  string
	SLN          string // primary key
	Section      string
	Credit       string
	MeetingTimes string // JSON representation
	Instructor   string
	Status       string
	TakenSpots   int
	TotalSpots   int
	Grades       string
	Fee          string
	Other        string
	Info         string
}

// A MeetingTime represents when a Sect is held. Some Sect's
// have multiple meeting times.
type MeetingTime struct {
	Days     string
	Time     string
	Building string
	Room     string
}
