package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/trace"
)

// Property: a review whose trace file fails to close returns an error that
// names the trace, instead of reporting success from a record that may be
// incomplete (E22: every error handled or named).
func TestReviewFailsWhenTheTraceDoesNotClose(t *testing.T) {
	t.Parallel()
	fx := loadDiscount(t)
	gen, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{Sample: true, LogDir: t.TempDir(), Generator: gen})
	if err != nil {
		t.Fatal(err)
	}
	diskFull := errors.New("no space left on device")
	p.closeTrace = func(w *trace.Writer) error {
		if err := w.Close(); err != nil {
			return err
		}
		return diskFull
	}
	_, err = p.Review(context.Background(), Request{Target: fx.Root})
	if !errors.Is(err, diskFull) {
		t.Fatalf("Review returned %v, want the close error", err)
	}
	if !strings.Contains(err.Error(), "closing the trace") {
		t.Fatalf("the error does not name the trace: %v", err)
	}
}
