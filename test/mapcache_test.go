package test

import (
	"roob.re/tcache"
	"testing"
)

func TestParallelism(t *testing.T) {
	testParallelism(t, tcache.New(tcache.NewMapStorage()))
}

func TestSequentialInvalidation(t *testing.T) {
	testSequentialInvalidation(t, tcache.New(tcache.NewMapStorage()))
}

func TestParallelWriteErrors(t *testing.T) {
	testParallelWriteErrors(t, tcache.New(tcache.NewMapStorage()))
}

func TestParallelReadErrors(t *testing.T) {
	testParallelReadErrors(t, tcache.New(tcache.NewMapStorage()))
}
