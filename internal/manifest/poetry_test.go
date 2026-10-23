package manifest

import (
	"strings"
	"testing"
)

func TestPoetryDiffAddedRaisedUnchangedRemoved(t *testing.T) {
	base := []byte(strings.Join([]string{
		"[[package]]",
		`name = "flask"`,
		`version = "1.0"`,
		"",
		"[[package]]",
		`name = "same"`,
		`version = "1.2.0"`,
		"",
		"[[package]]",
		`name = "lower"`,
		`version = "2.0"`,
		"",
		"[[package]]",
		`name = "gone"`,
		`version = "1.0"`,
		"",
		"[metadata]",
		`lock-version = "2.0"`,
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"[[package]]",
		`name = "flask"`,
		`version = "1.2.0"`,
		`description = "A web framework."`,
		"",
		"[[package]]",
		`name = "same"`,
		`version = "1.2.0"`,
		"",
		"[[package]]",
		`name = "lower"`,
		`version = "1.0"`,
		"",
		"[[package]]",
		`name = 'newpkg'`,
		`version = '1.0a1'`,
		"",
		"[package.dependencies]",
		`jinja2 = ">=2"`,
		"",
		"[metadata]",
		`lock-version = "2.0"`,
		`python-versions = "*"`,
		"",
	}, "\n"))
	got, err := diffPoetry(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "flask", Version: "1.2.0", Line: 3, Added: false},
		{Package: "newpkg", Version: "1.0a1", Line: 16, Added: true},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got[i], want[i])
		}
	}
}

func TestPoetryMalformed(t *testing.T) {
	if _, err := diffPoetry(nil, []byte("this is not a lock\n")); err == nil || !strings.Contains(err.Error(), "line 1: not a package table") {
		t.Fatal(err)
	}
	broken := "[[package]]\nname = \"flask\"\nversion = \"1.0\n"
	if _, err := diffPoetry(nil, []byte(broken)); err == nil || !strings.Contains(err.Error(), "unclosed string") {
		t.Fatal(err)
	}
	missing := "[[package]]\nname = \"flask\"\n\n[metadata]\nlock-version = \"2.0\"\n"
	if _, err := diffPoetry(nil, []byte(missing)); err == nil || !strings.Contains(err.Error(), "has no version") {
		t.Fatal(err)
	}
	got, err := diffPoetry(nil, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty: %v %#v", err, got)
	}
}

func TestPoetryMatch(t *testing.T) {
	r := Poetry()
	if r.Ecosystem != "PyPI" || !r.Lock || !r.NoteErr || !r.Match("poetry.lock") || !r.Match("svc/poetry.lock") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("poetry.lock.bak") || r.Match("Pipfile.lock") || r.Match("requirements.txt") {
		t.Fatal("matched a file that is not poetry.lock")
	}
}
