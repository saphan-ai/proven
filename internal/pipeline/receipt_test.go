package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/trace"
)

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
