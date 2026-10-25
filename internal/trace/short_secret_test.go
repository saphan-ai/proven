package trace_test

import (
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/trace"
)

// TestTraceOmitsAConfiguredSecretOfAnyLength proves a configured secret reaches
// the trace in no length: the redactor has no length floor, so even a value the
// settings loaders would refuse is replaced if it is handed over.
func TestTraceOmitsAConfiguredSecretOfAnyLength(t *testing.T) {
	for _, secret := range []string{"a", "abcde", "abcdefg", "abcdefgh"} {
		w, err := trace.Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		w.SetRedactor(redact.New(map[string]string{"credential": secret}))
		if err := w.Append(trace.Event{Text: secret}); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(w.Path())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), `"text":"`+secret+`"`) {
			t.Errorf("a configured secret of %d bytes reached the trace", len(secret))
		}
	}
}
