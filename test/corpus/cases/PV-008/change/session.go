// Package change is the session under review: one routine composes the final
// frame, stores it, and replays it to late subscribers.
package change

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
	final  string
	done   chan struct{} // closed once the final frame is stored and published
}

// New returns an open session over w.
func New(w Worker) *Session {
	return &Session{state: "open", worker: w, done: make(chan struct{})}
}

// Subscribe returns a channel of frames. When the final frame already exists,
// it is replayed once.
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

// composeFinal is the only composer of the final frame.
func (s *Session) composeFinal() string {
	return "ended"
}

// Close sets the state to closed under the lock and releases it, stops the
// worker with no time limit, then composes, stores and publishes the final
// frame. A later call waits for the first call to finish.
func (s *Session) Close() {
	s.mu.Lock()
	if s.state == "closed" {
		s.mu.Unlock()
		<-s.done
		return
	}
	s.state = "closed"
	s.mu.Unlock()

	s.worker.Stop()

	frame := s.composeFinal()
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
