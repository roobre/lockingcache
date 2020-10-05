// tcache Defines implementation of a write-once, transactional cache specially suited for caching
// data which takes more time to be produced than to be re-requested.
// Its transactional design ensures that subsequent accesses to the same key blocks until the key is filled
// by an ongoing transaction, making data available and then be served from cache.
package tcache

import (
	"errors"
	"io"
	"sync"
	"time"
)

// Cache is an object capable of serving different cache tables given a key
type Cache struct {
	storage Storage
	rows    map[string]*entry
	mtx     sync.Mutex
}

// New returns a new Cache backed by the specified storage
func New(storage Storage) *Cache {
	return &Cache{
		storage: storage,
		rows:    map[string]*entry{},
	}
}

// Access accesses a cache entry given its key. If found, and its age is less or equal than age,
// handler.Then will be invoked with an io.Reader containing the serialized entry as a parameter.
// If the entry is not found, handler.Else will be invoked instead with an io.Writer parameter, allowing
// the caller to fill the entry concurrently
func (c *Cache) Access(key string, maxAge time.Duration, handler Handler) error {
	c.mtx.Lock()
	ent, entryFound := c.rows[key]

	handlerErr := EntryMissingError
	if entryFound {
		c.mtx.Unlock()

		ent.RLock()
		handlerErr = ent.read(maxAge, handler.Then)
		ent.RUnlock()

		if handlerErr == nil {
			return nil
		}

		ent.Lock()
		ent.valid = false
		ent.Unlock()

		c.mtx.Lock()
		delete(c.rows, key)
	}

	// Exit early if we do not have an else handler
	if handler.Else == nil {
		c.mtx.Unlock()
		return handlerErr
	}

	// Create new key and lock in immediately
	ent = &entry{}
	ent.Lock()
	defer ent.Unlock()

	c.rows[key] = ent
	c.mtx.Unlock()

	ent.accessor = c.storage.Get(key)

	// Invoke handler
	return ent.write(handler.Else)
}

// Delete removes an entry from the cache given its key
func (c *Cache) Delete(key string) {
	c.mtx.Lock()
	delete(c.rows, key)
	c.mtx.Unlock()

	c.storage.Delete(key)
}

// Handler is sugar syntax for holding the Then and Else functions
type Handler struct {
	// Then is executed when the key is found in the cache, which is made available as an io.Reader
	Then func(r io.Reader) error
	// Else is executed when the key is missing. An io.Writer will be provided
	Else func(w io.Writer) error
}

// EntryMissingError is returned by Get if an accessor is not found, and no Else handler is provided
var EntryMissingError = errors.New("accessor missing")

// EntryInvalidError is returned by Get if an accessor is invalid (or gets invalidated by Then) and no Else handler is provided
var EntryInvalidError = errors.New("accessor invalidated")

// Storage is an object capable of keeping track of entries (through accessors)
type Storage interface {
	// Get must always return a valid non-nil Accessor, creating it if necessary
	Get(key string) Accessor
	// Delete must remove an accessor from the storage
	Delete(key string)
}

// Accessor is an object capable of providing read/write access to a cache entry
// For in-memory Storage implementations, an entry can be its own accessor, created by Storage.Get if it does not exist.
// If Reader or Writer return an error, the entry will be invalidated and the error will be echoed by Cache.Access
type Accessor interface {
	Reader() (io.Reader, error)
	Writer() (io.Writer, error)
}
