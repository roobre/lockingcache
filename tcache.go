// tcache Defines implementation of a write-once, transactional cache specially suited for caching
// data which takes more time to be produced than to be re-requested.
// Its transactional design ensures that subsequent accesses to the same key blocks until the key is filled
// by an ongoing transaction, making data available and then be served from cache.
package tcache

import (
	"io"
	"time"
)

// Cache is an object capable of serving different cache tables given a key
type Cache interface {
	// From returns a cache table given a key. It will be created if it does not exist
	From(key string) Table
}

// Table is the main tcache interface, supporting concurrent read-or-write operation on entries for a given key
type Table interface {
	// Access accesses a cache entry given its key. If found, and its age is less or equal than age,
	// handler.Then will be invoked with an io.Reader containing the serialized entry as a parameter.
	// If the entry is not found, handler.Else will be invoked instead with an io.Writer parameter, allowing
	// the caller to fill the entry concurrently
	Access(key string, maxAge time.Duration, handler Handler) error

	// Delete removes an entry from the cache given its key
	Delete(key string)
}

// Handler is sugar syntax for holding the Then and Else functions
type Handler struct {
	// Then is executed when the key is found in the cache, which is made available as an io.Reader
	Then func(io.Reader) error
	// Else is executed when the key is missing. An io.Writer will be provided
	Else func(io.Writer) error
}
