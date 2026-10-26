package hmem_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

// emptyState is a valid, empty state file body. JSON ignores trailing white
// space, so padding it with spaces gives a valid file of any size.
const emptyState = `{"version":2,"seq":0,"records":[]}`

// stateOfSize writes a valid state.json of exactly n bytes into a new directory.
func stateOfSize(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	body := emptyState + strings.Repeat(" ", n-len(emptyState))
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestStateFileOverTheCapIsRefusedByName proves the growth bound: a state file
// one byte over the cap is refused with ErrStateTooLarge, and the message names
// the file's size and the cap. It turns red when the cap is removed.
func TestStateFileOverTheCapIsRefusedByName(t *testing.T) {
	const limit = 4096
	dir := stateOfSize(t, limit+1)
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: limit}})
	if !errors.Is(err, hmem.ErrStateTooLarge) {
		t.Fatalf("a state file over the cap was not refused by name: %v", err)
	}
	if !strings.Contains(err.Error(), "4097 bytes") || !strings.Contains(err.Error(), "cap is 4096") {
		t.Fatalf("the refusal does not name the size and the cap: %v", err)
	}
}

// TestStateFileAtTheCapOpens proves the cap is inclusive: a file of exactly the
// cap opens, so the refusal is for a file over it and nothing smaller.
func TestStateFileAtTheCapOpens(t *testing.T) {
	const limit = 4096
	dir := stateOfSize(t, limit)
	s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: limit}})
	if err != nil || s == nil {
		t.Fatalf("a state file at the cap did not open: %v", err)
	}
}

// TestDefaultCapIsInForce proves a store opened with no Limits still refuses a
// file over the default cap, so leaving the setting out never means unbounded.
func TestDefaultCapIsInForce(t *testing.T) {
	dir := t.TempDir()
	body := emptyState + strings.Repeat(" ", int(hmem.DefaultMaxStateBytes)+1-len(emptyState))
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()})
	if !errors.Is(err, hmem.ErrStateTooLarge) {
		t.Fatalf("the default cap was not applied: %v", err)
	}
}

// TestQueryTokenLimitIsASetting proves the token limit is honoured at a
// non-default value: a record matched only by the ninth query word scores
// higher when the limit is nine than when it is the default eight.
func TestQueryTokenLimitIsASetting(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	clock := func() time.Time { return time.UnixMilli(fixed) }
	seed, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: clock})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Add(hmem.AddInput{Content: "zebra", Source: "manual"}); err != nil {
		t.Fatal(err)
	}
	const q = "alpha bravo charlie delta echo foxtrot golf hotel zebra"
	relevance := func(tokens int) float64 {
		s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: clock, Limits: hmem.Limits{QueryTokens: tokens}})
		if err != nil {
			t.Fatal(err)
		}
		res, err := s.Query(q, 5, 1e-9)
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Hits) == 0 {
			return 0
		}
		return res.Hits[0].Relevance
	}
	def, nine := relevance(0), relevance(9)
	if !(nine > def) {
		t.Fatalf("the ninth token did not count at limit 9: default %v, nine %v", def, nine)
	}
}

// TestLoadLimitsReadsAndRefuses proves the settings: empty means the default,
// a good value is read, and a bad value is refused by variable name.
func TestLoadLimitsReadsAndRefuses(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	got, err := hmem.LoadLimits(env(nil))
	if err != nil || got != hmem.DefaultLimits() {
		t.Fatalf("defaults: %v %v", got, err)
	}
	got, err = hmem.LoadLimits(env(map[string]string{hmem.EnvMaxStateBytes: "1000", hmem.EnvQueryTokens: "3"}))
	if err != nil || got.MaxStateBytes != 1000 || got.QueryTokens != 3 {
		t.Fatalf("values: %v %v", got, err)
	}
	for _, name := range []string{hmem.EnvMaxStateBytes, hmem.EnvQueryTokens} {
		for _, bad := range []string{"0", "-1", "x", "99999999999999999999"} {
			_, err := hmem.LoadLimits(env(map[string]string{name: bad}))
			var ce *hmem.ConfigError
			if !errors.As(err, &ce) || ce.Field != name {
				t.Fatalf("%s=%q was not refused by name: %v", name, bad, err)
			}
		}
	}
}
