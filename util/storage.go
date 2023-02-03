package util

import "sync"

type Storage interface {
	Get(key interface{}) (val interface{})
	Set(key, val interface{})
	Del(key interface{})
	Clear() (n int)
}

type ShardingStorage struct {
	Max      int
	Fn       func(key interface{}) (i int)
	Storages []*storage
}

type storage struct {
	rw *sync.RWMutex
	m  map[interface{}]interface{}
}

func alwaysOne(_ interface{}) int { return 0 }

func NewStorage(max int, fn func(key interface{}) (i int)) Storage {
	if fn == nil {
		fn = alwaysOne
		max = 1
	}
	if max < 1 {
		max = 1
	}
	ss := &ShardingStorage{
		Max: max,
		Fn:  fn,
	}
	var storages = make([]*storage, max)
	for i := 0; i < max; i++ {
		storages[i] = &storage{
			rw: new(sync.RWMutex),
			m:  make(map[interface{}]interface{}),
		}
	}
	ss.Storages = storages
	return ss
}

func (ss *ShardingStorage) Set(key, val interface{}) {
	i := ss.Fn(key)
	if i >= ss.Max || i < 0 {
		i = 0
	}
	s := ss.Storages[i]
	s.rw.Lock()
	defer s.rw.Unlock()
	s.m[key] = val
}

func (ss *ShardingStorage) Get(key interface{}) (val interface{}) {
	i := ss.Fn(key)
	if i >= ss.Max || i < 0 {
		i = 0
	}
	s := ss.Storages[i]
	s.rw.RLock()
	defer s.rw.RUnlock()
	val = s.m[key]
	return val
}

func (ss *ShardingStorage) Del(key interface{}) {
	i := ss.Fn(key)
	if i >= ss.Max || i < 0 {
		i = 0
	}
	s := ss.Storages[i]
	s.rw.Lock()
	defer s.rw.Unlock()
	delete(s.m, key)
}

func (ss *ShardingStorage) Clear() (n int) {
	for _, s := range ss.Storages {
		s.rw.Lock()
		n += len(s.m)
		s.m = make(map[interface{}]interface{})
		s.rw.Unlock()
	}
	return n
}

// SyncMapStorage use sync.map as a storage
type SyncMapStorage struct {
	m *sync.Map
}

func NewSyncStorage() Storage {
	return &SyncMapStorage{m: new(sync.Map)}
}

func (smap *SyncMapStorage) Get(key interface{}) (val interface{}) {
	val, _ = smap.m.Load(key)
	return val
}

func (smap *SyncMapStorage) Set(key, val interface{}) {
	smap.m.Store(key, val)
}

func (smap *SyncMapStorage) Del(key interface{}) {
	smap.m.Delete(key)
}

func (smap *SyncMapStorage) Clear() (n int) {
	smap.m.Range(func(k, v interface{}) bool {
		smap.Del(k)
		n++
		return true
	})
	return n
}
