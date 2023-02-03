package util

import "unsafe"

func StringToBytes(s string) []byte {
	uintPtrS := (*[2]uintptr)(unsafe.Pointer(&s))
	uintPtrB := [3]uintptr{uintPtrS[0], uintPtrS[1], uintPtrS[1]}
	return *(*[]byte)(unsafe.Pointer(&uintPtrB))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
