package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/composerver"
)

// ComposerLock reads composer.lock. The packages and packages-dev arrays
// contribute each object's name and version. The ecosystem is Packagist.
// The name is stored in lower case. An object with no version string is
// not a pin. A platform package is named in Notes and not checked.
func ComposerLock() Reader {
	return Reader{
		Ecosystem: "Packagist",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "composer.lock" || strings.HasSuffix(path, "/composer.lock")
		},
		Diff:    diffComposerLock,
		Notes:   composerLockNotes,
		Compare: composerver.Compare,
		Explain: composerver.Explain,
	}
}

// diffComposerLock compares two composer.lock files. The error names the
// side that did not parse, because Collect writes it into the trace as it is.
func diffComposerLock(base, change []byte) ([]Change, error) {
	before, _, err := parseComposerLock(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, _, err := parseComposerLock(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, composerver.Compare), nil
}

// composerLockNotes names platform packages on the change side.
func composerLockNotes(change []byte) []string {
	_, notes, err := parseComposerLock(change)
	if err != nil {
		return nil
	}
	return notes
}

// parseComposerLock reads the two package arrays. Other root keys, including
// the platform object and the content hash, are not pins.
func parseComposerLock(b []byte) (versions, []string, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("composer.lock: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, nil, fmt.Errorf("composer.lock: not an object")
	}
	var notes []string
	for dec.More() {
		key, _, err := jsonKey(dec, b, "composer.lock")
		if err != nil {
			return nil, nil, err
		}
		if key != "packages" && key != "packages-dev" {
			if err := jsonSkip(dec, "composer.lock"); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := composerPackages(dec, b, key, out, &notes); err != nil {
			return nil, nil, err
		}
	}
	if _, err := endObject(dec, "composer.lock"); err != nil {
		return nil, nil, err
	}
	if dec.More() {
		return nil, nil, fmt.Errorf("composer.lock: trailing data")
	}
	return out, notes, nil
}

// composerPackages reads one array of package objects.
func composerPackages(dec *json.Decoder, b []byte, key string, out versions, notes *[]string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("composer.lock: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '[' {
		return fmt.Errorf("composer.lock: %s is not an array", key)
	}
	for dec.More() {
		if err := oneLockedPackage(dec, b, out, notes); err != nil {
			return err
		}
	}
	tok, err = dec.Token()
	if err != nil {
		return fmt.Errorf("composer.lock: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != ']' {
		return fmt.Errorf("composer.lock: %s did not end", key)
	}
	return nil
}

// oneLockedPackage reads one object. The line of the version is the line a
// finding points at, because that is the pin the change raised.
func oneLockedPackage(dec *json.Decoder, b []byte, out versions, notes *[]string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("composer.lock: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '{' {
		return fmt.Errorf("composer.lock: a package is not an object")
	}
	var name, ver string
	var line int
	var hasVer bool
	for dec.More() {
		key, keyLine, err := jsonKey(dec, b, "composer.lock")
		if err != nil {
			return err
		}
		if key != "name" && key != "version" {
			if err := jsonSkip(dec, "composer.lock"); err != nil {
				return err
			}
			continue
		}
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("composer.lock: %w", err)
		}
		s, isString := tok.(string)
		if !isString {
			if d, ok := tok.(json.Delim); ok && (d == '{' || d == '[') {
				if err := jsonSkipRest(dec, closeDelim(d), "composer.lock"); err != nil {
					return err
				}
			}
			continue
		}
		if key == "name" {
			name = s
			continue
		}
		ver = s
		line = keyLine
		hasVer = true
	}
	if _, err := endObject(dec, "composer.lock"); err != nil {
		return err
	}
	if !hasVer || strings.TrimSpace(name) == "" {
		return nil
	}
	if platformPackage(name) {
		*notes = append(*notes, "locked "+name+" is a platform package, not a Packagist package")
		return nil
	}
	addVersion(out, strings.ToLower(strings.TrimSpace(name)), ver, line)
	return nil
}
