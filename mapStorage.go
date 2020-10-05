package tcache

import (
	"sync"
)

// MapStorage stores BufferEntry entries in memory using a map
type MapStorage struct {
	mtx   sync.Mutex
	index map[string]*BufferEntry
}

func NewMapStorage() Storage {
	return &MapStorage{
		index: map[string]*BufferEntry{},
	}
}

func (ms *MapStorage) Get(key string) Accessor {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	be, found := ms.index[key]
	if found {
		return be
	}

	be = &BufferEntry{}
	ms.index[key] = be

	return be
}

func (ms *MapStorage) Delete(key string) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	delete(ms.index, key)
}
