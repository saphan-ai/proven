package manifest

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestComposerDiffAddsRaisesAndSkips proves that composer.json yields an
// added package and a raised one, drops a lower version, keeps a constraint
// that does not order, and does not report a platform package.
func TestComposerDiffAddsRaisesAndSkips(t *testing.T) {
	base := []byte(`{
  "require": {
    "monolog/monolog": "2.0.0",
    "symfony/console": "6.0.0"
  }
}
`)
	change := []byte(`{
  "require": {
    "monolog/monolog": "2.9.1",
    "symfony/console": "5.0.0",
    "guzzlehttp/guzzle": "^7.0"
  },
  "require-dev": {
    "phpunit/phpunit": "10.5.0",
    "php": "^8.2",
    "ext-json": "*"
  }
}
`)
	got, err := diffComposer(base, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("%#v", got)
	}
	// diffNamed sorts by package name.
	want := []Change{
		{Package: "guzzlehttp/guzzle", Version: "^7.0", Line: 5, Added: true},
		{Package: "monolog/monolog", Version: "2.9.1", Line: 3, Added: false},
		{Package: "phpunit/phpunit", Version: "10.5.0", Line: 8, Added: true},
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
	notes := composerNotes(change)
	joined := strings.Join(notes, "\n")
	if !strings.Contains(joined, "php is a platform package") || !strings.Contains(joined, "ext-json is a platform package") {
		t.Fatalf("notes %q", notes)
	}
	if strings.Contains(joined, "monolog") {
		t.Fatal(notes)
	}
}

// TestComposerFoldsTheName proves a Packagist name is stored in lower case,
// so Monolog/monolog and monolog/monolog are one package.
func TestComposerFoldsTheName(t *testing.T) {
	got, err := diffComposer(nil, []byte("{\"require\":{\"Monolog/monolog\":\"2.0.0\"}}\n"))
	if err != nil || len(got) != 1 || got[0].Package != "monolog/monolog" {
		t.Fatalf("%#v %v", got, err)
	}
}

// TestComposerRejectsBrokenJSON proves a composer.json that is not a JSON
// object fails the check instead of reading as no dependency.
func TestComposerRejectsBrokenJSON(t *testing.T) {
	if _, err := diffComposer(nil, []byte("{")); err == nil {
		t.Fatal("broken composer.json was accepted")
	}
}

// TestComposerMatch proves the reader matches composer.json at any depth and
// not a lock file or a backup.
func TestComposerMatch(t *testing.T) {
	r := Composer()
	if r.Ecosystem != "Packagist" || r.Lock || r.Compare == nil || r.Explain == nil {
		t.Fatalf("%#v", r.Ecosystem)
	}
	if !r.Match("composer.json") || !r.Match("app/composer.json") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("composer.lock") || r.Match("not-composer.json") || r.Match("composer.json.bak") {
		t.Fatal("matched a file that is not composer.json")
	}
}

// TestComposerLockWinsOverTheManifest proves the lock file's exact version
// is the one kept when both files name the package, and a package only the
// manifest names is still kept.
func TestComposerLockWinsOverTheManifest(t *testing.T) {
	files := []change.File{
		{Path: "composer.json", Change: []byte("{\"require\":{\"Monolog/monolog\":\"^2.0\",\"only/manifest\":\"1.0.0\"}}\n")},
		{Path: "composer.lock", Change: []byte("{\n  \"packages\": [\n    {\"name\": \"monolog/monolog\", \"version\": \"2.9.1\"}\n  ]\n}\n")},
	}
	got, notes, err := Collect(files, []Reader{Composer(), ComposerLock()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	// Collect sorts by file path, so the manifest hit is printed first.
	if got[0].File != "composer.json" || got[0].Package != "only/manifest" || got[0].Lock {
		t.Fatalf("%#v", got[0])
	}
	if got[1].File != "composer.lock" || got[1].Package != "monolog/monolog" || got[1].Version != "2.9.1" || !got[1].Lock {
		t.Fatalf("%#v", got[1])
	}
}
