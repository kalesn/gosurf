package util

import (
	"bytes"
	"runtime"
	"strconv"
)

var goPrefix = []byte("goroutine ")

func getGoID() int64 {
	var b = make([]byte, 1<<8)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, goPrefix)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		return -1
	}
	b = b[:i]
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return -1
	}
	return n
}
