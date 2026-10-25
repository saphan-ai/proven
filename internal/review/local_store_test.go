package review

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// localEnv is a live environment with the local evidence store and the
// guardrail off, so the stub generator's calls reach the evidence store.
func localEnv(file string) func(string) string {
	extra := map[string]string{
		"PROVEN_EVIDENCE_STORE": "local",
		"PROVEN_GUARDRAIL_MODE": "off",
	}
	if file != "" {
		extra["PROVEN_EVIDENCE_FILE"] = file
	}
	return liveSecrets(extra)
}

// TestLocalStoreWithoutAFileIsRefusedByName proves a local store with no
// PROVEN_EVIDENCE_FILE stops Open and names the setting (ADR-0013,
// ADR-0047). The mutated field is PROVEN_EVIDENCE_FILE, left empty.
// Negative control: restoring a default file turns this red.
func TestLocalStoreWithoutAFileIsRefusedByName(t *testing.T) {
	_, err := Open(Options{LogDir: t.TempDir(), Env: localEnv(""), testRun: quietRun{}})
	var cfg *ConfigError
	if !errors.As(err, &cfg) || cfg.Field != "PROVEN_EVIDENCE_FILE" {
		t.Fatalf("want a refusal naming PROVEN_EVIDENCE_FILE, got %v", err)
	}
}

// TestLocalFileInsideTheTargetIsRefused proves a local evidence file that
// resolves inside the repository under review is refused by name, and that
// nothing is written into that repository (ADR-0047). The mutated field is
// PROVEN_EVIDENCE_FILE, set inside the target, then through a link that
// points into it. Negative control: removing the check turns this red.
func TestLocalFileInsideTheTargetIsRefused(t *testing.T) {
	repo, base, head := twoCommits(t)
	link := filepath.Join(t.TempDir(), "into")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	for name, file := range map[string]string{
		"plain": filepath.Join(repo, "evidence.sql"),
		"link":  filepath.Join(link, "evidence.sql"),
	} {
		rev, err := Open(Options{LogDir: t.TempDir(), Env: localEnv(file), gen: &countGen{}, testRun: quietRun{}})
		if err != nil {
			t.Fatalf("%s: open: %v", name, err)
		}
		_, err = reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
		var cfg *ConfigError
		if !errors.As(err, &cfg) || cfg.Field != "PROVEN_EVIDENCE_FILE" {
			t.Fatalf("%s: want a refusal naming PROVEN_EVIDENCE_FILE, got %v", name, err)
		}
		if _, serr := os.Stat(filepath.Join(repo, "evidence.sql")); !os.IsNotExist(serr) {
			t.Fatalf("%s: a file was written into the reviewed tree", name)
		}
	}
}

// TestEmptyLocalStoreNamesLocalAsDown proves a review on a fresh local store
// completes and names `local` among the failed services, instead of ending
// on a plain error (ADR-0047). The mutated field is the store: its file does
// not exist, so no table does. Negative control: removing the down conversion
// in the file's query makes the review return an error.
func TestEmptyLocalStoreNamesLocalAsDown(t *testing.T) {
	repo, base, head := twoCommits(t)
	file := filepath.Join(t.TempDir(), "evidence.sql")
	rev, err := Open(Options{LogDir: t.TempDir(), Env: localEnv(file), gen: &countGen{}, testRun: quietRun{}})
	if err != nil {
		t.Fatal(err)
	}
	res, err := reviewWith(t, rev, Request{Target: repo, BaselineCommit: base, HeadCommit: head})
	if err != nil {
		t.Fatalf("the review ended: %v", err)
	}
	var named bool
	for _, s := range res.FailedServices {
		named = named || s.Service == "local"
	}
	if !named {
		t.Fatalf("failed services %#v", res.FailedServices)
	}
}
