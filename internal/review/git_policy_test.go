package review_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/review"
)

// policyOf is how a review reads the repository: its git policy and limits.
type policyOf interface {
	ReadPolicy() change.Policy
}

func openIn(t *testing.T, network string) review.Reviewer {
	t.Helper()
	env := func(k string) string {
		if k == "PROVEN_NETWORK" {
			return network
		}
		return ""
	}
	r, err := review.Open(review.Options{Sample: true, LogDir: t.TempDir(), Env: env})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func offlineArgs(r review.Reviewer, t *testing.T) bool {
	t.Helper()
	g, ok := r.(policyOf)
	if !ok {
		t.Fatalf("%T has no read policy a test can read", r)
	}
	return slices.Contains(g.ReadPolicy().GitCommand("status").Args, "protocol.allow=never")
}

// TestGitPolicyBelongsToTheReview: the git policy of a review is fixed when the review is opened. A local
// review keeps protocol.allow=never when a public review is opened after it, and a public review keeps
// none when a local one is opened after it, because the policy is the review's and not the process's.
// Mutated field: the policy the review holds (read the process-wide switch instead and the local review
// loses it when the public one opens; this test turns red).
func TestGitPolicyBelongsToTheReview(t *testing.T) {
	defer change.UseOffline(change.UseOffline(false))
	local := openIn(t, "local")
	public := openIn(t, "")
	if !offlineArgs(local, t) {
		t.Fatalf("the local review lost its offline git policy when a public review was opened")
	}
	if offlineArgs(public, t) {
		t.Fatalf("the public review took the offline git policy of the local review")
	}
	laterLocal := openIn(t, "local")
	if offlineArgs(public, t) {
		t.Fatalf("the public review took the offline git policy of a local review opened after it")
	}
	if !offlineArgs(laterLocal, t) || !offlineArgs(local, t) {
		t.Fatalf("a local review lost its offline git policy")
	}
}

// TestGitPolicyHoldsWhileReviewsAreOpenedTogether is the same property with reviews opened and read from
// many goroutines at once. Run with -race, it also shows the policy is never written after Open.
func TestGitPolicyHoldsWhileReviewsAreOpenedTogether(t *testing.T) {
	defer change.UseOffline(change.UseOffline(false))
	local := openIn(t, "local")
	public := openIn(t, "")
	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				network := ""
				if (i+j)%2 == 0 {
					network = "local"
				}
				env := func(k string) string {
					if k == "PROVEN_NETWORK" {
						return network
					}
					return ""
				}
				if _, err := review.Open(review.Options{Sample: true, LogDir: t.TempDir(), Env: env}); err != nil {
					errs <- err.Error()
					return
				}
				if !slices.Contains(local.(policyOf).ReadPolicy().GitCommand("status").Args, "protocol.allow=never") {
					errs <- "the local review lost its offline git policy"
					return
				}
				if slices.Contains(public.(policyOf).ReadPolicy().GitCommand("status").Args, "protocol.allow=never") {
					errs <- "the public review took an offline git policy"
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

// fakeGit puts a git first on PATH that records each call (its arguments, one file per call) in the
// returned directory and succeeds. The review's git has a minimal environment, so a call is attributed
// to a review by the -C directory it names.
func fakeGit(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	rec := t.TempDir()
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > '" + rec + "'/call-$$\n" +
		"case \" $* \" in *\" rev-parse \"*) echo 0123456789abcdef0123456789abcdef01234567;; esac\nexit 0\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return rec
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixture", "discount")
	dst := filepath.Join(t.TempDir(), name)
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

// callsFor returns, for each git call that named target with -C, whether it carried protocol.allow=never.
func callsFor(t *testing.T, rec, target string) (offline, online int) {
	t.Helper()
	entries, err := os.ReadDir(rec)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(rec, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		args := strings.Split(string(b), "\n")
		if !slices.Contains(args, target) {
			continue
		}
		if slices.Contains(args, "protocol.allow=never") {
			offline++
		} else {
			online++
		}
	}
	return offline, online
}

// TestEveryGitCommandOfAReviewCarriesItsOwnPolicy runs two reviews of different modes in one process,
// one after the other and then at the same time, and reads the arguments of every git command they
// started. Every command of the local review carries protocol.allow=never and none of the public one's
// does, so the pipeline's reads follow the review's policy too and not only its own helpers'.
// Mutated field: the policy the runner puts in the context and in the request.
func TestEveryGitCommandOfAReviewCarriesItsOwnPolicy(t *testing.T) {
	defer change.UseOffline(change.UseOffline(false))
	rec := fakeGit(t)
	localDir := copyFixture(t, "local")
	publicDir := copyFixture(t, "public")
	local := openIn(t, "local")
	public := openIn(t, "")
	run := func(r review.Reviewer, dir string) {
		// The fake git answers every call with success and no output; what the review makes of that
		// does not matter, only the commands it started.
		_, _ = r.Review(context.Background(), review.Request{Target: dir, BaselineCommit: "0123456789abcdef0123456789abcdef01234567", HeadCommit: "0123456789abcdef0123456789abcdef01234567"})
	}
	run(local, localDir)
	run(public, publicDir)
	off, on := callsFor(t, rec, localDir)
	if off == 0 || on != 0 {
		t.Fatalf("the local review started %d offline and %d online git commands; want all offline", off, on)
	}
	off, on = callsFor(t, rec, publicDir)
	if on == 0 || off != 0 {
		t.Fatalf("the public review started %d offline and %d online git commands; want all online", off, on)
	}

	// The same two, at the same time, each opened again while the other runs.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); run(local, localDir); _ = openIn(t, "") }()
		go func() { defer wg.Done(); run(public, publicDir); _ = openIn(t, "local") }()
	}
	wg.Wait()
	off, on = callsFor(t, rec, localDir)
	if on != 0 {
		t.Fatalf("a local review's git command carried no offline policy (%d offline, %d online)", off, on)
	}
	off, on = callsFor(t, rec, publicDir)
	if off != 0 {
		t.Fatalf("a public review's git command carried the offline policy (%d offline, %d online)", off, on)
	}
}
