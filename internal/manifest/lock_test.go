package manifest

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

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
