package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/modelcall"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/trace"
)

type failedObserved struct{}

func (failedObserved) Complete(context.Context, string, string) (string, error) {
	return "", &service.Down{Service: "model", Reason: "timeout"}
}
func (failedObserved) LastAttempts() []modelcall.Attempt {
	return []modelcall.Attempt{{Error: "timeout", ReasoningBytes: 23}}
}

func TestFailedHTTPAttemptIsRecordedWithoutInventingModelIdentity(t *testing.T) {
	w, err := trace.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	var models []candidate.ModelUse
	g := recording{inner: failedObserved{}, trace: w, models: &models}
	if _, err := g.Complete(context.Background(), "planner", "prompt"); err == nil {
		t.Fatal("failure lost")
	}
	events := readTrace(t, w.Path())
	if len(models) != 0 || len(events) != 1 || events[0].Kind != "model-attempt" || events[0].Model.Attempt.Error != "timeout" || events[0].Model.Attempt.ReasoningBytes != 23 {
		t.Fatalf("models=%#v events=%#v", models, events)
	}
}

// receipted names a fixed route, model and receipt id for every call.
// receipted answers like the sample generator and reports a fixed receipt for
// every call.
type receipted struct {
	inner generator.Generator
}

func (g *receipted) Complete(ctx context.Context, agent, prompt string) (string, error) {
	return g.inner.Complete(ctx, agent, prompt)
}

func (g *receipted) LastUse() (generator.Use, bool) {
	return generator.Use{Route: "nvidia", Model: "nvidia/test-model", ReceiptID: "rcpt-0001", FromReceipt: true}, true
}

// Every model call writes one "model" event to the trace carrying the route,
// the model and the receipt id, so a reader can match the trace to the
// provider's receipts.
func TestEveryModelCallWritesItsReceiptIDToTheTrace(t *testing.T) {
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{Sample: true, LogDir: t.TempDir(), Generator: &receipted{inner: inner}})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(d.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	calls := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		var e trace.Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatal(err)
		}
		if e.Kind != "model" {
			continue
		}
		calls++
		if e.Model == nil || e.Model.ReceiptID != "rcpt-0001" || e.Model.Model != "nvidia/test-model" || !e.Model.FromReceipt || e.Agent == "" {
			t.Fatalf("model event without its receipt: %#v", e)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if calls == 0 || calls != len(d.Models) {
		t.Fatalf("model events %d, model calls %d", calls, len(d.Models))
	}
}
