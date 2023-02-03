package util

import (
	"testing"
)

func TestGetGoID(t *testing.T) {
	var n1, n2 int64
	n1 = getGoID()
	n2 = getGoID()
	if n1 != -1 && n2 != -1 && n1 == n2 {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}

func TestGetGoID2(t *testing.T) {
	var (
		flag   = make(chan int)
		n1, n2 int64
	)
	n1 = getGoID()

	go func() {
		n2 = getGoID()
		close(flag)
	}()

	<-flag

	if n1 != -1 && n2 != -1 && n1 != n2 {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}
