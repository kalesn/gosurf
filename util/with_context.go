package util

import "sync"

func WithLocker(locker sync.Locker, fn func()) {
	locker.Lock()
	defer locker.Unlock()
	fn()
}
