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

// asked reports a use that asked for one model and was answered by another,
// with a response id, so the trace has something to record.
type asked struct {
	inner generator.Generator
}

func (g *asked) Complete(ctx context.Context, agent, prompt string) (string, error) {
	return g.inner.Complete(ctx, agent, prompt)
}

func (g *asked) LastUse() (generator.Use, bool) {
	return generator.Use{
		Route: "nvidia", Model: "nvidia/answered", ReceiptID: "rcpt-0002", FromReceipt: true,
		Requested: "nvidia/asked", ResponseID: "chatcmpl-0002",
	}, true
}

// Property: every "model" event of the trace carries the requested model and
// the response id beside the routed model and the receipt id, so the audit can
// join the trace to the receipts file and compare asked with answered.
func TestTraceModelEventCarriesRequestedModelAndResponseID(t *testing.T) {
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{Sample: true, LogDir: t.TempDir(), Generator: &asked{inner: inner}})
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
		m := e.Model
		if m == nil || m.Requested != "nvidia/asked" || m.ResponseID != "chatcmpl-0002" || m.Model != "nvidia/answered" {
			t.Fatalf("model event lacks its custody fields: %#v", e.Model)
		}
	}
	if calls == 0 {
		t.Fatal("no model events")
	}
}
