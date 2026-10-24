package manifest

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestCargoDiffAddsRaisesAndSkips proves Cargo.toml yields an added crate
// and a raised one, drops a lower version, keeps a constraint that does not
// order, reads a renamed inline table and a dev-dependency, and does not
// read build-dependencies or a crate with no version.
func TestCargoDiffAddsRaisesAndSkips(t *testing.T) {
	base := []byte(strings.Join([]string{
		"[dependencies]",
		"serde = \"1.0.100\"",
		"libc = \"0.2.150\"",
		"old = \"1.0.0\"",
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"[dependencies]",
		"serde = \"1.0.210\"",
		"libc = \"0.2.100\"",
		"old = \"^2.0.0\"",
		"regex = \"1.9.1\"",
		"local = { path = \"../local\" }",
		"mine = { package = \"bytes\", version = \"1.5.0\" }",
		"num = 1",
		"",
		"[dev-dependencies]",
		"tempfile = \"3.10.1\"",
		"",
		"[build-dependencies]",
		"cc = \"1.0.0\"",
		"",
		"[workspace.dependencies]",
		"shared = \"1.2.3\"",
		"",
	}, "\n"))
	got, err := diffCargo(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "bytes", Version: "1.5.0", Line: 7, Added: true},
		{Package: "old", Version: "^2.0.0", Line: 4, Added: false},
		{Package: "regex", Version: "1.9.1", Line: 5, Added: true},
		{Package: "serde", Version: "1.0.210", Line: 2, Added: false},
		{Package: "tempfile", Version: "3.10.1", Line: 11, Added: true},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
	notes := cargoNotes(change)
	joined := strings.Join(notes, "\n")
	if !strings.Contains(joined, "dependencies.local has no version") || !strings.Contains(joined, "dependencies.num is not a string") {
		t.Fatalf("notes %q", notes)
	}
	if strings.Contains(joined, "cc") || strings.Contains(joined, "shared") || strings.Contains(joined, "serde") {
		t.Fatal(notes)
	}
}

// TestCargoReadsADottedTable proves [dependencies.name] is the same crate
// as a line in [dependencies], including a quoted name, and a table with
// no version is a note.
func TestCargoReadsADottedTable(t *testing.T) {
	change := []byte(strings.Join([]string{
		"[dependencies.tokio]",
		"version = \"1.35.0\"",
		"features = [",
		"  \"full\",",
		"]",
		"",
		"[dependencies.\"serde_json\"]",
		"version = \"1.0.108\"",
		"",
		"[dev-dependencies.local]",
		"path = \"../local\"",
		"",
	}, "\n"))
	got, err := diffCargo(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "serde_json", Version: "1.0.108", Line: 8, Added: true},
		{Package: "tokio", Version: "1.35.0", Line: 2, Added: true},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
	notes := strings.Join(cargoNotes(change), "\n")
	if !strings.Contains(notes, "dev-dependencies.local has no version") {
		t.Fatal(notes)
	}
}

// TestCargoRejectsABrokenTable proves a dependencies line that is not an
// assignment fails the check instead of reading as no dependency.
func TestCargoRejectsABrokenTable(t *testing.T) {
	if _, err := diffCargo(nil, []byte("[dependencies]\nnot a dependency\n")); err == nil {
		t.Fatal("broken Cargo.toml was accepted")
	}
}

// TestCargoEmptyManifestIsNoDependency proves an empty file and a manifest
// with no dependency table contribute nothing.
func TestCargoEmptyManifestIsNoDependency(t *testing.T) {
	got, err := diffCargo(nil, []byte("[package]\nname = \"app\"\nversion = \"0.1.0\"\n"))
	if err != nil || len(got) != 0 {
		t.Fatalf("%#v %v", got, err)
	}
	got, err = diffCargo(nil, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("%#v %v", got, err)
	}
}

// TestCargoMatch proves the reader matches Cargo.toml at any depth and not
// a lock file, a backup, or a different capitalisation.
func TestCargoMatch(t *testing.T) {
	r := Cargo()
	if r.Ecosystem != "crates.io" || r.Lock || r.Compare == nil || r.Explain == nil {
		t.Fatalf("%#v", r.Ecosystem)
	}
	if !r.Match("Cargo.toml") || !r.Match("app/Cargo.toml") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("Cargo.lock") || r.Match("cargo.toml") || r.Match("notCargo.toml") || r.Match("Cargo.toml.bak") {
		t.Fatal("matched a file that is not Cargo.toml")
	}
}

// TestCargoLockWinsOverTheManifest proves the lock file's exact version is
// the one kept when both files name the crate, and a crate only the manifest
// names is still kept.
func TestCargoLockWinsOverTheManifest(t *testing.T) {
	files := []change.File{
		{Path: "Cargo.toml", Change: []byte("[dependencies]\nregex = \"^1.9\"\nonly = \"1.2.3\"\n")},
		{Path: "Cargo.lock", Change: []byte("[[package]]\nname = \"regex\"\nversion = \"1.9.1\"\n")},
	}
	got, notes, err := Collect(files, []Reader{Cargo(), CargoLock()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	if got[0].File != "Cargo.lock" || got[0].Package != "regex" || got[0].Version != "1.9.1" || !got[0].Lock {
		t.Fatalf("%#v", got[0])
	}
	if got[1].File != "Cargo.toml" || got[1].Package != "only" || got[1].Lock {
		t.Fatalf("%#v", got[1])
	}
}
