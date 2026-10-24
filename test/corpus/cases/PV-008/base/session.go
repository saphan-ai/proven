// Package base is the session before the change: the final frame goes to the
// current subscribers only. A viewer that subscribes after the end gets nothing.
package base

import "sync"

// Worker is the long-running process the session relays to.
type Worker interface {
	// Stop ends the worker. It may take any time.
	Stop()
}

// Session relays frames from one worker to its viewers.
type Session struct {
	mu     sync.Mutex
	state  string
	subs   []chan string
	worker Worker
}

// New returns an open session over w.
func New(w Worker) *Session {
	return &Session{state: "open", worker: w}
}

// Subscribe returns a channel of frames. A subscription made after the end is
// never fed.
func (s *Session) Subscribe() <-chan string {
	ch := make(chan string, 4)
	s.mu.Lock()
	s.subs = append(s.subs, ch)
	s.mu.Unlock()
	return ch
}

// Close sets the state to closed, stops the worker, and publishes the final
// frame to the viewers present at that moment.
func (s *Session) Close() {
	s.mu.Lock()
	if s.state == "closed" {
		s.mu.Unlock()
		return
	}
	s.state = "closed"
	s.mu.Unlock()

	s.worker.Stop()

	s.mu.Lock()
	subs := append([]chan string(nil), s.subs...)
	s.mu.Unlock()
	for _, ch := range subs {
		ch <- "ended"
	}
}

// State reads the session's state.
func (s *Session) State() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}
