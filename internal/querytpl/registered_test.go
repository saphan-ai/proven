package querytpl_test

import (
	"errors"
	"regexp"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/source"
)

// concatAdapter is a registered source whose builder places its parameter in
// the query text without checking it. The format it declares is the one the
// test gives, so the only other guard is the one Proven owns.
type concatAdapter struct{ match *regexp.Regexp }

func (concatAdapter) ID() string { return "concat-probe" }
func (concatAdapter) Load() source.Load {
	return source.Load{Origin: "test", Refresh: "never", Table: "t"}
}
func (a concatAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   "concat-by-name",
			Answer: "a row",
			Params: []source.Param{{Name: "n", Format: "anything", Match: a.match}},
		},
		Build: func(p map[string]string, _ source.Rel) (string, error) {
			return "SELECT a FROM t WHERE n = '" + p["n"] + "'", nil
		},
	}}
}
func (concatAdapter) Rule(in source.Evidence) (source.Evidence, bool) { return in, len(in.Rows) > 0 }

// TestRegisteredTemplateValueIsHeldToItsFormatByProven proves that Proven, not
// the adapter, refuses a parameter value that holds a quote, a NUL, a line
// break, a backslash or a byte above 0x7e, through both Build and Validate,
// even when the adapter's format admits every text and its builder
// concatenates. The mutated field is the parameter value; the format is held
// permissive. Negative control: deleting the byte floor in source.CheckValue
// (the backslash and the bytes above 0x7e) turns the new cases red. The quote,
// NUL and newline cases are the ones that were already here.
func TestRegisteredTemplateValueIsHeldToItsFormatByProven(t *testing.T) {
	if err := source.Register(concatAdapter{match: regexp.MustCompile(`(?s)^.*$`)}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister("concat-probe") })

	for name, value := range map[string]string{
		"quote":      "x' OR '1'='1",
		"nul":        "x\x00y",
		"newline":    "x\ny",
		"backslash":  `a\b`,
		"above 0x7e": "a\x7fb",
	} {
		req := querytpl.Request{Template: "concat-by-name", Params: map[string]string{"n": value}}
		for verb, run := range map[string]func() error{
			"Build":    func() error { _, err := querytpl.Build(req, querytpl.Relations{}); return err },
			"Validate": func() error { return querytpl.Validate(req) },
		} {
			var ref *querytpl.Refusal
			if err := run(); !errors.As(err, &ref) || ref.Reason != querytpl.ReasonBadValue {
				t.Errorf("%s %s: err %v, want ReasonBadValue", verb, name, err)
			}
		}
	}
	good := querytpl.Request{Template: "concat-by-name", Params: map[string]string{"n": "left-pad"}}
	if sql, err := querytpl.Build(good, querytpl.Relations{}); err != nil || sql != "SELECT a FROM t WHERE n = 'left-pad'" {
		t.Fatalf("a plain value: %q %v", sql, err)
	}
}

// TestRegisteredTemplateFormatMismatchIsRefusedByProven proves a value outside
// the format the adapter declared is refused before the builder runs. The
// mutated field is the parameter value; the format admits digits only.
func TestRegisteredTemplateFormatMismatchIsRefusedByProven(t *testing.T) {
	if err := source.Register(concatAdapter{match: regexp.MustCompile(`^[0-9]+$`)}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister("concat-probe") })
	err := querytpl.Validate(querytpl.Request{Template: "concat-by-name", Params: map[string]string{"n": "abc"}})
	var ref *querytpl.Refusal
	if !errors.As(err, &ref) || ref.Reason != querytpl.ReasonBadValue {
		t.Fatalf("err %v, want ReasonBadValue", err)
	}
}

// TestRegisterRefusesAParameterWithNoFormat proves a template whose parameter
// declares no machine format is not admitted to the registry. The mutated
// field is Param.Match.
func TestRegisterRefusesAParameterWithNoFormat(t *testing.T) {
	if err := source.Register(concatAdapter{match: nil}); err == nil {
		source.Unregister("concat-probe")
		t.Fatal("Register admitted a parameter with no format")
	}
}
