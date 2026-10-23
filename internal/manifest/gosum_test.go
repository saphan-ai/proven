package manifest

import (
	"strings"
	"testing"
)

// TestGoSumDiffAddedRaisedUnchangedRemoved proves that go.sum yields added and
// raised modules, and nothing for an unchanged, lowered or removed one.
func TestGoSumDiffAddedRaisedUnchangedRemoved(t *testing.T) {
	base := []byte(strings.Join([]string{
		"example.com/gone v0.1.0 h1:gone",
		"example.com/lower v1.2.0 h1:lower",
		"example.com/lower v1.2.0/go.mod h1:lowermod",
		"example.com/old v1.0.0 h1:old",
		"example.com/old v1.0.0/go.mod h1:oldmod",
		"example.com/same v1.2.3 h1:same",
		"example.com/same v1.2.3/go.mod h1:samemod",
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"example.com/lower v1.0.0 h1:lower",
		"example.com/lower v1.0.0/go.mod h1:lowermod",
		"example.com/new v0.1.0 h1:new",
		"example.com/new v0.1.0/go.mod h1:newmod",
		"example.com/old v1.2.0 h1:old",
		"example.com/old v1.2.0/go.mod h1:oldmod",
		"example.com/same v1.2.3 h1:same",
		"example.com/same v1.2.3/go.mod h1:samemod",
		"",
	}, "\n"))
	got, err := diffGoSum(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "example.com/new", Version: "v0.1.0", Line: 3, Added: true},
		{Package: "example.com/old", Version: "v1.2.0", Line: 5, Added: false},
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

// TestGoSumDropsTheGoModDuplicate proves that the /go.mod checksum line is not
// read as a second module version.
func TestGoSumDropsTheGoModDuplicate(t *testing.T) {
	// The /go.mod line is last on purpose. It is the checksum of the module's
	// go.mod, not a second version. Keeping it reports that line as its own
	// version and this test goes red.
	change := []byte("example.com/mod v1.2.3 h1:abc\nexample.com/mod v1.2.3/go.mod h1:def\n")
	got, err := diffGoSum(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Package != "example.com/mod" || got[0].Version != "v1.2.3" || got[0].Line != 1 || !got[0].Added {
		t.Fatalf("duplicate go.mod line kept: %#v", got)
	}
}

// TestGoSumReportsEveryNewVersion proves that a new module pinned at two
// versions reports both, each with its own line.
func TestGoSumReportsEveryNewVersion(t *testing.T) {
	change := []byte("example.com/mod v1.2.3 h1:new\nexample.com/mod v1.0.0 h1:old\n")
	got, err := diffGoSum(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "example.com/mod", Version: "v1.0.0", Line: 2, Added: true},
		{Package: "example.com/mod", Version: "v1.2.3", Line: 1, Added: true},
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

// TestGoSumMalformed proves that a line that is not path, version and h1 hash
// refuses the file and names the line, and that an empty file has no pins.
func TestGoSumMalformed(t *testing.T) {
	if _, err := diffGoSum(nil, []byte("not a sum line\n")); err == nil || !strings.Contains(err.Error(), "line 1: not a module version hash") {
		t.Fatal(err)
	}
	if _, err := diffGoSum(nil, []byte("example.com/mod v1.2.3\n")); err == nil {
		t.Fatal("a line without a hash was accepted")
	}
	got, err := diffGoSum(nil, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty: %v %#v", err, got)
	}
}

// TestGoSumMatch proves that the go.sum reader is a lock reader whose parse
// errors are notes, and that it matches go.sum at any depth and nothing else.
func TestGoSumMatch(t *testing.T) {
	r := GoSum()
	if r.Ecosystem != "Go" || !r.Lock || !r.NoteErr || !r.Match("go.sum") || !r.Match("svc/go.sum") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("go.mod") || r.Match("notgo.sum") || r.Match("go.sum.bak") {
		t.Fatal("matched a file that is not go.sum")
	}
}
