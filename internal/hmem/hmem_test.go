package hmem_test

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

func TestAddRefusesABlankSource(t *testing.T) {
	s := openMem(t)
	if _, err := s.Add(hmem.AddInput{Content: "a fact", Source: "  "}); err == nil {
		t.Fatal("blank source was stored")
	}
}

func TestLoadRefusesARecordWithNoSource(t *testing.T) {
	dir := t.TempDir()
	body := []byte(`{"version":1,"seq":1,"records":[{"id":"m1","content":"x","type":"factual","layer":"working","confidence":1,"provenance":{"origin":"manual","learnedAt":1,"label":"x"}}]}` + "\n")
	if err := os.WriteFile(filepath.Join(dir, "state.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()}); err == nil {
		t.Fatal("loaded a record with no source")
	}
}

func TestLoadRefusesABadVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":9,"seq":0,"records":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()}); err == nil {
		t.Fatal("loaded version 9")
	}
}

func TestStateInsideTheRepositoryIsRefused(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "state")
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir(), RepoRoot: root})
	if err == nil {
		t.Fatal("state directory inside the repository was accepted")
	}
}

func TestScoreMatchesTheFormula(t *testing.T) {
	s := openMem(t)
	conf := 1.0
	rec, err := s.Add(hmem.AddInput{
		Content: "alpha beta", Source: "test", Confidence: &conf, At: fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := hmem.Score(rec, "alpha beta", fixed)
	if math.Abs(got-0.74) > 1e-9 {
		t.Fatalf("score %v, want 0.74", got)
	}
	rec.SupersededBy = "mem-other"
	lower := hmem.Score(rec, "alpha beta", fixed)
	if lower >= got {
		t.Fatalf("superseded %v is not below %v", lower, got)
	}
}

func TestMarkSupersededLowersTheRank(t *testing.T) {
	s := openMem(t)
	conf := 0.9
	fresh, err := s.Add(hmem.AddInput{
		Content: "we deploy the api with docker on mondays", Source: "conformance",
		Confidence: &conf, Layer: hmem.LayerLong, At: fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.Add(hmem.AddInput{
		Content: "we deploy the api with docker on fridays", Source: "conformance",
		Confidence: &conf, Layer: hmem.LayerLong, At: fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkSuperseded(old.ID, fresh.ID); err != nil {
		t.Fatal(err)
	}
	q, err := s.Query("when do we deploy the api with docker", 10, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Hits) < 2 || q.Hits[0].Record.ID != fresh.ID {
		t.Fatalf("hits %#v", q.Hits)
	}
	got, _ := s.Get(old.ID)
	if got.SupersededBy != fresh.ID {
		t.Fatalf("supersededBy %q", got.SupersededBy)
	}
}

func TestQueryDropsBelowTheThreshold(t *testing.T) {
	s := openMem(t)
	zero := 0.0
	if _, err := s.Add(hmem.AddInput{Content: "alpha beta", Source: "test", Confidence: &zero, At: fixed}); err != nil {
		t.Fatal(err)
	}
	q, err := s.Query("alpha beta", 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Hits) != 0 {
		t.Fatalf("hits %#v", q.Hits)
	}
}

func TestFeedbackDeltas(t *testing.T) {
	s := openMem(t)
	start := 0.5
	rec, err := s.Add(hmem.AddInput{Content: "feedback target record", Source: "test", Confidence: &start, At: fixed})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Feedback(rec.ID, "helpful"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(rec.ID)
	if math.Abs(got.Confidence-0.55) > 1e-9 || got.AccessCount != 1 {
		t.Fatalf("helpful %#v", got)
	}
	if err := s.Feedback(rec.ID, "not_helpful"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get(rec.ID)
	if math.Abs(got.Confidence-0.52) > 1e-9 {
		t.Fatalf("not_helpful %v", got.Confidence)
	}
	if err := s.Feedback(rec.ID, "wrong"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Get(rec.ID)
	if math.Abs(got.Confidence-0.32) > 1e-9 || !hasTag(got.Tags, "review") {
		t.Fatalf("wrong %#v", got)
	}
}

func TestRememberIsIdempotentAndSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	logs := t.TempDir()
	s := openAt(t, dir, logs)
	key := hmem.RefutationKey{
		ChangeKey: "chg", RuleID: "discount-applied-twice", Path: "billing/discount.go",
		Title: "Discount is applied twice", Reason: "the helper applies it once", At: fixed,
	}
	a, err := s.RememberRefutation(key)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.RememberRefutation(key)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID || len(s.All()) != 1 {
		t.Fatalf("id %s %s count %d", a.ID, b.ID, len(s.All()))
	}
	again := openAt(t, dir, logs)
	got, ok := again.MatchRefutation(key.ChangeKey, key.RuleID, key.Path, key.Title)
	if !ok || got.ID != a.ID {
		t.Fatalf("match %#v %v", got, ok)
	}
	if _, ok := again.MatchRefutation("other", key.RuleID, key.Path, key.Title); ok {
		t.Fatal("a different change matched")
	}
}

func TestClearDropsTheMatch(t *testing.T) {
	s := openMem(t)
	key := hmem.RefutationKey{ChangeKey: "chg", RuleID: "r", Path: "p.go", Title: "t", Reason: "because", At: fixed}
	if _, err := s.RememberRefutation(key); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.MatchRefutation(key.ChangeKey, key.RuleID, key.Path, key.Title); ok {
		t.Fatal("cleared store still matched")
	}
}

func TestAuditIsAppendOnly(t *testing.T) {
	dir := t.TempDir()
	logs := t.TempDir()
	s := openAt(t, dir, logs)
	a, err := s.Add(hmem.AddInput{Content: "one fact here", Source: "test", At: fixed})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(hmem.AddInput{Content: "two facts here", Source: "test", At: fixed}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(logs, "hmem.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), a.ID) || strings.Count(string(b), "\n") != 2 {
		t.Fatalf("%s", b)
	}
}

func TestGeneratorIsNotCalled(t *testing.T) {
	s := openMem(t)
	g := &boom{}
	s.SetGenerator(g)
	if _, err := s.Add(hmem.AddInput{Content: "alpha beta", Source: "test", At: fixed}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Query("alpha beta", 5, 0); err != nil {
		t.Fatal(err)
	}
	if g.called {
		t.Fatal("the store called the generator")
	}
}

type boom struct{ called bool }

func (b *boom) Complete(context.Context, string, string) (string, error) {
	b.called = true
	return "", nil
}

const fixed = int64(1_700_000_000_000)

func openMem(t *testing.T) *hmem.Store {
	t.Helper()
	return openAt(t, t.TempDir(), t.TempDir())
}

func openAt(t *testing.T, dir, logs string) *hmem.Store {
	t.Helper()
	s, err := hmem.Open(hmem.Options{
		Dir: dir, LogDir: logs, Clock: func() time.Time { return time.UnixMilli(fixed) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}
