// Package reference models the intended behaviour. Stopping the worker is
// bounded by a time limit, and the final frame is published either way, with an
// alarm in the frame when the worker has not released. Used only by the check.
package reference

import (
	"sync"
	"time"
)

// Worker is the long-running process the session relays to.
type Worker interface {
	// Stop ends the worker. It may take any time.
	Stop()
}

// Session relays frames from one worker to its viewers.
type Session struct {
	mu        sync.Mutex
	state     string
	subs      []chan string
	worker    Worker
	stopLimit time.Duration
	final     string
	done      chan struct{}
}

// New returns an open session over w whose stop is bounded by stopLimit.
func New(w Worker, stopLimit time.Duration) *Session {
	return &Session{state: "open", worker: w, stopLimit: stopLimit, done: make(chan struct{})}
}

// Subscribe returns a channel of frames. A stored final frame is replayed once.
func (s *Session) Subscribe() <-chan string {
	ch := make(chan string, 4)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.final != "" {
		ch <- s.final
		return ch
	}
	s.subs = append(s.subs, ch)
	return ch
}

// Close marks the session closed, stops the worker for at most the limit, and
// publishes the final frame either way.
func (s *Session) Close() {
	s.mu.Lock()
	if s.state == "closed" {
		s.mu.Unlock()
		<-s.done
		return
	}
	s.state = "closed"
	s.mu.Unlock()

	stopped := make(chan struct{})
	go func() {
		s.worker.Stop()
		close(stopped)
	}()
	frame := "ended"
	select {
	case <-stopped:
	case <-time.After(s.stopLimit):
		frame = "ended+alarm:worker-not-released"
	}

	s.mu.Lock()
	s.final = frame
	subs := s.subs
	s.subs = nil
	s.mu.Unlock()
	for _, ch := range subs {
		ch <- frame
	}
	close(s.done)
}

// State reads the session's state.
func (s *Session) State() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
