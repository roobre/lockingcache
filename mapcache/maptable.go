package mapcache

import (
	lc "roob.re/tcache"
	"sync"
	"time"
)

type mapTable struct {
	sync.Mutex
	rows map[string]*mapEntry
}

func (mt *mapTable) Access(key string, maxAge time.Duration, handler lc.Handler) error {
	// noop
	if handler.Then == nil && handler.Else == nil {
		return nil
	}

	mt.Lock()

	entry, entryFound := mt.rows[key]
	if entryFound {
		// We found an entry, attempt to lock it
		entry.RLock()
		defer entry.RUnlock()

		// Check for validity and age
		if entry.valid && time.Since(entry.modified) < maxAge {
			// If valid, unlock index and process it
			mt.Unlock()

			// Do nothing if we dont have a found handler
			if handler.Then == nil {
				return nil
			}

			err := handler.Then(entry)
			entry.Reset()
			return err
		} else {
			// If expired or not valid, delete it from the map
			delete(mt.rows, key)
		}
	}

	// Exit early if we do not have an else handler
	if handler.Else == nil {
		mt.Unlock()
		return nil
	}

	// Create new key and lock in immediately
	entry = &mapEntry{}
	mt.rows[key] = entry
	entry.Lock()
	defer entry.Unlock()
	// Release index mutex
	// As entry mutex is locked for writing, subsequent accesses will block before validity check
	mt.Unlock()

	// Invoke handler
	err := handler.Else(entry)
	entry.Reset()
	if err == nil {
		// Mark entry as valid if else handler did not error
		entry.valid = true
	}

	return err
}

func (mt *mapTable) Delete(key string) {
	mt.Lock()
	defer mt.Unlock()

	delete(mt.rows, key)
}
