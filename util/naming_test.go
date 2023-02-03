package util

import "testing"

func TestCamelCase(t *testing.T) {
	if CamelCase("abc_test") == "AbcTest" {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestSnakeCase(t *testing.T) {
	if SnakeCase("AbcTest") == "abc_test" && SnakeCase("abcTest") == "abc_test" {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}
