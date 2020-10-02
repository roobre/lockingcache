package mapcache

import (
	"bytes"
	"testing"
)

var aaaa = []byte("AAAA")
var bbbb = []byte("BBBB")

func TestMapEntry(t *testing.T) {
	me := &mapEntry{}

	readbuf := make([]byte, 16)

	n, _ := me.Read(readbuf)
	if n > 0 {
		t.Fatalf("Initial read returned %d bytes", n)
	}

	me.Write(aaaa)

	n, _ = me.Read(readbuf)
	if n != len(aaaa) {
		t.Fatalf("Read after write returned %d, expected %d", n, len(aaaa))
	}

	if bytes.Compare(readbuf[:len(aaaa)], aaaa) != 0 {
		t.Fatalf("Read did not return the expected value")
	}

	n, _ = me.Read(readbuf)
	if n != 0 {
		t.Fatalf("Subsequent read returned %d bytes", n)
	}

	me.Write(append(aaaa, bbbb...))
	expected := append(aaaa, bbbb...)

	n, _ = me.Read(readbuf)
	if n != len(expected) {
		t.Fatalf("Read after write returned %d, expected %d", n, len(expected))
	}

	if bytes.Compare(readbuf[:len(expected)], expected) != 0 {
		t.Fatalf("Read did not return the expected value")
	}
}