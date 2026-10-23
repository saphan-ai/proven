package main

import (
	"sync"
	"testing"
)

// This negative control keeps the lock while Stop is wedged. All observations
// must return a named timeout instead of hanging the check.
type heldLockSession struct {
	mu     sync.Mutex
	w      *worker
	closed bool
}

func (s *heldLockSession) Subscribe() <-chan string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return make(chan string)
}

func (s *heldLockSession) State() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return "closed"
}

func (s *heldLockSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.w.Stop()
}

func TestHeldLockObservationsAreBounded(t *testing.T) {
	i := impl{"held-lock-control", func(w *worker) session { return &heldLockSession{w: w} }}
	s, w, failure := start(i, true)
	defer close(w.release)
	if failure != "" {
		t.Fatalf("start: %s", failure)
	}
	if got := lateFrame(s); got != "subscribe:blocked" {
		t.Errorf("subscribe = %q", got)
	}
	if _, ok := observe(s.State); ok {
		t.Error("state returned through the held lock")
	}
	if waitFor(closeAsync(s)) {
		t.Error("second close returned through the held lock")
	}
}

type noStopSession struct{ release <-chan struct{} }

func (s *noStopSession) Subscribe() <-chan string { return make(chan string) }
func (s *noStopSession) State() string            { return "closed" }
func (s *noStopSession) Close()                   { <-s.release }

func TestStopNotEnteredIsNamed(t *testing.T) {
	i := impl{"no-stop-control", func(w *worker) session { return &noStopSession{release: w.release} }}
	_, w, failure := start(i, true)
	defer close(w.release)
	if failure != "stop-not-entered" {
		t.Fatalf("start = %q, want stop-not-entered", failure)
	}
}
