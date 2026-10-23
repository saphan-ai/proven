package manifest

import (
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

func TestRequirementsDiffAddsAndRaises(t *testing.T) {
	base := []byte("flask==1.0\nsame==1.2.0\nlower==2.0\npre==1.0a1\n")
	change := []byte("flask==1.2.0\nsame==1.2\nlower==1.0\npre==1.0\nnewpkg==1.0a1\nDjango>=3.2\n# comment\n\n-r extra.txt\n")
	got, err := diffRequirements(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "flask", Version: "1.2.0", Line: 1, Added: false},
		{Package: "newpkg", Version: "1.0a1", Line: 5, Added: true},
		{Package: "pre", Version: "1.0", Line: 4, Added: false},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got[i], want[i])
		}
	}
	notes := requirementsNotes(change)
	wantNotes := []string{
		`line 6 skipped ("Django>=3.2"): not a pinned name==version`,
		`line 9 skipped ("-r extra.txt"): not a pinned name==version`,
	}
	if len(notes) != len(wantNotes) {
		t.Fatalf("%#v", notes)
	}
	for i := range wantNotes {
		if notes[i] != wantNotes[i] {
			t.Fatalf("note %q", notes[i])
		}
	}
}

func TestRequirementsPinsCommonForms(t *testing.T) {
	change := []byte(joinLines(
		"pkg[extra]==1.4.0",
		"spaced == 1.0",
		"marked==1.2.3 ; python_version == \"3.11\"",
		"hashed==1.0 --hash=sha256:abc",
		"commented==1.0 # pinned",
		"exact===1.0",
		"Flask==2.0.1",
		"url @ https://example.com/pkg.whl",
	))
	got, err := diffRequirements(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "Flask", Version: "2.0.1", Line: 7, Added: true},
		{Package: "commented", Version: "1.0", Line: 5, Added: true},
		{Package: "hashed", Version: "1.0", Line: 4, Added: true},
		{Package: "marked", Version: "1.2.3", Line: 3, Added: true},
		{Package: "pkg", Version: "1.4.0", Line: 1, Added: true},
		{Package: "spaced", Version: "1.0", Line: 2, Added: true},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got[i], want[i])
		}
	}
	notes := requirementsNotes(change)
	if len(notes) != 2 {
		t.Fatalf("%#v", notes)
	}
	if notes[0] != `line 6 skipped ("exact===1.0"): not a pinned name==version` {
		t.Fatalf("note %q", notes[0])
	}
	if notes[1] != `line 8 skipped ("url @ https://example.com/pkg.whl"): not a pinned name==version` {
		t.Fatalf("note %q", notes[1])
	}
}

func TestRequirementsLastLineWins(t *testing.T) {
	got, err := diffRequirements(nil, []byte("pkg==1.0\npkg==2.0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Version != "2.0" || got[0].Line != 2 {
		t.Fatalf("%#v", got)
	}
}

func TestRequirementsUnparsedChangeIsStillReported(t *testing.T) {
	got, err := diffRequirements([]byte("pkg==1.0\n"), []byte("pkg==latest\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Version != "latest" || got[0].Added {
		t.Fatalf("%#v", got)
	}
}

func TestRequirementsMatch(t *testing.T) {
	r := PyPI()
	if r.Ecosystem != "PyPI" || !r.Match("requirements.txt") || !r.Match("svc/requirements.txt") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("requirements-dev.txt") || r.Match("Requirements.txt") || r.Match("pkg/requirements.txt.bak") {
		t.Fatal("matched a file that is not requirements.txt")
	}
}

func TestCollectCarriesSkipNotes(t *testing.T) {
	got, notes, err := Collect([]change.File{{
		Path:   "app/requirements.txt",
		Change: []byte("Flask==1.0\nDjango>=2\n"),
	}}, []Reader{PyPI()})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Ecosystem != "PyPI" || got[0].Package != "Flask" {
		t.Fatalf("%#v", got)
	}
	if len(notes) != 1 || notes[0] != `app/requirements.txt: line 2 skipped ("Django>=2"): not a pinned name==version` {
		t.Fatalf("%#v", notes)
	}
}

func joinLines(lines ...string) string {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}
	return out
}
