package util

import (
	"testing"
	"time"
)

func TestParseTimeYear(t *testing.T) {
	ago := ParseTime("-1 year")
	if dur := time.Now().Sub(ago); dur >= (time.Hour*24*365) && dur <= (time.Hour*24*366)+(time.Second*2) {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestParseTimeMonth(t *testing.T) {
	ago := ParseTime("-1 month")
	if dur := time.Now().Sub(ago); dur >= (time.Hour*24*28) && dur <= (time.Hour*24*31)+(time.Second*2) {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestParseTimeWeek(t *testing.T) {
	ago := ParseTime("-1 week")
	if dur := time.Now().Sub(ago); dur >= (time.Hour*24*7) && dur <= (time.Hour*24*7)+(time.Second*2) {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestParseTimeDay(t *testing.T) {
	ago := ParseTime("-3 days")
	if dur := time.Now().Sub(ago); dur >= (time.Hour*24*3) && dur <= (time.Hour*24*3)+(time.Second*2) {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestParseTimeDuration(t *testing.T) {
	ago := ParseTime("-1 Hour +10 minutes -100 seconds")
	checkTime := time.Now().Add(-(time.Hour * 1) + (time.Minute * 10) - (time.Second * 100))
	if dur := checkTime.Sub(ago); dur >= 0 && dur <= time.Second*2 {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}
