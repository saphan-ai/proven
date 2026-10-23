package manifest

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestLockVersionWinsOverTheManifest proves that when a manifest and the lock
// file beside it name one package, only the lock file's exact version is kept,
// for Go, npm and PyPI, that PyPI names pair up without regard to case, and
// that a package only the manifest names is still kept.
func TestLockVersionWinsOverTheManifest(t *testing.T) {
	files := []change.File{
		{Path: "go.mod", Change: []byte("module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n")},
		{Path: "go.sum", Change: []byte("example.com/mod v1.2.3 h1:abc\nexample.com/mod v1.2.3/go.mod h1:def\nexample.com/trans v0.4.0 h1:xyz\n")},
		{Path: "package.json", Change: []byte("{\"dependencies\":{\"left-pad\":\"^1.0.0\"}}\n")},
		{Path: "package-lock.json", Change: []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "node_modules/left-pad": {"version": "1.2.3"},
    "node_modules/once": {"version": "1.4.0"}
  }
}
`)},
		{Path: "requirements.txt", Change: []byte("Flask==1.0.0\nonlyreq==2.0.0\n")},
		{Path: "poetry.lock", Change: []byte("[[package]]\nname = \"flask\"\nversion = \"1.2.3\"\n")},
	}
	got, notes, err := Collect(files, []Reader{Go(), GoSum(), NPM(), NPMLock(), PyPI(), Poetry()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	want := []Hit{
		{Ecosystem: "Go", File: "go.sum", Lock: true, Change: Change{Package: "example.com/mod", Version: "v1.2.3", Line: 1, Added: true}},
		{Ecosystem: "Go", File: "go.sum", Lock: true, Change: Change{Package: "example.com/trans", Version: "v0.4.0", Line: 3, Added: true}},
		{Ecosystem: "npm", File: "package-lock.json", Lock: true, Change: Change{Package: "left-pad", Version: "1.2.3", Line: 4, Added: true}},
		{Ecosystem: "npm", File: "package-lock.json", Lock: true, Change: Change{Package: "once", Version: "1.4.0", Line: 5, Added: true}},
		{Ecosystem: "PyPI", File: "poetry.lock", Lock: true, Change: Change{Package: "flask", Version: "1.2.3", Line: 3, Added: true}},
		{Ecosystem: "PyPI", File: "requirements.txt", Lock: false, Change: Change{Package: "onlyreq", Version: "2.0.0", Line: 2, Added: true}},
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

// TestPyPILockPairsNamesAfterPEP503 proves that requirements.txt and poetry.lock
// in one directory are the same package when the spellings differ only by PEP
// 503. The lock file's version is kept. A package only the manifest names is kept.
func TestPyPILockPairsNamesAfterPEP503(t *testing.T) {
	got, notes, err := Collect([]change.File{
		{Path: "requirements.txt", Change: []byte("zope_interface==5.0.0\nRequests==2.28.0\nonlyreq==1.0\n")},
		{Path: "poetry.lock", Change: []byte("[[package]]\nname = \"zope.interface\"\nversion = \"5.4.0\"\n\n[[package]]\nname = \"requests\"\nversion = \"2.31.0\"\n")},
	}, []Reader{PyPI(), Poetry()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	want := []Hit{
		{Ecosystem: "PyPI", File: "poetry.lock", Lock: true, Change: Change{Package: "requests", Version: "2.31.0", Line: 7, Added: true}},
		{Ecosystem: "PyPI", File: "poetry.lock", Lock: true, Change: Change{Package: "zope-interface", Version: "5.4.0", Line: 3, Added: true}},
		{Ecosystem: "PyPI", File: "requirements.txt", Lock: false, Change: Change{Package: "onlyreq", Version: "1.0", Line: 3, Added: true}},
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

// TestLockInAnotherDirectoryDoesNotWin proves that a lock file pairs only with
// the manifest in its own directory: a go.sum at the root does not hide a
// require in svc/go.mod.
func TestLockInAnotherDirectoryDoesNotWin(t *testing.T) {
	got, _, err := Collect([]change.File{
		{Path: "svc/go.mod", Change: []byte("module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n")},
		{Path: "go.sum", Change: []byte("example.com/mod v9.9.9 h1:abc\n")},
	}, []Reader{Go(), GoSum()})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].File != "go.sum" || got[1].File != "svc/go.mod" || got[1].Version != "v1.0.0" {
		t.Fatalf("%#v", got)
	}
}

// TestUnparsedLockIsANote proves that a lock file that does not parse is a
// trace note naming the file and the reason, not an error, and that the
// manifest beside it is still read.
func TestUnparsedLockIsANote(t *testing.T) {
	got, notes, err := Collect([]change.File{
		{Path: "go.mod", Change: []byte("module m\n\ngo 1.26\n\nrequire example.com/a v1.0.0\n")},
		{Path: "go.sum", Change: []byte("not-a-sum\n")},
		{Path: "package-lock.json", Change: []byte("{\n")},
		{Path: "poetry.lock", Change: []byte("this is not a lock\n")},
	}, []Reader{Go(), GoSum(), NPMLock(), Poetry()})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].File != "go.mod" || got[0].Package != "example.com/a" {
		t.Fatalf("%#v", got)
	}
	body := strings.Join(notes, "\n")
	for _, want := range []string{
		"go.sum: change: line 1: not a module version hash",
		"package-lock.json: change: invalid JSON:",
		"poetry.lock: change: line 1: not a package table",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}
