package mapcache

import (
	"bytes"
	"io"
	"sync"
	"time"
)

type mapEntry struct {
	sync.RWMutex
	valid    bool
	modified time.Time

	buf         []byte
	readOffset  int
	writeOffset int
}

func (me *mapEntry) Reader() io.Reader {
	return bytes.NewReader(me.buf)
}

func (me *mapEntry) Write(p []byte) (int, error) {
	me.modified = time.Now()

	me.buf = append(me.buf[:me.writeOffset], p...)
	me.writeOffset += len(p)

	return len(p), nil
}

func (me *mapEntry) Reset() {
	me.writeOffset = 0
}
