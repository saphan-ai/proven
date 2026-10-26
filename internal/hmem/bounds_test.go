package hmem_test

import (
	"errors"
	"io"
	"math"
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

// countedFile is a state file whose Stat reports size and whose body is
// supply bytes of white space after a valid empty state. It counts every byte
// read from it. A size below the body models a file that grew after the size
// check; a size of zero models a source whose size cannot be learned.
type countedFile struct {
	size   int64
	body   *strings.Reader
	closed bool
	read   int64
	// closeErr is what Close returns, to model a close that fails.
	closeErr error
	// readErr is what Read returns after the body, to model a failed read.
	readErr error
}

func newCountedFile(size int64, supply int) *countedFile {
	return &countedFile{size: size, body: strings.NewReader(emptyState + strings.Repeat(" ", supply))}
}

func (c *countedFile) Read(p []byte) (int, error) {
	n, err := c.body.Read(p)
	c.read += int64(n)
	if err == io.EOF && c.readErr != nil {
		err = c.readErr
	}
	return n, err
}

func (c *countedFile) Close() error { c.closed = true; return c.closeErr }

func (c *countedFile) Stat() (os.FileInfo, error) { return sizedInfo{c.size}, nil }

// sizedInfo is a FileInfo that reports only a size.
type sizedInfo struct{ size int64 }

func (i sizedInfo) Name() string       { return "state.json" }
func (i sizedInfo) Size() int64        { return i.size }
func (i sizedInfo) Mode() os.FileMode  { return 0o600 }
func (i sizedInfo) ModTime() time.Time { return time.Time{} }
func (i sizedInfo) IsDir() bool        { return false }
func (i sizedInfo) Sys() any           { return nil }

// TestStateReadConsumesAtMostCapPlusOneByte proves the read itself is bounded,
// not only the refusal: when the reported size passes the check but the source
// holds far more than the cap (a file that grew after the check, or a source
// whose size is unknown), the store still takes at most cap+1 bytes from it and
// refuses by name. A read that checks the size and then reads without a bound
// consumes the whole source and turns this test red on the byte count.
func TestStateReadConsumesAtMostCapPlusOneByte(t *testing.T) {
	const limit = 4096
	for name, reported := range map[string]int64{"grew after the size check": limit - 1, "size unknown": 0} {
		t.Run(name, func(t *testing.T) {
			f := newCountedFile(reported, limit*16)
			_, err := hmem.OpenWithStateFile(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: limit}}, f)
			if !errors.Is(err, hmem.ErrStateTooLarge) {
				t.Fatalf("a source past the cap was not refused by name: %v", err)
			}
			if f.read > limit+1 {
				t.Fatalf("the read took %d bytes from a source past the cap, at most %d allowed", f.read, limit+1)
			}
			if !f.closed {
				t.Fatal("the state file was not closed")
			}
		})
	}
}

// TestStateReadBelowTheCapReadsAll proves the bound does not cut a good file:
// a source under the cap is read in full and opens.
func TestStateReadBelowTheCapReadsAll(t *testing.T) {
	const limit = 4096
	f := newCountedFile(limit-1, 100)
	s, err := hmem.OpenWithStateFile(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Limits: hmem.Limits{MaxStateBytes: limit}}, f)
	if err != nil || s == nil {
		t.Fatalf("a source under the cap did not open: %v", err)
	}
	if f.read != int64(len(emptyState)+100) {
		t.Fatalf("the read took %d bytes, the source holds %d", f.read, len(emptyState)+100)
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

// TestScoreAgreesWithQueryAtANonDefaultLimit proves the Score helper trims the
// query to the configured token limit exactly as Query does: at limit 9 the
// score of a record matched only by the ninth word equals the relevance Query
// gives it, and at the default it equals the default-limit relevance.
func TestScoreAgreesWithQueryAtANonDefaultLimit(t *testing.T) {
	const q = "alpha bravo charlie delta echo foxtrot golf hotel zebra"
	for _, tokens := range []int{0, 9} {
		s, err := hmem.Open(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir(), Clock: func() time.Time { return time.UnixMilli(fixed) }, Limits: hmem.Limits{QueryTokens: tokens}})
		if err != nil {
			t.Fatal(err)
		}
		rec, err := s.Add(hmem.AddInput{Content: "zebra", Source: "manual", At: fixed})
		if err != nil {
			t.Fatal(err)
		}
		res, err := s.Query(q, 5, 1e-9)
		if err != nil || len(res.Hits) == 0 {
			t.Fatalf("limit %d: query returned no hit: %v", tokens, err)
		}
		if got := hmem.Score(rec, q, fixed, tokens); math.Abs(got-res.Hits[0].Relevance) > 1e-9 {
			t.Fatalf("limit %d: Score %v disagrees with Query %v", tokens, got, res.Hits[0].Relevance)
		}
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
