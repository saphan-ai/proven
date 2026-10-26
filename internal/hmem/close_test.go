package hmem_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestStateCloseFailureIsReturned proves a close error is not dropped: a valid
// state whose file fails to close is refused, and the error names the file and
// wraps the close cause. It turns red when the close error is discarded.
func TestStateCloseFailureIsReturned(t *testing.T) {
	cause := errors.New("close failed")
	f := newCountedFile(100, 100)
	f.closeErr = cause
	s, err := hmem.OpenWithStateFile(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: 4096}}, f)
	if err == nil || s != nil {
		t.Fatalf("a valid state whose close failed was opened: store %v, error %v", s, err)
	}
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "close state") {
		t.Fatalf("the refusal does not name the close failure: %v", err)
	}
}

// TestStateReadErrorStaysPrimaryWhenCloseFails proves that when the read and
// the close both fail, the read error is the one returned.
func TestStateReadErrorStaysPrimaryWhenCloseFails(t *testing.T) {
	readCause := errors.New("read failed")
	f := newCountedFile(100, 100)
	f.readErr = readCause
	f.closeErr = errors.New("close failed")
	_, err := hmem.OpenWithStateFile(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: 4096}}, f)
	if !errors.Is(err, readCause) || errors.Is(err, f.closeErr) {
		t.Fatalf("the read error is not the primary error: %v", err)
	}
}

// TestStateOverTheCapStaysPrimaryWhenCloseFails proves the cap refusal is not
// replaced by a close error either.
func TestStateOverTheCapStaysPrimaryWhenCloseFails(t *testing.T) {
	f := newCountedFile(8192, 8192)
	f.closeErr = errors.New("close failed")
	_, err := hmem.OpenWithStateFile(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: 4096}}, f)
	if !errors.Is(err, hmem.ErrStateTooLarge) {
		t.Fatalf("the cap refusal is not the primary error: %v", err)
	}
}
