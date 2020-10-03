package mapcache

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"roob.re/tcache"
	"testing"
	"time"
)

func logTimestamp(t *testing.T, msg string, args ...interface{}) {
	now := time.Now()
	t.Logf("[%02d:%02d:%02d] %s", now.Hour(), now.Minute(), now.Second(), fmt.Sprintf(msg, args...))
}

type tableTester struct {
	rchan chan result
	table tcache.Table
	t     *testing.T
}

type result struct {
	key        string
	hit        bool
	err        error
	thenCalled bool
	elseCalled bool
}

var rerr = errors.New("rerr")
var werr = errors.New("werr")

func (tt *tableTester) request(key string, latency time.Duration, readError, writeError error) {
	rt := result{
		key: key,
	}

	reqId := fmt.Sprintf("%02d", rand.Int()%100)

	logTimestamp(tt.t, "> %s Requesting %s...", reqId, key)
	rt.err = tt.table.Access(key, 0, tcache.Handler{
		Then: func(r io.Reader) error {
			rt.thenCalled = true
			rt.hit = true
			logTimestamp(tt.t, "  %s %s found!", reqId, key)
			return readError
		},
		Else: func(w io.Writer) error {
			rt.elseCalled = true
			logTimestamp(tt.t, "  %s %s not found, waiting...", reqId, key)
			time.Sleep(latency)
			logTimestamp(tt.t, "  %s %s written", reqId, key)
			return writeError
		},
	})

	logTimestamp(tt.t, "< %s %s completed (%v, %v)", reqId, key, rt.hit, rt.err)

	tt.rchan <- rt
}

func TestParallelism(t *testing.T) {
	tester := tableTester{
		rchan: make(chan result),
		table: &mapTable{
			rows: map[string]*mapEntry{},
		},
		t: t,
	}

	go tester.request("first", 0, nil, nil)
	<-tester.rchan

	t.Log("Cache preheated")

	iters := 0

	go tester.request("second", 3*time.Second, nil, nil)
	iters++
	go tester.request("second", 3*time.Second, nil, nil)
	iters++
	go tester.request("second", 3*time.Second, nil, nil)
	iters++
	go tester.request("second", 3*time.Second, nil, nil)
	iters++
	go tester.request("second", 3*time.Second, nil, nil)
	iters++
	go tester.request("first", 5*time.Second, nil, nil)
	iters++
	go tester.request("first", 5*time.Second, nil, nil)
	iters++

	secondFound := 0
	for i := 0; i < iters; i++ {
		r := <-tester.rchan

		switch r.key {
		case "first":
			if !r.hit {
				t.Fatal("request to first failed")
			}

			if i > 2 {
				t.Fatalf("got response from first on iteration %d, it <= 2 expected", i)
			}

		case "second":
			if r.hit {
				secondFound++
			}
		}
	}

	if secondFound != 4 {
		t.Fatalf("expected 2 hits for second, got %d", secondFound)
	}
}

func TestSequentialInvalidation(t *testing.T) {
	tester := tableTester{
		rchan: make(chan result),
		table: &mapTable{
			rows: map[string]*mapEntry{},
		},
		t: t,
	}

	go tester.request("first", 0, nil, nil)
	<-tester.rchan

	go tester.request("first", 0, rerr, nil)
	r := <-tester.rchan
	if !r.elseCalled {
		log.Fatal("Else handler was not called after erroring read on hit")
	}

	go tester.request("first", 0, nil, nil)
	r = <-tester.rchan
	if !r.hit {
		t.Fatal("entry not re-populated after error")
	}

	go tester.request("first", 0, rerr, werr)
	<-tester.rchan

	go tester.request("first", 0, nil, nil)
	r = <-tester.rchan
	if r.hit {
		t.Fatal("entry was not invalidated by an erroring write")
	}

	go tester.request("first", 0, nil, nil)
	r = <-tester.rchan
	if !r.hit {
		t.Fatal("entry not re-populated after error")
	}
}

func TestParallelWriteErrors(t *testing.T) {
	tester := tableTester{
		rchan: make(chan result),
		table: &mapTable{
			rows: map[string]*mapEntry{},
		},
		t: t,
	}

	go tester.request("first", 4*time.Second, nil, werr)
	go tester.request("first", 4*time.Second, nil, werr)
	go tester.request("first", 4*time.Second, nil, werr)

	thensCalled := 0
	elsesCalled := 0

	for i := 0; i < 3; i++ {
		r := <-tester.rchan

		if r.thenCalled {
			thensCalled++
		}

		if r.elseCalled {
			elsesCalled++
		}
	}

	if thensCalled != 0 || elsesCalled != 3 {
		t.Fatalf("expected 0 thens and 3 elses called, got %d and %d", thensCalled, elsesCalled)
	}
}

func TestParallelReadErrors(t *testing.T) {
	tester := tableTester{
		rchan: make(chan result),
		table: &mapTable{
			rows: map[string]*mapEntry{},
		},
		t: t,
	}

	go tester.request("first", 0, nil, nil)
	<-tester.rchan

	go tester.request("first", 4*time.Second, rerr, nil)
	go tester.request("first", 4*time.Second, rerr, nil)
	go tester.request("first", 4*time.Second, rerr, nil)

	thensCalled := 0
	elsesCalled := 0

	for i := 0; i < 3; i++ {
		r := <-tester.rchan

		if r.thenCalled {
			thensCalled++
		}

		if r.elseCalled {
			elsesCalled++
		}
	}

	if thensCalled != 3 || elsesCalled != 3 {
		t.Fatalf("expected 0 thens and 3 elses called, got %d and %d", thensCalled, elsesCalled)
	}
}
