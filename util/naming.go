package util

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile("[A-Z][a-z0-9]*|[a-z0-9]*")

func SnakeCase(s string) string {
	res := re.FindAllString(s, -1)
	strSlice := make([]string, len(res))
	for i, e := range res {
		strSlice[i] = strings.ToLower(e)
	}
	return strings.Join(strSlice, "_")
}

func CamelCase(s string) string {
	var str string
	for _, e := range strings.Split(s, "_") {
		str += strings.Title(e)
	}
	return str
}
