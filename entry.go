package tcache

import (
	"errors"
	"io"
	"sync"
	"time"
)

type entry struct {
	sync.RWMutex
	accessor Accessor
	valid    bool
	modified time.Time
}

func (e *entry) read(maxAge time.Duration, handler func(io.Reader) error) error {
	// Check for validity and age
	if e.valid && (maxAge == 0 || time.Since(e.modified) < maxAge) {
		// If valid, unlock index and process it

		// Do nothing if we dont have a found handler
		if handler == nil {
			return nil
		}

		r, err := e.accessor.Reader()
		if err != nil {
			return errors.New("accessor backend returned an error: " + err.Error())
		}
		return handler(r)
	}

	return EntryInvalidError
}

func (e *entry) write(handler func(writer io.Writer) error) error {
	w, err := e.accessor.Writer()
	if err != nil {
		return errors.New("accessor backend returned an error: " + err.Error())
	}

	err = handler(w)
	e.modified = time.Now()
	if err == nil {
		// Mark accessor as valid if else handler did not error
		e.valid = true
	}

	return err
}
