package tcache

import (
	"bytes"
	"io"
	"sync"
)

// MemStorage keeps track of memAccessor entries in memory using a map
type MemStorage struct {
	mtx   sync.Mutex
	index map[string]*memAccessor
}

func NewMemStorage() Storage {
	return &MemStorage{
		index: map[string]*memAccessor{},
	}
}

func (ms *MemStorage) Get(key string) Accessor {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	be, found := ms.index[key]
	if found {
		return be
	}

	be = &memAccessor{}
	ms.index[key] = be

	return be
}

func (ms *MemStorage) Delete(key string) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	delete(ms.index, key)
}

// memAccessor implements Accessor using bytes.Buffer as backend
type memAccessor struct {
	buffer bytes.Buffer
}

func (be *memAccessor) Reader() (io.Reader, error) {
	return bytes.NewReader(be.buffer.Bytes()), nil
}

func (be *memAccessor) Writer() (io.Writer, error) {
	return &be.buffer, nil
}
