package review

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

// TestLiveGitReviewReachesThePlanner proves a live review of a scratch git
// repository with two commits and no fixture.json reaches the Planner and
// returns a result, and that the planner prompt contains the file that
// changed between the commits. Only tracked files are in that prompt.
func TestLiveGitReviewReachesThePlanner(t *testing.T) {
	repo, base, head := twoCommits(t)
	if err := os.WriteFile(filepath.Join(repo, "secret.txt"), []byte("not-tracked-secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	gen := &recordingGen{}
	res := reviewLive(t, work, gen, Request{
		Target: repo, BaselineCommit: base, HeadCommit: head,
	})
	if res.Partial {
		t.Fatal("partial")
	}
	prompt := plannerPrompt(t, gen)
	const want = "FILE note.txt\nBASE\nbase-text\nCHANGE\ncommitted-change\n"
	if !strings.Contains(prompt, want) {
		t.Fatalf("planner prompt missing the changed file\n%s", prompt)
	}
	if strings.Contains(prompt, "not-tracked-secret") {
		t.Fatal("an untracked file was copied into the trees")
	}
	if !traceSnapshot(t, res.TracePath, base+".."+head) {
		t.Fatalf("snapshot %s not in the trace", base+".."+head)
	}
}

// TestLiveWorktreeReviewSeesAnUncommittedEdit proves that an empty HeadCommit
// reads the working tree, so an edit that was not committed is in the planner
// prompt.
func TestLiveWorktreeReviewSeesAnUncommittedEdit(t *testing.T) {
	repo, _, head := twoCommits(t)
	if err := os.WriteFile(filepath.Join(repo, "note.txt"), []byte("uncommitted-edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "secret.txt"), []byte("not-tracked-secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gen := &recordingGen{}
	reviewLive(t, t.TempDir(), gen, Request{
		Target: repo, BaselineCommit: head, HeadCommit: "",
	})
	prompt := plannerPrompt(t, gen)
	const want = "FILE note.txt\nBASE\ncommitted-change\nCHANGE\nuncommitted-edit\n"
	if !strings.Contains(prompt, want) {
		t.Fatalf("planner prompt missing the uncommitted edit\n%s", prompt)
	}
	if strings.Contains(prompt, "not-tracked-secret") {
		t.Fatal("an untracked file was copied into the trees")
	}
}

// TestLiveGitTreesAreRemovedAfterTheReview proves the temporary trees exist
// while the Planner runs and are gone when the review returns. A review that
// never wrote them would fail the during-run check.
func TestLiveGitTreesAreRemovedAfterTheReview(t *testing.T) {
	repo, base, head := twoCommits(t)
	work := t.TempDir()
	gen := &recordingGen{work: work}
	reviewLive(t, work, gen, Request{
		Target: repo, BaselineCommit: base, HeadCommit: head,
	})
	if !gen.sawTrees {
		t.Fatal("the temporary trees were not present while the Planner ran")
	}
	entries, err := os.ReadDir(work)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "logs" && e.Name() != "state" {
			t.Fatalf("left behind: %s", e.Name())
		}
	}
}

// TestSampleModeWithoutFixtureJSONStillFails proves sample mode still refuses
// a git repository that has no fixture.json. Mutated field: the target, which
// has two commits and no fixture.json. The baseline is set so the file list
// can be read and the failure is the fixture load.
func TestSampleModeWithoutFixtureJSONStillFails(t *testing.T) {
	repo, base, head := twoCommits(t)
	rev, err := Open(Options{Sample: true, LogDir: t.TempDir(), Env: func(string) string { return "" }})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rev.Review(context.Background(), Request{
		Target: repo, BaselineCommit: base, HeadCommit: head,
	})
	if err == nil || !strings.Contains(err.Error(), "fixture.json") {
		t.Fatalf("sample mode error: %v", err)
	}
}

// TestLiveGitWithoutBaselineIsRefused proves a git repository with no
// baseline commit and no base/ and change/ is refused with a sentence that
// says what to pass.
func TestLiveGitWithoutBaselineIsRefused(t *testing.T) {
	repo, _, _ := twoCommits(t)
	gen := &recordingGen{}
	rev := openLive(t, t.TempDir(), gen)
	_, err := rev.Review(context.Background(), Request{Target: repo})
	if err == nil || !strings.Contains(err.Error(), "baseline commit") || !strings.Contains(err.Error(), "base/ and change/") {
		t.Fatalf("refusal: %v", err)
	}
	if len(gen.prompts) != 0 {
		t.Fatal("the Planner ran")
	}
}

// TestLivePreparedTreesNeedNoFixtureJSON proves a live review of a directory
// that already has base/ and change/ does not read fixture.json, and records
// the snapshot name "live".
func TestLivePreparedTreesNeedNoFixtureJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "change"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "base", "hello.txt"), []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "change", "hello.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gen := &recordingGen{}
	res := reviewLive(t, t.TempDir(), gen, Request{Target: dir})
	prompt := plannerPrompt(t, gen)
	const want = "FILE hello.txt\nBASE\nold\nCHANGE\nnew\n"
	if !strings.Contains(prompt, want) {
		t.Fatalf("planner prompt\n%s", prompt)
	}
	if !traceSnapshot(t, res.TracePath, "live") {
		t.Fatal("snapshot live was not written")
	}
}

// recordingGen is the stub generator. It records planner prompts and, when
// work is set, notices the temporary trees while the Planner is running.
// marker and linkName, when set, record whether a symlink or its target
// content was present in those trees at that moment.
type recordingGen struct {
	work        string
	prompts     []string
	sawTrees    bool
	marker      string
	linkName    string
	markerSeen  bool
	linkPresent bool
}

// Complete returns an empty planner list, so the review finishes with no
// finding and without calling a network.
func (g *recordingGen) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		return "{}", nil
	}
	g.prompts = append(g.prompts, prompt)
	if g.work != "" {
		g.sawTrees = treesPresent(g.work)
	}
	if g.marker != "" && g.work != "" {
		g.markerSeen, g.linkPresent = treesHold(g.work, g.linkName, g.marker)
	}
	return "[]", nil
}

// quietQuery is the stub query adapter. It returns no rows and dials nothing.
type quietQuery struct{}

// Query returns no rows.
func (quietQuery) Query(context.Context, string) (query.Result, error) {
	return query.Result{}, nil
}

// quietRun is the stub test runner. It runs no container.
type quietRun struct{}

// Run runs nothing.
func (quietRun) Run(context.Context, testrun.Request) (testrun.Outcome, error) {
	return testrun.Outcome{}, nil
}

// reviewLive runs one live review and fails the test on an error.
func reviewLive(t *testing.T, work string, gen *recordingGen, req Request) Result {
	t.Helper()
	rev := openLive(t, filepath.Join(work, "logs"), gen)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := rev.Review(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// openLive opens a live reviewer with the stub generator, query and runner.
func openLive(t *testing.T, logDir string, gen *recordingGen) Reviewer {
	t.Helper()
	rev, err := Open(Options{
		Sample:    false,
		LogDir:    logDir,
		Env:       liveTestEnv(),
		gen:       gen,
		liveQuery: quietQuery{},
		testRun:   quietRun{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

// liveTestEnv is enough for Open to accept live mode. The stubs replace the
// clients that would dial, so these values are never sent.
func liveTestEnv() func(string) string {
	m := map[string]string{
		"METERLESS_RUNTIME_TOKEN":   "test-token",
		"PROVEN_GENERATOR_BASE_URL": "http://127.0.0.1:9",
		"PROVEN_GENERATOR_TIMEOUT":  "2s",
		"PROVEN_GUARDRAIL_MODE":     "off",
		"ZETARIS_REST_URL":          "http://127.0.0.1:9",
		"ZETARIS_API_KEY":           "zetaris-test-key",
		"ZETARIS_ORG_ID":            "1",
		"ZETARIS_TIMEOUT":           "1s",
	}
	return func(k string) string { return m[k] }
}

// plannerPrompt is the one planner prompt the review sent.
func plannerPrompt(t *testing.T, gen *recordingGen) string {
	t.Helper()
	if len(gen.prompts) != 1 {
		t.Fatalf("planner calls %d", len(gen.prompts))
	}
	return gen.prompts[0]
}

// treesPresent reports whether a temporary tree under work holds note.txt in
// both sides and does not hold a git directory.
func treesPresent(work string) bool {
	entries, err := os.ReadDir(work)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		root := filepath.Join(work, e.Name())
		if _, err := os.Stat(filepath.Join(root, "base", "note.txt")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "change", "note.txt")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "base", ".git")); err == nil {
			return false
		}
		if _, err := os.Stat(filepath.Join(root, "change", ".git")); err == nil {
			return false
		}
		return true
	}
	return false
}

// traceSnapshot reports whether the trace has a snapshot event with text.
func traceSnapshot(t *testing.T, path, text string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var ev struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if json.Unmarshal(line, &ev) != nil {
			continue
		}
		if ev.Kind == "snapshot" && ev.Text == text {
			return true
		}
	}
	return false
}

// TestLiveGitReviewSkipsATrackedSymlink proves a live review of two commits
// still finishes when the second commit adds a tracked symlink beside a
// changed regular file. The regular file reaches the Planner. The link is
// named once on the trace. The file the link points at, outside the
// repository, is not read: its content is not in the planner prompt, the
// trace, or the result, and the link is not in the trees while the Planner runs.
func TestLiveGitReviewSkipsATrackedSymlink(t *testing.T) {
	repo, base, head, marker := commitsWithSymlink(t)
	work := t.TempDir()
	gen := &recordingGen{work: work, marker: marker, linkName: "linked.txt"}
	res := reviewLive(t, work, gen, Request{
		Target: repo, BaselineCommit: base, HeadCommit: head,
	})
	if res.Partial {
		t.Fatal("partial")
	}
	prompt := plannerPrompt(t, gen)
	const want = "FILE note.txt\nBASE\nbase-text\nCHANGE\ncommitted-change\n"
	if !strings.Contains(prompt, want) {
		t.Fatalf("planner prompt missing the changed file\n%s", prompt)
	}
	assertSymlinkSkipped(t, res, gen, prompt, marker, "linked.txt")
}

// TestLiveWorktreeReviewSkipsATrackedSymlink proves the same property when
// the change side is the working tree rather than a second commit. The
// baseline commit already tracks the symlink, and an uncommitted edit of the
// regular file still reaches the Planner.
func TestLiveWorktreeReviewSkipsATrackedSymlink(t *testing.T) {
	repo, _, head, marker := commitsWithSymlink(t)
	if err := os.WriteFile(filepath.Join(repo, "note.txt"), []byte("uncommitted-edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	gen := &recordingGen{work: work, marker: marker, linkName: "linked.txt"}
	res := reviewLive(t, work, gen, Request{
		Target: repo, BaselineCommit: head, HeadCommit: "",
	})
	prompt := plannerPrompt(t, gen)
	const want = "FILE note.txt\nBASE\ncommitted-change\nCHANGE\nuncommitted-edit\n"
	if !strings.Contains(prompt, want) {
		t.Fatalf("planner prompt missing the uncommitted edit\n%s", prompt)
	}
	assertSymlinkSkipped(t, res, gen, prompt, marker, "linked.txt")
}

// assertSymlinkSkipped checks the link was named once and its target was not read.
func assertSymlinkSkipped(t *testing.T, res Result, gen *recordingGen, prompt, marker, rel string) {
	t.Helper()
	if strings.Contains(prompt, marker) || strings.Contains(prompt, "FILE "+rel+"\n") {
		t.Fatalf("planner prompt contains the symlink or its target\n%s", prompt)
	}
	if gen.linkPresent {
		t.Fatal("the symlink was in a tree while the Planner ran")
	}
	if gen.markerSeen {
		t.Fatal("the symlink target was read into a tree")
	}
	raw, err := os.ReadFile(res.TracePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(marker)) {
		t.Fatal("the symlink target was written into the trace")
	}
	if n := countSkipped(t, raw, rel); n != 1 {
		t.Fatalf("skipped events for %s: %d", rel, n)
	}
	encoded, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(marker)) {
		t.Fatal("the symlink target was written into the result")
	}
	if len(res.SkippedPaths) != 1 || res.SkippedPaths[0] != rel {
		t.Fatalf("skipped paths: %#v", res.SkippedPaths)
	}
}

// countSkipped counts trace lines that name rel as a skipped symlink.
func countSkipped(t *testing.T, raw []byte, rel string) int {
	t.Helper()
	n := 0
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var ev struct {
			Kind string `json:"kind"`
			Path string `json:"path"`
			Text string `json:"text"`
		}
		if json.Unmarshal(line, &ev) != nil {
			continue
		}
		if ev.Kind == "skipped" && ev.Path == rel && ev.Text == "symlink" {
			n++
		}
	}
	return n
}

// treesHold reports whether any tree under work holds marker in a regular
// file, or holds an entry named linkName. A symlink is not followed.
func treesHold(work, linkName, marker string) (markerSeen, linkPresent bool) {
	_ = filepath.Walk(work, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if linkName != "" && filepath.Base(path) == linkName {
				linkPresent = true
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if linkName != "" && filepath.Base(path) == linkName {
			linkPresent = true
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if bytes.Contains(b, []byte(marker)) {
			markerSeen = true
		}
		return nil
	})
	return markerSeen, linkPresent
}

// commitsWithSymlink builds a scratch repository with two commits. The second
// changes note.txt and adds linked.txt, a tracked symlink to a file outside
// the repository. marker is that file's content.
func commitsWithSymlink(t *testing.T) (dir, base, head, marker string) {
	t.Helper()
	dir = t.TempDir()
	gitCmd(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("base-text\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "note.txt")
	gitCmd(t, dir, "commit", "-m", "[fixture] note: base")
	base = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("committed-change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	marker = "symlink-target-secret-not-in-repo"
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "linked.txt")); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "note.txt", "linked.txt")
	gitCmd(t, dir, "commit", "-m", "[fixture] note: change and link")
	head = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	return dir, base, head, marker
}

// twoCommits builds a scratch git repository with two commits of note.txt and
// returns the directory, the first commit and the second commit.
func twoCommits(t *testing.T) (dir, base, head string) {
	t.Helper()
	dir = t.TempDir()
	gitCmd(t, dir, "init")
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("base-text\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "note.txt")
	gitCmd(t, dir, "commit", "-m", "[fixture] note: base")
	base = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("committed-change\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "note.txt")
	gitCmd(t, dir, "commit", "-m", "[fixture] note: change")
	head = strings.TrimSpace(string(gitCmd(t, dir, "rev-parse", "HEAD")))
	return dir, base, head
}

// gitCmd runs git in dir with a fixed identity, so the test does not depend
// on the machine's git settings.
func gitCmd(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	full := append([]string{
		"-c", "user.name=Proven Test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
		"-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}
