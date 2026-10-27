package change

import (
	"os/exec"
	"strings"
	"testing"
)

// Property: when a command fails, the message comes from its standard error and never from its
// standard output, which is repository content (ADR-0013).
func TestFailureMessageNeverQuotesStandardOutput(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo repository-content; echo the-error >&2; exit 1")
	_, err := ProcessPolicy().outputBounded(cmd, 1<<10, "the test output", EnvMaxFileBytes)
	if err == nil {
		t.Fatal("a failing command returned no error")
	}
	if strings.Contains(err.Error(), "repository-content") {
		t.Fatalf("the failure quoted standard output: %v", err)
	}
	if !strings.Contains(err.Error(), "the-error") {
		t.Fatalf("the failure lost the standard error: %v", err)
	}
}

// Property: a failing command that writes a great deal to standard error gives a clipped message,
// so the error cannot grow with what a repository makes git print.
func TestFailureMessageIsClipped(t *testing.T) {
	cmd := exec.Command("sh", "-c", "head -c 200000 /dev/zero | tr '\\0' 'e' >&2; exit 1")
	_, err := ProcessPolicy().outputBounded(cmd, 1<<10, "the test output", EnvMaxFileBytes)
	if err == nil {
		t.Fatal("a failing command returned no error")
	}
	if len(err.Error()) > 1<<10 {
		t.Fatalf("the failure message is %d bytes", len(err.Error()))
	}
}

// Property: the commit log has the file bound on its own: a log over the bound is refused naming the
// log, and one at or under it is read.
func TestMessagesRefuseALogOverTheFileBound(t *testing.T) {
	defer UseLimits(UseLimits(Limits{Archive: 1 << 20, File: 1 << 10, Config: 1 << 20, Record: 1 << 20}))
	dir, base := commitFile(t, "a.txt", 1)
	write(t, dir+"/b.txt", "b")
	git(t, dir, "add", "b.txt")
	git(t, dir, "commit", "-m", "[change] git: second", "-m", strings.Repeat("m", 4<<10))
	_, err := Messages(dir, base, "")
	if err == nil || !strings.Contains(err.Error(), "the commit log") || strings.Contains(err.Error(), "mmmm") {
		t.Fatalf("a commit log over the bound was not refused by name: %v", err)
	}
	UseLimits(Limits{Archive: 1 << 20, File: 8 << 10, Config: 1 << 20, Record: 1 << 20})
	if _, err := Messages(dir, base, ""); err != nil {
		t.Fatalf("a commit log under the bound was refused: %v", err)
	}
}

// Property: a size setting that is not a positive number of bytes is refused by name, and the value
// the refusal shows is clipped (ADR-0013).
func TestLoadLimitsRefusesAWrongValueByName(t *testing.T) {
	for _, raw := range []string{"0", "-5", "ten", "1.5", strings.Repeat("9", 100)} {
		_, err := LoadLimits(func(k string) string {
			if k == EnvMaxRecordBytes {
				return raw
			}
			return ""
		})
		if err == nil || !strings.Contains(err.Error(), EnvMaxRecordBytes) {
			t.Fatalf("%q: not refused by name: %v", raw, err)
		}
		if len(err.Error()) > 300 {
			t.Fatalf("%q: the refusal echoed the whole value", raw)
		}
	}
}

// Property: a partly filled Limits never becomes an unbounded one: a field at zero keeps its default.
func TestUseLimitsKeepsTheDefaultForAZeroField(t *testing.T) {
	defer UseLimits(UseLimits(DefaultLimits()))
	UseLimits(Limits{File: 7})
	got, d := Bounds(), DefaultLimits()
	if got.File != 7 || got.Archive != d.Archive || got.Config != d.Config || got.Record != d.Record {
		t.Fatalf("%+v", got)
	}
}
