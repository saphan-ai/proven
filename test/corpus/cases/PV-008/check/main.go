// Command check runs every input of PV-008 against base, change and the
// reference, and prints one line per input: name, base, change, reference.
// Every wait is bounded, so the check always ends.
package main

import (
	"fmt"
	"strings"
	"time"

	"pv008/base"
	"pv008/change"
	"pv008/reference"
)

const (
	waitLimit = 1 * time.Second        // for each observation
	stopLimit = 100 * time.Millisecond // the reference's bound on stop
)

// worker is the stand-in. Stop returns at once, or never returns until release.
type worker struct {
	wedged  bool
	entered chan struct{}
	release chan struct{}
}

func newWorker(wedged bool) *worker {
	return &worker{wedged: wedged, entered: make(chan struct{}), release: make(chan struct{})}
}

func (w *worker) Stop() {
	close(w.entered)
	if w.wedged {
		<-w.release
	}
}

// session is what the three versions have in common.
type session interface {
	Subscribe() <-chan string
	Close()
	State() string
}

type impl struct {
	name string
	make func(w *worker) session
}

var impls = []impl{
	{"base", func(w *worker) session { return base.New(w) }},
	{"change", func(w *worker) session { return change.New(w) }},
	{"reference", func(w *worker) session { return reference.New(w, stopLimit) }},
}

// closeAsync starts Close and returns a channel that closes when it returns.
func closeAsync(s session) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		s.Close()
		close(done)
	}()
	return done
}

func waitFor(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	case <-time.After(waitLimit):
		return false
	}
}

// observe bounds calls that may block before they return an observation.
func observe[T any](call func() T) (T, bool) {
	result := make(chan T, 1)
	go func() { result <- call() }()
	select {
	case value := <-result:
		return value, true
	case <-time.After(waitLimit):
		var zero T
		return zero, false
	}
}

func lateFrame(s session) string {
	ch, ok := observe(s.Subscribe)
	if !ok {
		return "subscribe:blocked"
	}
	return frame(ch)
}

func frame(ch <-chan string) string {
	select {
	case f := <-ch:
		return "frame:" + f
	case <-time.After(waitLimit):
		return "no-frame"
	}
}

// start closes a session over a fresh worker and waits until the stop has been
// entered (and, for a normal worker, until Close has returned).
func start(i impl, wedged bool) (session, *worker, string) {
	w := newWorker(wedged)
	s := i.make(w)
	done := closeAsync(s)
	if !waitFor(w.entered) {
		return s, w, "stop-not-entered"
	}
	if !wedged && !waitFor(done) {
		return s, w, "close-not-returned"
	}
	return s, w, ""
}

var inputs = []struct {
	name string
	run  func(i impl) string
}{
	{"normal-close-then-late-subscriber", func(i impl) string {
		s, w, failure := start(i, false)
		defer close(w.release)
		if failure != "" {
			return failure
		}
		return lateFrame(s)
	}},
	{"wedged-stop-then-late-subscriber", func(i impl) string {
		s, w, failure := start(i, true)
		defer close(w.release)
		if failure != "" {
			return failure
		}
		return lateFrame(s)
	}},
	{"wedged-stop-then-read-state", func(i impl) string {
		s, w, failure := start(i, true)
		defer close(w.release)
		if failure != "" {
			return failure
		}
		state, ok := observe(s.State)
		if !ok {
			return "state:blocked"
		}
		return "state:" + state
	}},
	{"wedged-stop-then-second-close", func(i impl) string {
		s, w, failure := start(i, true)
		defer close(w.release)
		if failure != "" {
			return failure
		}
		if waitFor(closeAsync(s)) {
			return "returned"
		}
		return "blocked"
	}},
}

func main() {
	for _, in := range inputs {
		res := make([]string, len(impls))
		for k, i := range impls {
			res[k] = in.run(i)
		}
		fmt.Printf("%s | %s\n", in.name, strings.Join(res, " | "))
	}
}
