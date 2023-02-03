package util

import "testing"

var lockFlag, unlockFlag bool

type lock struct{}

func (*lock) Lock()   { lockFlag = true }
func (*lock) Unlock() { unlockFlag = true }

func TestWithLocker(t *testing.T) {
	l := new(lock)
	WithLocker(l, func() {})
	if lockFlag && unlockFlag {
		t.Log("ok")
	} else {
		t.Error("fail")
	}
}
