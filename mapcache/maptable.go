package mapcache

import (
	"roob.re/tcache"
	"sync"
	"time"
)

type mapTable struct {
	sync.Mutex
	rows map[string]*mapEntry
}

func (mt *mapTable) Access(key string, maxAge time.Duration, handler tcache.Handler) error {
	mt.Lock()
	entry, entryFound := mt.rows[key]

	handlerErr := tcache.EntryMissingError
	if entryFound {
		mt.Unlock()

		entry.RLock()
		handlerErr = entry.HandleRead(maxAge, handler.Then)
		entry.RUnlock()

		if handlerErr == nil {
			return nil
		}

		entry.Lock()
		entry.Invalidate()
		entry.Unlock()

		mt.Lock()
		delete(mt.rows, key)
	}

	// Exit early if we do not have an else handler
	if handler.Else == nil {
		mt.Unlock()
		return handlerErr
	}

	// Create new key and lock in immediately
	entry = &mapEntry{}
	entry.Lock()
	defer entry.Unlock()

	mt.rows[key] = entry
	mt.Unlock()

	// Invoke handler
	return entry.HandleWrite(handler.Else)
}

func (mt *mapTable) Delete(key string) {
	mt.Lock()
	defer mt.Unlock()

	delete(mt.rows, key)
}
