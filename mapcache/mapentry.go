package mapcache

import (
	"bytes"
	"io"
	"roob.re/tcache"
	"sync"
	"time"
)

type mapEntry struct {
	sync.RWMutex
	valid    bool
	modified time.Time

	buffer      bytes.Buffer
	readOffset  int
	writeOffset int
}

func (me *mapEntry) HandleRead(maxAge time.Duration, handler func(io.Reader) error) error {
	// Check for validity and age
	if me.valid && (maxAge == 0 || time.Since(me.modified) < maxAge) {
		// If valid, unlock index and process it

		// Do nothing if we dont have a found handler
		if handler == nil {
			return nil
		}

		return handler(me.Reader())
	}

	return tcache.EntryInvalidatedError
}

func (me *mapEntry) HandleWrite(handler func(writer io.Writer) error) error {
	err := handler(&me.buffer)
	me.modified = time.Now()
	if err == nil {
		// Mark entry as valid if else handler did not error
		me.valid = true
	}

	return err
}

func (me *mapEntry) Invalidate() {
	me.valid = false
}

func (me *mapEntry) Reader() io.Reader {
	return bytes.NewReader(me.buffer.Bytes())
}
