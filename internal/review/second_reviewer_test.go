package review

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/redact"
)

// secondTestKey is the second reviewer's key in these tests. It is never a real one.
const secondTestKey = "sk-second-reviewer-test-key-7731"

// secondModel is the model name the second reviewer is configured with in these tests.
const secondModel = "gpt-second"

// secondAPI is the second reviewer's API for a test: the transport of the real second-reviewer client,
// so no port is opened. It answers the Planner and the Skeptic from recorded answers, keeps every
// prompt it was sent, and fails every call with status when status is set.
type secondAPI struct {
	planner string
	skeptic string
	status  int
	prompts []string
}

// RoundTrip records the prompt and answers it as one streamed event.
func (a *secondAPI) RoundTrip(r *http.Request) (*http.Response, error) {
	raw, _ := io.ReadAll(r.Body)
	var req struct {
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(raw, &req); err != nil || len(req.Messages) != 1 {
		return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	}
	prompt := req.Messages[0].Content
	a.prompts = append(a.prompts, prompt)
	if a.status != 0 {
		return &http.Response{StatusCode: a.status, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	}
	// The Skeptic's prompt names a risk on an "id: " line; the Planner's does not.
	answer := a.planner
	if promptID(prompt) != "" {
		answer = a.skeptic
	}
	event, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": answer}}}})
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: " + string(event) + "\n\ndata: [DONE]\n\n")),
		Request:    r,
	}, nil
}

// reviewWithSecond runs one review with the first reviewer def and the real second-reviewer client
// over api, as Open wires them when the three settings are present. The redactor holds the key.
func reviewWithSecond(t *testing.T, cfg fileSettings, def generator.Generator, api *secondAPI) Result {
	t.Helper()
	second := generator.NewSecond(generator.SecondSettings{
		BaseURL: "https://api.example.test/v1", Key: secondTestKey, Model: secondModel, Timeout: 5e9,
	}, &http.Client{Transport: api})
	env := func(k string) string {
		if k == generator.EnvSecondKey {
			return secondTestKey
		}
		return ""
	}
	p, err := pipeline.Build(pipeline.Options{
		Sample: true, LogDir: t.TempDir(), Generator: def, Special: second, Redactor: redact.FromEnv(env),
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := &runner{p: p, cfg: cfg, explicit: true, secondModel: secondModel}
	res, err := r.Review(context.Background(), Request{Target: discountFixture(t)})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// verdictOf returns the verdict a reviewer of role gave on a finding.
func verdictOf(f Finding, role string) (ReviewerVerdict, bool) {
	for _, v := range f.Verdicts {
		if v.Role == role {
			return v, true
		}
	}
	return ReviewerVerdict{}, false
}

// TestTwoReviewersThatDisagreeShowBothVerdicts proves ADR-0037 point 4: where the two reviewers
// differ on a finding, the result carries each reviewer's own verdict and reason side by side, flags
// the disagreement and counts it, and averages nothing. Three kinds of difference are present: the
// second reviewer's own Skeptic refutes a risk both named (B1), only the first named a risk (D1),
// and only the second named one (S1). A risk with no evidence is still dropped (N1): evidence decides.
func TestTwoReviewersThatDisagreeShowBothVerdicts(t *testing.T) {
	def := newScript("NVIDIA", risks(
		risk("B1", "seen by both"), risk("D1", "first only"), risk("N1", "no evidence"),
	), `{"refuted":false,"reason":"first-keeps-it"}`)
	def.sql = map[string]string{"B1": rowAnswer, "D1": rowAnswer, "S1": rowAnswer}
	api := &secondAPI{
		planner: risks(risk("B1", "seen by both"), risk("S1", "second only")),
		skeptic: `{"refuted":true,"reason":"second-refutes-it"}`,
	}
	res := reviewWithSecond(t, defaultSettings(), def, api)

	b1, ok := finding(res, "B1")
	if !ok {
		t.Fatal("B1 missing")
	}
	first, _ := verdictOf(b1, RoleDefault)
	second, _ := verdictOf(b1, RoleSpecial)
	if first.Verdict != VerdictReported || first.Model != "NVIDIA" {
		t.Fatalf("B1 first %+v", first)
	}
	if second.Verdict != VerdictRefuted || second.Reason != "second-refutes-it" || second.Model != secondModel {
		t.Fatalf("B1 second %+v", second)
	}
	if !b1.Disagree || len(b1.Verdicts) != 2 || b1.Verdicts[0].Role != RoleDefault {
		t.Fatalf("B1 disagree %v verdicts %+v", b1.Disagree, b1.Verdicts)
	}
	d1, _ := finding(res, "D1")
	if v, _ := verdictOf(d1, RoleSpecial); v.Verdict != VerdictNotReported || !d1.Disagree {
		t.Fatalf("D1 %+v", d1.Verdicts)
	}
	s1, _ := finding(res, "S1")
	if v, _ := verdictOf(s1, RoleDefault); v.Verdict != VerdictNotReported || !s1.Disagree || s1.Verdicts[0].Role != RoleDefault {
		t.Fatalf("S1 %+v", s1.Verdicts)
	}
	if _, ok := finding(res, "N1"); ok {
		t.Fatal("N1 has no evidence and was reported")
	}
	if res.Disagreements != 3 {
		t.Fatalf("disagreements %d", res.Disagreements)
	}
}

// TestReviewersThatAgreeShowNoDisagreement proves that two reviewers who both name a risk and neither
// refutes it are two matching verdicts and no disagreement, so the shell has no block to show.
func TestReviewersThatAgreeShowNoDisagreement(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("B1", "seen by both")), `{"refuted":false,"reason":"first-keeps-it"}`)
	def.sql = map[string]string{"B1": rowAnswer}
	api := &secondAPI{planner: risks(risk("B1", "seen by both")), skeptic: `{"refuted":false,"reason":"second-keeps-it"}`}
	res := reviewWithSecond(t, defaultSettings(), def, api)
	b1, ok := finding(res, "B1")
	if !ok || b1.Disagree || len(b1.Verdicts) != 2 || res.Disagreements != 0 {
		t.Fatalf("B1 %+v disagreements %d", b1, res.Disagreements)
	}
	for _, v := range b1.Verdicts {
		if v.Verdict != VerdictReported {
			t.Fatalf("verdict %+v", v)
		}
	}
}

// TestOneReviewerAloneIsUnchanged proves that with no second reviewer connected a finding carries no
// verdicts, no disagreement is counted, and the review names one reviewer: the review is as it was.
// Mutated field: whether the second reviewer is connected (it is not).
func TestOneReviewerAloneIsUnchanged(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("D1", "first only")), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"D1": rowAnswer}
	res := reviewScripted(t, defaultSettings(), def, newScript("unused", risks(), `{}`))
	d1, ok := finding(res, "D1")
	if !ok || len(d1.Verdicts) != 0 || d1.Disagree || res.Disagreements != 0 {
		t.Fatalf("D1 %+v disagreements %d", d1, res.Disagreements)
	}
	if len(res.Reviewers) != 1 || res.Reviewers[0].Role != RoleDefault || res.Reviewers[0].Model != "NVIDIA" {
		t.Fatalf("reviewers %+v", res.Reviewers)
	}
}

// TestTheResultNamesBothReviewersAndTheAuthorsModel proves ADR-0037 point 5: each reviewer is named
// with its family and whether it is the author's own model, and "unknown" stays unknown when the change
// names no author.
func TestTheResultNamesBothReviewersAndTheAuthorsModel(t *testing.T) {
	def := newScript("NVIDIA", risks(), `{}`)
	res := reviewWithSecond(t, defaultSettings(), def, &secondAPI{planner: risks(), skeptic: `{}`})
	if len(res.Reviewers) != 2 || res.Reviewers[1].Role != RoleSpecial || res.Reviewers[1].Model != secondModel {
		t.Fatalf("reviewers %+v", res.Reviewers)
	}
	for _, r := range res.Reviewers {
		if r.IsAuthorModel != "unknown" {
			t.Fatalf("no author was named, so the answer is unknown: %+v", r)
		}
	}
	got := reviewerOf(RoleSpecial, "gpt-5", defaultSettings().Families, []AuthorModel{{Model: "gpt-4o", Family: "OpenAI"}}, nil)
	if got.IsAuthorModel != "yes" || got.Family != "OpenAI" {
		t.Fatalf("same family as the author: %+v", got)
	}
	got = reviewerOf(RoleSpecial, "gpt-5", defaultSettings().Families, []AuthorModel{{Model: "claude-x", Family: "Anthropic"}}, nil)
	if got.IsAuthorModel != "no" {
		t.Fatalf("another family: %+v", got)
	}
	got = reviewerOf(RoleSpecial, "mystery", defaultSettings().Families, []AuthorModel{{Model: "claude-x", Family: "Anthropic"}}, nil)
	if got.IsAuthorModel != "unknown" {
		t.Fatalf("a family unknown stays unknown: %+v", got)
	}
}

// TestAFailedSecondReviewerIsNotADisagreement proves that when the second reviewer's service fails, its
// verdict is "no answer", the failure is named under the second reviewer, and nothing is flagged as a
// disagreement: a missing answer is not a different answer.
// Mutated field: the second reviewer's HTTP status.
func TestAFailedSecondReviewerIsNotADisagreement(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("D1", "first only")), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"D1": rowAnswer}
	res := reviewWithSecond(t, defaultSettings(), def, &secondAPI{status: http.StatusInternalServerError})
	d1, ok := finding(res, "D1")
	if !ok {
		t.Fatal("D1 missing")
	}
	if v, _ := verdictOf(d1, RoleSpecial); v.Verdict != VerdictNoAnswer || d1.Disagree || res.Disagreements != 0 {
		t.Fatalf("D1 %+v disagreements %d", d1.Verdicts, res.Disagreements)
	}
	named := false
	for _, f := range res.FailedServices {
		named = named || f.Service == generator.SecondService
	}
	if !named {
		t.Fatalf("failed services %+v", res.FailedServices)
	}
}

// TestASecondReviewerWidensTheDoubledClasses proves ADR-0037 point 3: a connected second reviewer
// doubles every class unless proven.yaml names the classes, security is doubled either way, and with
// none connected the settings are returned untouched.
// Mutated field: the doubled classes and whether the file named them.
func TestASecondReviewerWidensTheDoubledClasses(t *testing.T) {
	plain := defaultSettings()
	if got := withSecond(plain, ""); got.SpecialReviewer != "" || len(got.Doubled) != 1 || !got.Doubled["security"] {
		t.Fatalf("not connected: %+v", got)
	}
	all := withSecond(plain, secondModel)
	for _, class := range classNames() {
		if !all.Doubled[class] {
			t.Errorf("class %s is not doubled", class)
		}
	}
	if all.SpecialReviewer != secondModel {
		t.Fatalf("special reviewer %q", all.SpecialReviewer)
	}
	narrow := defaultSettings()
	narrow.DoubledSet, narrow.Doubled = true, map[string]bool{"docs": true}
	got := withSecond(narrow, secondModel)
	if !got.Doubled["docs"] || !got.Doubled["security"] || got.Doubled["backend"] {
		t.Fatalf("narrowed: %+v", got.Doubled)
	}
	off := defaultSettings()
	off.DoubledSet, off.Doubled = true, map[string]bool{}
	if got := withSecond(off, secondModel); !got.Doubled["security"] || got.Doubled["docs"] {
		t.Fatalf("off: %+v", got.Doubled)
	}
	if plain.Doubled["docs"] {
		t.Fatal("withSecond changed the settings it was given")
	}
}

// TestASecondReviewerReviewsAClassThatWasNotDoubledBefore proves the widening end to end: the discount
// change is in the backend class, which is not doubled by default, and with a second reviewer
// connected both reviewers run on it. When proven.yaml narrows the classes to docs, the second reviewer
// is not called.
func TestASecondReviewerReviewsAClassThatWasNotDoubledBefore(t *testing.T) {
	def := newScript("NVIDIA", risks(risk("D1", "first only")), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"D1": rowAnswer}
	api := &secondAPI{planner: risks(), skeptic: `{}`}
	cfg := defaultSettings()
	cfg.PathRules = []pathRule{{Pattern: "**/*.md", Class: "docs"}}
	res := reviewWithSecond(t, cfg, def, api)
	if res.Class != "backend" || !res.Doubled || len(api.prompts) == 0 {
		t.Fatalf("class %s doubled %v calls %d", res.Class, res.Doubled, len(api.prompts))
	}
	narrow := &secondAPI{planner: risks(), skeptic: `{}`}
	cfg.DoubledSet, cfg.Doubled = true, map[string]bool{"docs": true}
	res = reviewWithSecond(t, cfg, def, narrow)
	if res.Doubled || len(narrow.prompts) != 0 {
		t.Fatalf("doubled %v calls %d", res.Doubled, len(narrow.prompts))
	}
}

// TestSecondReviewerNeverSeesTheFirstsAnswer proves ADR-0037 point 3: neither reviewer sees the
// other's answer. The two name the same risk in different words. Every prompt the second reviewer
// receives must hold its own words and none of the first reviewer's: not its title, its message, its
// risk id, nor its Skeptic's reason. And the first reviewer's prompts hold none of the second's.
func TestSecondReviewerNeverSeesTheFirstsAnswer(t *testing.T) {
	first := risk("FIRST-ID-31", "FIRST-WORDING-4d1")
	first.RuleID = "shared-rule"
	first.Message = "FIRST-MESSAGE-77c (evil.example/leftpad 1.2.3)"
	second := risk("SECOND-ID-58", "SECOND-WORDING-9e2")
	second.RuleID = "shared-rule"
	second.Message = "SECOND-MESSAGE-12b (evil.example/leftpad 1.2.3)"
	def := newScript("NVIDIA", risks(first), `{"refuted":false,"reason":"FIRST-SKEPTIC-REASON-5c8"}`)
	def.sql = map[string]string{"FIRST-ID-31": rowAnswer}
	api := &secondAPI{planner: risks(second), skeptic: `{"refuted":false,"reason":"SECOND-SKEPTIC-REASON-6a0"}`}
	res := reviewWithSecond(t, defaultSettings(), def, api)

	merged, ok := finding(res, "FIRST-ID-31")
	if !ok || len(merged.Verdicts) != 2 {
		t.Fatalf("the two wordings of one risk are one finding with two verdicts: %+v", res.Findings)
	}
	if _, ok := finding(res, "SECOND-ID-58"); ok {
		t.Fatal("the second wording became a second finding")
	}
	firstWords := []string{"FIRST-ID-31", "FIRST-WORDING-4d1", "FIRST-MESSAGE-77c", "FIRST-SKEPTIC-REASON-5c8"}
	secondWords := []string{"SECOND-ID-58", "SECOND-WORDING-9e2", "SECOND-MESSAGE-12b", "SECOND-SKEPTIC-REASON-6a0"}
	sawOwn := false
	for _, prompt := range api.prompts {
		for _, w := range firstWords {
			if strings.Contains(prompt, w) {
				t.Fatalf("the second reviewer was shown %q from the first:\n%s", w, prompt)
			}
		}
		sawOwn = sawOwn || strings.Contains(prompt, "SECOND-WORDING-9e2")
	}
	if !sawOwn {
		t.Fatal("the second Skeptic was not shown the second reviewer's own wording")
	}
	for _, c := range def.calls {
		for _, w := range secondWords {
			if strings.Contains(c.prompt, w) {
				t.Fatalf("the first reviewer was shown %q from the second", w)
			}
		}
	}
}

// TestTheSecondReviewersKeyNeverAppearsInAnyOutput proves ADR-0037 point 2: a second reviewer whose
// service echoes its key in a risk's message and in a Skeptic's reason does not put the key into the
// result, the verdicts, or the trace file.
// Mutated field: the text the second reviewer's service returns.
func TestTheSecondReviewersKeyNeverAppearsInAnyOutput(t *testing.T) {
	echo := risk("S1", "second only")
	echo.Message = "token " + secondTestKey + " (evil.example/leftpad 1.2.3)"
	def := newScript("NVIDIA", risks(risk("B1", "seen by both")), `{"refuted":false,"reason":"no refutation"}`)
	def.sql = map[string]string{"B1": rowAnswer, "S1": rowAnswer}
	api := &secondAPI{planner: risks(risk("B1", "seen by both"), echo), skeptic: `{"refuted":true,"reason":"echo ` + secondTestKey + `"}`}
	res := reviewWithSecond(t, defaultSettings(), def, api)

	out, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), secondTestKey) {
		t.Fatal("the key is in the result")
	}
	if !strings.Contains(string(out), "[redacted: "+generator.EnvSecondKey+"]") {
		t.Fatal("the echoed key was dropped without its marker, so the test proves nothing")
	}
	trace, err := os.ReadFile(res.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(trace), secondTestKey) {
		t.Fatal("the key is in the trace")
	}
}
