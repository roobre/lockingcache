package tcache

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestParallelism(t *testing.T) {
	tester := tester{
		rchan: make(chan result),
		cache: New(&nilStorage{}),
		t:     t,
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
	tester := tester{
		rchan: make(chan result),
		cache: New(&nilStorage{}),
		t:     t,
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
	tester := tester{
		rchan: make(chan result),
		cache: New(&nilStorage{}),
		t:     t,
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
	tester := tester{
		rchan: make(chan result),
		cache: New(&nilStorage{}),
		t:     t,
	}

	go tester.request("first", 0, nil, nil)
	<-tester.rchan

	go tester.request("first", 4*time.Second, rerr, nil)
	go tester.request("first", 4*time.Second, rerr, nil)
	go tester.request("first", 4*time.Second, rerr, nil)

	elsesCalled := 0

	for i := 0; i < 3; i++ {
		r := <-tester.rchan

		if r.elseCalled {
			elsesCalled++
		}
	}

	if elsesCalled != 3 {
		t.Fatalf("expected 3 elses called, got and %d", elsesCalled)
	}
}

func logTimestamp(t *testing.T, msg string, args ...interface{}) {
	now := time.Now()
	t.Logf("[%02d:%02d:%02d] %s", now.Hour(), now.Minute(), now.Second(), fmt.Sprintf(msg, args...))
}

type tester struct {
	rchan chan result
	cache *Cache
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

func (t *tester) request(key string, latency time.Duration, readError, writeError error) {
	rt := result{
		key: key,
	}

	reqId := fmt.Sprintf("%02d", rand.Int()%100)

	logTimestamp(t.t, "> %s Requesting %s...", reqId, key)
	rt.err = t.cache.Access(key, 0, Handler{
		Then: func(r io.Reader) error {
			rt.thenCalled = true
			rt.hit = true
			logTimestamp(t.t, "  %s %s found!", reqId, key)
			return readError
		},
		Else: func(w io.Writer) error {
			rt.elseCalled = true
			logTimestamp(t.t, "  %s %s not found, waiting...", reqId, key)
			time.Sleep(latency)
			logTimestamp(t.t, "  %s %s written", reqId, key)
			return writeError
		},
	})

	logTimestamp(t.t, "< %s %s completed (%v, %v)", reqId, key, rt.hit, rt.err)

	t.rchan <- rt
}

type nilStorage struct{}

func (*nilStorage) Get(key string) Accessor {
	return &nilAcessor{}
}

func (ns *nilStorage) Delete(key string) {}

type nilAcessor struct{}

func (*nilAcessor) Reader() (io.Reader, error) {
	return nil, nil
}

func (*nilAcessor) Writer() (io.Writer, error) {
	return nil, nil
}
