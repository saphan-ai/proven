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

// TestAddRefusesABlankSource proves a memory with a blank source is never stored.
func TestAddRefusesABlankSource(t *testing.T) {
	s := openMem(t)
	if _, err := s.Add(hmem.AddInput{Content: "a fact", Source: "  "}); err == nil {
		t.Fatal("blank source was stored")
	}
}

// TestLoadRefusesARecordWithNoSource proves Open refuses a state file that
// holds a record with no source.
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

// TestLoadRefusesABadVersion proves Open refuses a state file of an unknown version.
func TestLoadRefusesABadVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":9,"seq":0,"records":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()}); err == nil {
		t.Fatal("loaded version 9")
	}
}

// TestAStateFromBeforeEvidencedRefutationsIsRefused proves a version 1 state,
// whose refutations were not required to carry evidence, is refused by name
// instead of being read.
func TestAStateFromBeforeEvidencedRefutationsIsRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{"version":1,"seq":0,"records":[]}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "state version 1") || !strings.Contains(err.Error(), "remove") {
		t.Fatalf("a version 1 state was not refused by name: %v", err)
	}
}

// TestStateInsideTheRepositoryIsRefused proves memory cannot be kept inside
// the repository under review.
func TestStateInsideTheRepositoryIsRefused(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "state")
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir(), RepoRoot: root})
	if err == nil {
		t.Fatal("state directory inside the repository was accepted")
	}
}

// TestScoreMatchesTheFormula proves the hybrid score of a known record is the
// hand-computed 0.74, and a superseded record scores lower.
func TestScoreMatchesTheFormula(t *testing.T) {
	s := openMem(t)
	conf := 1.0
	rec, err := s.Add(hmem.AddInput{
		Content: "alpha beta", Source: "test", Confidence: &conf, At: fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := hmem.Score(rec, "alpha beta", fixed, 0)
	if math.Abs(got-0.74) > 1e-9 {
		t.Fatalf("score %v, want 0.74", got)
	}
	rec.SupersededBy = "mem-other"
	lower := hmem.Score(rec, "alpha beta", fixed, 0)
	if lower >= got {
		t.Fatalf("superseded %v is not below %v", lower, got)
	}
}

// TestMarkSupersededLowersTheRank proves a superseded memory ranks below the
// one that replaced it, and both are kept.
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

// TestQueryDropsBelowTheThreshold proves the default threshold drops a
// memory with zero confidence.
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

// TestFeedbackDeltas proves each feedback kind moves confidence by its fixed
// delta, helpful counts an access, and wrong tags the memory for review.
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

// TestRememberIsIdempotentAndSurvivesReopen proves the same refutation is
// stored once, is found again after a reopen, and does not match another change.
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

// TestClearDropsTheMatch proves Clear removes a stored refutation, so it no longer matches.
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

// TestAuditIsAppendOnly proves each write adds one audit line naming its record.
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

// TestGeneratorIsNotCalled proves Add and Query never call the injected
// model (ADR-0018: no model call inside the engine).
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

// boom is a generator that records whether it was called.
type boom struct{ called bool }

func (b *boom) Complete(context.Context, string, string) (string, error) {
	b.called = true
	return "", nil
}

// fixed is a fixed time in Unix milliseconds, so scores do not depend on the clock.
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

// TestFailedWriteLeavesMemoryAsItWas proves a failed audit or persist does not
// keep the new record: the call returns an error and the in-memory store
// matches the state file (F28).
func TestFailedWriteLeavesMemoryAsItWas(t *testing.T) {
	t.Run("add audit", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		blockAudit(t, logs)
		if _, err := s.Add(hmem.AddInput{Content: "should not stick", Source: "test", At: fixed}); err == nil {
			t.Fatal("add succeeded while the audit could not be written")
		}
		if len(s.All()) != 0 {
			t.Fatalf("memory kept %d records after a failed audit", len(s.All()))
		}
		if len(openAt(t, dir, logs).All()) != 0 {
			t.Fatal("the state file kept a record the call rejected")
		}
	})
	t.Run("add persist", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		blockPersist(t, dir)
		if _, err := s.Add(hmem.AddInput{Content: "should not stick", Source: "test", At: fixed}); err == nil {
			t.Fatal("add succeeded while the state file could not be written")
		}
		if len(s.All()) != 0 {
			t.Fatalf("memory kept %d records after a failed persist", len(s.All()))
		}
	})
	t.Run("refutation", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		blockAudit(t, logs)
		key := hmem.RefutationKey{ChangeKey: "chg", RuleID: "r", Path: "p.go", Title: "t", Reason: "because", At: fixed}
		if _, err := s.RememberRefutation(key); err == nil {
			t.Fatal("refutation succeeded while the audit could not be written")
		}
		if _, ok := s.MatchRefutation(key.ChangeKey, key.RuleID, key.Path, key.Title); ok {
			t.Fatal("memory kept a refutation after a failed audit")
		}
	})
	t.Run("feedback", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		rec, err := s.Add(hmem.AddInput{Content: "one fact here", Source: "test", At: fixed})
		if err != nil {
			t.Fatal(err)
		}
		blockAudit(t, logs)
		if err := s.Feedback(rec.ID, "helpful"); err == nil {
			t.Fatal("feedback succeeded while the audit could not be written")
		}
		got, _ := s.Get(rec.ID)
		if got.AccessCount != 0 || got.Confidence != 0.7 {
			t.Fatalf("memory changed after a failed feedback: count %d confidence %v", got.AccessCount, got.Confidence)
		}
	})
	t.Run("clear", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		if _, err := s.Add(hmem.AddInput{Content: "one fact here", Source: "test", At: fixed}); err != nil {
			t.Fatal(err)
		}
		blockPersist(t, dir)
		if err := s.Clear(); err == nil {
			t.Fatal("clear succeeded while the state file could not be written")
		}
		if len(s.All()) != 1 {
			t.Fatalf("memory has %d records after a failed clear, want the original one", len(s.All()))
		}
	})
	t.Run("supersede", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		fresh, err := s.Add(hmem.AddInput{Content: "the new fact about deploys", Source: "owner", At: fixed})
		if err != nil {
			t.Fatal(err)
		}
		old, err := s.Add(hmem.AddInput{Content: "the old fact about deploys", Source: "owner", At: fixed})
		if err != nil {
			t.Fatal(err)
		}
		blockPersist(t, dir)
		if err := s.MarkSuperseded(old.ID, fresh.ID); err == nil {
			t.Fatal("supersede succeeded while the state file could not be written")
		}
		gotOld, _ := s.Get(old.ID)
		gotNew, _ := s.Get(fresh.ID)
		if gotOld.SupersededBy != "" || gotNew.Supersedes != "" {
			t.Fatalf("old supersededBy %q new supersedes %q", gotOld.SupersededBy, gotNew.Supersedes)
		}
	})
	t.Run("query", func(t *testing.T) {
		dir, logs := t.TempDir(), t.TempDir()
		s := openAt(t, dir, logs)
		rec, err := s.Add(hmem.AddInput{Content: "alpha beta", Source: "test", At: fixed})
		if err != nil {
			t.Fatal(err)
		}
		blockPersist(t, dir)
		if _, err := s.Query("alpha beta", 5, 0); err == nil {
			t.Fatal("query succeeded while the state file could not be written")
		}
		got, _ := s.Get(rec.ID)
		if got.AccessCount != 0 {
			t.Fatalf("access count %d after a failed query", got.AccessCount)
		}
	})
}

// TestClearAuditsAndResetsSeq proves Clear writes an audit line and the next
// id after a reopen is mem-1 again (F28).
func TestClearAuditsAndResetsSeq(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openAt(t, dir, logs)
	if _, err := s.Add(hmem.AddInput{Content: "one fact here", Source: "test", At: fixed}); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(logs, "hmem.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"action":"clear"`) {
		t.Fatalf("audit has no clear line: %s", b)
	}
	next, err := openAt(t, dir, logs).Add(hmem.AddInput{Content: "after clear", Source: "test", At: fixed})
	if err != nil {
		t.Fatal(err)
	}
	if next.ID != "mem-1" {
		t.Fatalf("id %s, want mem-1 after clear reset the sequence", next.ID)
	}
}

// TestMarkSupersededSetsBothSides proves the replaced record names its
// replacement, the replacement names what it replaced, and the audit actor
// is that replacement rather than the retrieval role (F29).
func TestMarkSupersededSetsBothSides(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openAt(t, dir, logs)
	fresh, err := s.Add(hmem.AddInput{Content: "the new fact about deploys", Source: "owner", At: fixed})
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.Add(hmem.AddInput{Content: "the old fact about deploys", Source: "owner", At: fixed})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkSuperseded(old.ID, fresh.ID); err != nil {
		t.Fatal(err)
	}
	gotOld, _ := s.Get(old.ID)
	gotNew, _ := s.Get(fresh.ID)
	if gotOld.SupersededBy != fresh.ID {
		t.Fatalf("supersededBy %q", gotOld.SupersededBy)
	}
	if gotNew.Supersedes != old.ID {
		t.Fatalf("supersedes %q, want %s", gotNew.Supersedes, old.ID)
	}
	b, err := os.ReadFile(filepath.Join(logs, "hmem.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"actor":"`+fresh.ID+`"`) {
		t.Fatalf("audit does not name %s: %s", fresh.ID, b)
	}
	if strings.Contains(string(b), `"actor":"retrieval"`) {
		t.Fatalf("actor is still retrieval: %s", b)
	}
}

// blockAudit replaces hmem.jsonl with a directory, so the next audit open fails.
func blockAudit(t *testing.T, logs string) {
	t.Helper()
	path := filepath.Join(logs, "hmem.jsonl")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// blockPersist puts a directory where the state file's temporary name goes,
// so the next persist cannot write it.
func blockPersist(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, "state.json.tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestStateNamedWithTwoDotsInsideIsRefused proves a state directory inside the
// repository whose name starts with ".." (here "..state") is still refused:
// only a path that leaves the repository is outside it.
func TestStateNamedWithTwoDotsInsideIsRefused(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "..state")
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir(), RepoRoot: root})
	if err == nil || !strings.Contains(err.Error(), "inside the repository") {
		t.Fatalf("err %v, want the state directory refused as inside the repository", err)
	}
}
