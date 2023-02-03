package util

import (
	"reflect"
	"testing"
)

const s = "test测试"

func TestStringToBytes(t *testing.T) {
	x, y := StringToBytes(s), []byte(s)
	if reflect.DeepEqual(x, y) {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestBytesToString(t *testing.T) {
	b := []byte(s)
	x, y := BytesToString(b), string(b)
	if x == y {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}
