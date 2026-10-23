package manifest

import (
	"strings"
	"testing"
)

// TestPoetryDiffAddedRaisedUnchangedRemoved proves that poetry.lock yields
// added and raised packages with the line of their version, that keys in
// sub-tables and [metadata] are not read as a package, and that an unchanged,
// lowered or removed package yields nothing.
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

// TestPoetryMalformed proves that a file that is not a lock, an unclosed
// string and a package with no version each refuse the file, and that an
// empty file has no pins.
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

// TestPoetryRealisticLockFindsARaisedPackage proves that a poetry.lock with
// multi-line dependency arrays and a [metadata.files] table still yields a
// raised package. Those lines are not pins; refusing them would drop the file.
func TestPoetryRealisticLockFindsARaisedPackage(t *testing.T) {
	base := realisticPoetry("1.22.0")
	change := realisticPoetry("1.26.4")
	got, err := diffPoetry([]byte(base), []byte(change))
	if err != nil {
		t.Fatal(err)
	}
	wantLine := 0
	for i, line := range strings.Split(change, "\n") {
		if line == `version = "1.26.4"` {
			wantLine = i + 1
			break
		}
	}
	want := []Change{
		{Package: "numpy", Version: "1.26.4", Line: wantLine, Added: false},
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

// TestPoetryWithoutAPackageTableIsRefused proves that a file with no
// [[package]] table is refused, even when [metadata.files] holds a multi-line
// array. It is not a lock file.
func TestPoetryWithoutAPackageTableIsRefused(t *testing.T) {
	body := strings.Join([]string{
		"[metadata.files]",
		"numpy = [",
		`    {file = "numpy-1.26.4.tar.gz", hash = "sha256:abc"},`,
		"]",
		"",
	}, "\n")
	_, err := diffPoetry(nil, []byte(body))
	if err == nil || !strings.Contains(err.Error(), "no [[package]] table") {
		t.Fatal(err)
	}
}

// TestPoetryPEP503NamesAreTheSamePackage proves that Requests and requests are
// one package, and that zope_interface and zope.interface are one package. A
// higher pin is a raise. It is not an added package.
func TestPoetryPEP503NamesAreTheSamePackage(t *testing.T) {
	base := []byte(strings.Join([]string{
		"[[package]]",
		`name = "Requests"`,
		`version = "2.31.0"`,
		"",
		"[[package]]",
		`name = "zope_interface"`,
		`version = "5.0.0"`,
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"[[package]]",
		`name = "requests"`,
		`version = "2.31.0"`,
		"",
		"[[package]]",
		`name = "zope.interface"`,
		`version = "6.0.0"`,
		"",
	}, "\n"))
	got, err := diffPoetry(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "zope-interface", Version: "6.0.0", Line: 7, Added: false},
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

// realisticPoetry is a poetry.lock body. numpyVersion is the pin the test
// raises. The dependency arrays and [metadata.files] span more than one line,
// which is how Poetry writes them.
func realisticPoetry(numpyVersion string) string {
	return strings.Join([]string{
		"[[package]]",
		`name = "charset-normalizer"`,
		`version = "3.3.2"`,
		`description = "The Real First Universal Charset Detector."`,
		"",
		"[package.dependencies]",
		`python = ">=3.7"`,
		"",
		"[[package]]",
		`name = "numpy"`,
		`version = "` + numpyVersion + `"`,
		`description = "Fundamental package for array computing in Python."`,
		"",
		"[package.dependencies]",
		`python = ">=3.9"`,
		"numpy = [",
		`    {version = ">=1.22", python = "<3.12"},`,
		"]",
		"",
		"[[package]]",
		`name = "requests"`,
		`version = "2.31.0"`,
		`description = "Python HTTP for Humans."`,
		"",
		"[package.dependencies]",
		"charset-normalizer = [",
		`    {version = ">=2", markers = "python_version < \"3.12\""},`,
		"]",
		"",
		"[metadata]",
		`lock-version = "2.0"`,
		`python-versions = "^3.9"`,
		`content-hash = "0123456789abcdef"`,
		"",
		"[metadata.files]",
		"charset-normalizer = [",
		`    {file = "charset_normalizer-3.3.2.tar.gz", hash = "sha256:aaa"},`,
		"]",
		"numpy = [",
		`    {file = "numpy-` + numpyVersion + `.tar.gz", hash = "sha256:bbb"},`,
		"]",
		"requests = [",
		`    {file = "requests-2.31.0.tar.gz", hash = "sha256:ccc"},`,
		"]",
		"",
	}, "\n")
}

// TestPoetryMatch proves that the poetry.lock reader is a PyPI lock reader
// whose parse errors are notes, and that it matches only poetry.lock.
func TestPoetryMatch(t *testing.T) {
	r := Poetry()
	if r.Ecosystem != "PyPI" || !r.Lock || !r.NoteErr || !r.Match("poetry.lock") || !r.Match("svc/poetry.lock") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("poetry.lock.bak") || r.Match("Pipfile.lock") || r.Match("requirements.txt") {
		t.Fatal("matched a file that is not poetry.lock")
	}
}
