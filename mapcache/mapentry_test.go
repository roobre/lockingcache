package mapcache

import (
	"bytes"
	"testing"
)

var aaaa = []byte("AAAA")
var bbbb = []byte("BBBB")

func TestMapEntry(t *testing.T) {
	me := &mapEntry{}

	me.Write(aaaa)

	if len(me.buf) != len(aaaa) {
		t.Fatalf("Buffer has %d bytes, expected %d", len(me.buf), len(aaaa))
	}

	if bytes.Compare(me.buf[:len(aaaa)], aaaa) != 0 {
		t.Fatalf("Buffer does not contain the expected value")
	}

	me.Write(append(aaaa, bbbb...))
	expected := append(aaaa, append(aaaa, bbbb...)...)

	if len(me.buf) != len(expected) {
		t.Fatalf("Buffer has %d bytes, expected %d", len(me.buf), len(aaaa))
	}

	if bytes.Compare(me.buf[:len(expected)], expected) != 0 {
		t.Fatalf("Buffer does not contain the expected value")
	}
}
