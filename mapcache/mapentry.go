package mapcache

import (
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

func (me *mapEntry) Read(p []byte) (int, error) {
	if me.readOffset >= len(me.buf) {
		return 0, io.EOF
	}

	n := copy(p, me.buf[me.readOffset:])
	me.readOffset += n
	return n, nil
}

func (me *mapEntry) Write(p []byte) (int, error) {
	me.modified = time.Now()

	me.readOffset = me.writeOffset
	me.buf = append(me.buf[:me.writeOffset], p...)
	me.writeOffset += len(p)

	return len(p), nil
}

func (me *mapEntry) Reset() {
	me.readOffset = 0
	me.writeOffset = 0
}
