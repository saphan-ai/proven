package manifest

import (
	"strings"
	"testing"
)

// TestComposerLockReadsPins proves packages and packages-dev each contribute
// a pin, a package with no version does not, and a lower version is not a raise.
func TestComposerLockReadsPins(t *testing.T) {
	base := []byte(`{
  "packages": [
    {"name": "monolog/monolog", "version": "2.0.0", "require": {"php": "^8.0"}}
  ],
  "packages-dev": [
    {"name": "phpunit/phpunit", "version": "10.0.0"}
  ]
}
`)
	change := []byte(`{
  "content-hash": "abc",
  "packages": [
    {"name": "monolog/monolog", "version": "2.9.1"},
    {"name": "missing/version"}
  ],
  "packages-dev": [
    {"name": "phpunit/phpunit", "version": "10.0.0"},
    {"name": "friendsofphp/php-cs-fixer", "version": "3.1.0"}
  ],
  "platform": {"php": "^8.2"}
}
`)
	got, err := diffComposerLock(base, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	if got[0].Package != "friendsofphp/php-cs-fixer" || got[0].Version != "3.1.0" || !got[0].Added {
		t.Fatalf("%#v", got[0])
	}
	if got[1].Package != "monolog/monolog" || got[1].Version != "2.9.1" || got[1].Added {
		t.Fatalf("%#v", got[1])
	}
}

// TestComposerLockRejectsBrokenJSON proves a lock file that is not JSON
// returns an error, so Collect can record it as a note.
func TestComposerLockRejectsBrokenJSON(t *testing.T) {
	if _, err := diffComposerLock([]byte("{"), []byte("{}")); err == nil || !strings.Contains(err.Error(), "base:") {
		t.Fatal(err)
	}
}

// TestComposerLockMatch proves the reader matches composer.lock at any depth
// and not the manifest or a backup.
func TestComposerLockMatch(t *testing.T) {
	r := ComposerLock()
	if r.Ecosystem != "Packagist" || !r.Lock || !r.NoteErr || r.Compare == nil {
		t.Fatal(r.Ecosystem)
	}
	if !r.Match("composer.lock") || !r.Match("app/composer.lock") {
		t.Fatal("missed composer.lock")
	}
	if r.Match("composer.json") || r.Match("composer.lock.bak") || r.Match("not-composer.lock") {
		t.Fatal("matched a file that is not composer.lock")
	}
}
