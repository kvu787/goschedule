package database

import (
	"encoding/json"
	"testing"
)

func TestGetMeetingTimes(t *testing.T) {
	mts := []MeetingTime{
		MeetingTime{123, "a", "b", "c", "d"},
		MeetingTime{345, "f", "g", "h", "di"},
	}
	byteJSON, _ := json.Marshal(mts)
	strJSON := string(byteJSON)

	sect := Sect{MeetingTimes: strJSON}
	getMts, err := sect.GetMeetingTimes()
	if err != nil {
		t.Errorf("GetMeetingTimes error")
	}
	if len(getMts) != 2 {
		t.Errorf("GetMeetingTimes fail: expected 2 MeetingTimes in slice")
	}
}
