package pipeline

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/redact"
)

// secretEnv is an environment holding the given variables.
func secretEnv(values map[string]string) func(string) string {
	return func(k string) string { return values[k] }
}

// echoing answers like the sample generator, and reports a route, a model and
// a receipt that carry a configured value, the way a service can echo what it
// was sent.
type echoing struct {
	inner generator.Generator
	model string
}

func (g *echoing) Complete(ctx context.Context, agent, prompt string) (string, error) {
	return g.inner.Complete(ctx, agent, prompt)
}

func (g *echoing) LastUse() (generator.Use, bool) {
	return generator.Use{Route: "route-" + g.model, Model: "openai/" + g.model, ReceiptID: "rcpt_" + g.model, FromReceipt: true}, true
}

// unsafeEcho is a guardrail that judges everything unsafe and names one
// string as the category.
type unsafeEcho struct{ category string }

func (u unsafeEcho) Check(context.Context, string, string) (guardrail.Verdict, error) {
	return guardrail.Verdict{User: "unsafe", Categories: []string{u.category}}, nil
}

// reviewEchoing runs one sample review with env, a generator that echoes
// model and the guardrail rail (nil means the sample guardrail), and returns
// the draft and the trace text.
func reviewEchoing(t *testing.T, env func(string) string, model string, rail guardrail.Adapter) (candidate.Draft, string) {
	t.Helper()
	fx := loadDiscount(t)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), Redactor: redact.FromEnv(env), Guardrail: rail,
		Generator: &echoing{inner: inner, model: model},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fx.Root})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(d.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	return d, string(b)
}

// TestConfiguredValueEchoedInAReceiptIsNotWritten proves that a configured
// secret or organisation id a service echoes in its route, model or receipt id
// reaches neither the trace nor the draft, and that the trace names the variable.
// Mutated field: the model, route and receipt id the generator reports.
func TestConfiguredValueEchoedInAReceiptIsNotWritten(t *testing.T) {
	for name, marker := range map[string]string{
		generator.EnvToken: "marker-token-7d41c0",
		guardrail.EnvKey:   "marker-guardkey-5e92ab",
		query.EnvAPIKey:    "marker-zetaris-key-30b8",
		query.EnvOrgID:     "918273645501",
	} {
		t.Run(name, func(t *testing.T) {
			d, trace := reviewEchoing(t, secretEnv(map[string]string{name: marker}), marker, nil)
			if strings.Contains(trace, marker) {
				t.Errorf("the trace carries the value:\n%s", trace)
			}
			if !strings.Contains(trace, "[redacted: "+name+"]") {
				t.Errorf("the trace does not name %s", name)
			}
			for _, m := range d.Models {
				if strings.Contains(m.Model+m.Route+m.ReceiptID, marker) {
					t.Errorf("the draft's model use carries the value: %#v", m)
				}
			}
		})
	}
}

// TestGuardrailKeyEchoedInTheCategoriesIsNotWritten proves that a guardrail
// key echoed in the safety categories reaches neither the trace nor the
// refusal reason of the draft.
// Mutated field: the categories of the guardrail's verdict.
func TestGuardrailKeyEchoedInTheCategoriesIsNotWritten(t *testing.T) {
	const marker = "marker-guardkey-5e92ab"
	env := secretEnv(map[string]string{guardrail.EnvKey: marker})
	d, trace := reviewEchoing(t, env, "plain-model", unsafeEcho{category: marker})
	if strings.Contains(trace, marker) {
		t.Errorf("the trace carries the key:\n%s", trace)
	}
	var seen bool
	for _, f := range d.Failed {
		if f.Service == guardrail.Service {
			seen = true
		}
		if strings.Contains(f.Reason, marker) {
			t.Errorf("a refusal reason carries the key: %#v", f)
		}
	}
	if !seen {
		t.Errorf("no guardrail refusal was recorded: %#v", d.Failed)
	}
}
