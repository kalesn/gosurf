package util

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reYear   = regexp.MustCompile(`([+-]?\d+)\s*years?`)
	reMonth  = regexp.MustCompile(`([+-]?\d+)\s*months?`)
	reWeek   = regexp.MustCompile(`([+-]?\d+)\s*weeks?`)
	reDay    = regexp.MustCompile(`([+-]?\d+)\s*days?`)
	reHour   = regexp.MustCompile(`([+-]?\d+)\s*hours?`)
	reMinute = regexp.MustCompile(`([+-]?\d+)\s*minutes?`)
	reSecond = regexp.MustCompile(`([+-]?\d+)\s*seconds?`)
)

func ParseTime(s string, args ...time.Time) time.Time {
	var t time.Time

	switch len(args) {
	case 0:
		t = time.Now()
	case 1:
		t = args[0]
	default:
		panic("parse: only receive 1 time.Time argument")
	}

	var (
		years, months, weeks, days int
		hours, minutes, seconds    int64
	)

	s = strings.ToLower(s)

	if match := reYear.FindStringSubmatch(s); len(match) > 1 {
		years, _ = strconv.Atoi(match[1])
	}
	if match := reMonth.FindStringSubmatch(s); len(match) > 1 {
		months, _ = strconv.Atoi(match[1])
	}
	if match := reWeek.FindStringSubmatch(s); len(match) > 1 {
		weeks, _ = strconv.Atoi(match[1])
	}
	if match := reDay.FindStringSubmatch(s); len(match) > 1 {
		days, _ = strconv.Atoi(match[1])
	}
	if match := reHour.FindStringSubmatch(s); len(match) > 1 {
		hours, _ = strconv.ParseInt(match[1], 10, 64)
	}
	if match := reMinute.FindStringSubmatch(s); len(match) > 1 {
		minutes, _ = strconv.ParseInt(match[1], 10, 64)
	}
	if match := reSecond.FindStringSubmatch(s); len(match) > 1 {
		seconds, _ = strconv.ParseInt(match[1], 10, 64)
	}

	t = t.AddDate(years, months, 7*weeks+days)
	t = t.Add(
		time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second,
	)

	return t
}
