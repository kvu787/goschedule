package database

import (
	"encoding/json"
	"strings"
)

// A Parent is a type that will be referenced by
// children.
// ForeignKey returns whatever the children will store
// to establish their relationship to the parent.
type Parent interface {
	ForeignKey() interface{}
}

// Structs that implement Queryer can be stored
// and retrieved from the database with Put and
// Get, respectively.
// PrimaryKey returns the unique identifier for
// the struct.
// TableName returns the name of the database table
// that should be queried.
// NOTE: The corresponding database table must
// be setup with the proper field values.
type Queryer interface {
	PrimaryKey() interface{}
	TableName() string
}

// A Department is a UW department that has many classes
type Dept struct {
	Title        string
	Abbreviation string // primary key
	Link         string
}

// PrimaryKey returns the dept's Abbreviation.
func (d Dept) PrimaryKey() interface{} {
	return d.Abbreviation
}

func (d Dept) TableName() string {
	return "depts"
}

// A Class is UW class that has many sections.
type Class struct {
	DeptAbbreviation string // foreign key
	AbbreviationCode string // primary key
	Abbreviation     string
	Code             string
	Title            string
	Index            int
}

// PrimaryKey returns the class's AbbreviationCode.
func (c Class) PrimaryKey() interface{} {
	return c.AbbreviationCode
}

func (c Class) TableName() string {
	return "classes"
}

// Classes wraps a slice of Class structs so they
// can implement a ForeignKey method.
type Classes []Class

// func (c Classes) SortByCode() []Classes {

// 	for i, v := range Classes {
// 		v.Code
// 	}
// }

// A Class is a UW class represented on the time schedule.
type Sect struct {
	ClassDeptAbbreviation string // foreign key
	Restriction           string
	SLN                   string // primary key
	Section               string
	Credit                string
	MeetingTimes          string // JSON representation, TODO (kvu787): represent as seperate struct
	Instructor            string
	Status                string
	TakenSpots            int
	TotalSpots            int
	Grades                string
	Fee                   string
	Other                 string
	Info                  string
}

// PrimaryKey returns the sect's SLN.
func (s Sect) PrimaryKey() interface{} {
	return s.SLN
}

func (s Sect) TableName() string {
	return "sects"
}

func (s Sect) GetMeetingTimes() ([]MeetingTime, error) {
	var meetingTimes []MeetingTime
	if err := json.Unmarshal([]byte(s.MeetingTimes), &meetingTimes); err != nil {
		return nil, err
	}
	return meetingTimes, nil
}

func (s Sect) IsQuizSection() bool {
	if s.Credit == "QZ" {
		return true
	} else {
		return false
	}
}

func (s Sect) DerivedStatus() string {
	return ""
	// not implemented
}

func (s Sect) IsOpen() bool {
	if s.TotalSpots-s.TakenSpots < 1 {
		return false
	} else {
		return true
	}
}

func (s Sect) GetRestriction() []map[string]bool {
	allTokens := []string{"Restr", "IS", ">"}
	tokens := make([]map[string]bool, len(allTokens))
	for i, v := range allTokens {
		if strings.Contains(s.Restriction, v) {
			tokens[i] = map[string]bool{v: true}
		} else {
			tokens[i] = map[string]bool{v: false}
		}
	}
	return tokens
}

func (s Sect) GetGradesTokens() []map[string]bool {
	allTokens := []string{"CR/NC"}
	tokens := make([]map[string]bool, len(allTokens))
	for i, v := range allTokens {
		if strings.Contains(s.Credit, v) {
			tokens[i] = map[string]bool{v: true}
		} else {
			tokens[i] = map[string]bool{v: false}
		}
	}
	return tokens
}

func (s Sect) GetOtherTokens() []map[string]bool {
	allTokens := []string{"D", "H", "J", "R", "S", "W", "%", "#"}
	tokens := make([]map[string]bool, len(allTokens))
	for i, v := range allTokens {
		if strings.Contains(s.Other, v) {
			tokens[i] = map[string]bool{v: true}
		} else {
			tokens[i] = map[string]bool{v: false}
		}
	}
	return tokens
}

// A MeetingTime represents when a class is held. Some Sect's
// have multiple meeting times.
// A MeetingTime belongs to the Sect with Sln 'SectSln'.
type MeetingTime struct {
	Days     string
	Time     string
	Building string
	Room     string
}

func (m MeetingTime) MapDays() map[string]bool {
	days := strings.ToLower(m.Days)
	dayMap := make(map[string]bool)
	daysSlice := []string{"m", "w", "f", "th", "t"}
	for _, day := range daysSlice {
		if strings.Contains(days, day) {
			dayMap[day] = true
			days = strings.Replace(days, day, "", -1)
		}
	}
	return dayMap
}
