package tcache

import (
	"bytes"
	"io"
)

// BufferEntry implements Accessor using bytes.Buffer as backend
type BufferEntry struct {
	buffer bytes.Buffer
}

func (be *BufferEntry) Reader() (io.Reader, error) {
	return bytes.NewReader(be.buffer.Bytes()), nil
}

func (be *BufferEntry) Writer() (io.Writer, error) {
	return &be.buffer, nil
}
