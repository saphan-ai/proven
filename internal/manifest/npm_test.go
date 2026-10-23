package manifest

import "testing"

// TestNPMDiffAddsRaisesAndSkips proves that package.json yields added and
// raised packages from dependencies and devDependencies with their lines,
// drops a lowered or equal version, keeps an unordered range change, and does
// not read peerDependencies.
func TestNPMDiffAddsRaisesAndSkips(t *testing.T) {
	base := []byte(`{
  "dependencies": {
    "old": "1.0.0",
    "lower": "2.0.0",
    "same": "1.2.3",
    "loose": "^1.0.0"
  }
}
`)
	change := []byte(`{
  "dependencies": {
    "old": "1.2.0",
    "lower": "1.0.0",
    "same": "v1.2.3",
    "loose": "^2.0.0",
    "added": "1.0.0",
    "@scope/pkg": "1.4.0"
  },
  "devDependencies": {
    "dev-added": "2.0.0"
  },
  "peerDependencies": {
    "peer": "18.0.0"
  }
}
`)
	got, err := diffNPM(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "@scope/pkg", Version: "1.4.0", Line: 8, Added: true},
		{Package: "added", Version: "1.0.0", Line: 7, Added: true},
		{Package: "dev-added", Version: "2.0.0", Line: 11, Added: true},
		{Package: "loose", Version: "^2.0.0", Line: 6, Added: false},
		{Package: "old", Version: "1.2.0", Line: 3, Added: false},
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

// TestNPMNotesANonString proves that a dependency whose value is not a string
// is skipped and named in a note, not read as a version.
func TestNPMNotesANonString(t *testing.T) {
	change := []byte("{\n  \"dependencies\": {\n    \"foo\": {\"version\": \"1.0.0\"}\n  }\n}\n")
	got, err := diffNPM(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("%#v", got)
	}
	notes := npmNotes(change)
	if len(notes) != 1 || notes[0] != "dependencies.foo is not a string" {
		t.Fatalf("%#v", notes)
	}
}

// TestNPMRejectsBrokenJSON proves that a package.json that is not a JSON
// object is an error, never an empty set of dependencies.
func TestNPMRejectsBrokenJSON(t *testing.T) {
	if _, err := diffNPM(nil, []byte("{")); err == nil {
		t.Fatal("broken package.json was accepted")
	}
	if _, err := diffNPM(nil, []byte("[]")); err == nil {
		t.Fatal("an array was accepted")
	}
}

// TestNPMLaterSectionWins proves that a name in both sections takes the later
// section's version and line.
func TestNPMLaterSectionWins(t *testing.T) {
	change := []byte("{\n  \"dependencies\": {\n    \"a\": \"1.0.0\"\n  },\n  \"devDependencies\": {\n    \"a\": \"2.0.0\"\n  }\n}\n")
	got, err := diffNPM(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Version != "2.0.0" || got[0].Line != 6 || !got[0].Added {
		t.Fatalf("%#v", got)
	}
}

// TestNPMMatch proves that the npm reader matches package.json at any depth and
// no other file name.
func TestNPMMatch(t *testing.T) {
	r := NPM()
	if r.Ecosystem != "npm" || !r.Match("package.json") || !r.Match("web/package.json") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("package-lock.json") || r.Match("not-package.json") || r.Match("pkg/package.json.bak") {
		t.Fatal("matched a file that is not package.json")
	}
}
