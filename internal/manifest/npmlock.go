package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/semver"
)

// NPMLock reads package-lock.json lockfileVersion 2 or 3.
// It reads the packages map. The ecosystem is npm.
// The root entry, and any path that is not under node_modules/, is not a
// dependency. The name is the entry's name field, or the path after the
// last node_modules/. An entry with no version string is not a pin.
func NPMLock() Reader {
	return Reader{
		Ecosystem: "npm",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "package-lock.json" || strings.HasSuffix(path, "/package-lock.json")
		},
		Diff: diffNPMLock,
	}
}

func diffNPMLock(base, change []byte) ([]Change, error) {
	before, err := parseNPMLock(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, err := parseNPMLock(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, semver.Compare), nil
}

func parseNPMLock(b []byte) (versions, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("not an object")
	}
	sawVersion := false
	version := ""
	sawPackages := false
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is not a string")
		}
		switch key {
		case "lockfileVersion":
			sawVersion = true
			version, err = lockVersion(dec)
		case "packages":
			sawPackages = true
			err = lockPackages(dec, b, out)
		default:
			err = skipLock(dec)
		}
		if err != nil {
			return nil, err
		}
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '}' {
		return nil, fmt.Errorf("object did not end")
	}
	if dec.More() {
		return nil, fmt.Errorf("trailing data")
	}
	if !sawVersion {
		return nil, fmt.Errorf("lockfileVersion is missing")
	}
	if version != "2" && version != "3" {
		return nil, fmt.Errorf("lockfileVersion %s is not 2 or 3", version)
	}
	if !sawPackages {
		return nil, fmt.Errorf("packages is missing")
	}
	return out, nil
}

func lockVersion(dec *json.Decoder) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	switch n := tok.(type) {
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64), nil
	case json.Number:
		return n.String(), nil
	default:
		return "", fmt.Errorf("lockfileVersion is not a number")
	}
}

func lockPackages(dec *json.Decoder, b []byte, out versions) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '{' {
		return fmt.Errorf("packages is not an object")
	}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
		path, ok := tok.(string)
		if !ok {
			return fmt.Errorf("packages key is not a string")
		}
		name, ver, line, err := lockOne(dec, b, path)
		if err != nil {
			return err
		}
		if name == "" || ver == "" {
			continue
		}
		addVersion(out, name, ver, line)
	}
	tok, err = dec.Token()
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '}' {
		return fmt.Errorf("packages did not end")
	}
	return nil
}

func lockOne(dec *json.Decoder, b []byte, path string) (string, string, int, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid JSON: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '{' {
		return "", "", 0, fmt.Errorf("packages entry %q is not an object", path)
	}
	var name, ver string
	line := 0
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid JSON: %w", err)
		}
		field, ok := tok.(string)
		if !ok {
			return "", "", 0, fmt.Errorf("packages entry %q has a key that is not a string", path)
		}
		switch field {
		case "name":
			name, err = lockString(dec, "name")
		case "version":
			line = lineAt(b, dec.InputOffset())
			ver, err = lockString(dec, "version")
		default:
			err = skipLock(dec)
		}
		if err != nil {
			return "", "", 0, err
		}
	}
	tok, err = dec.Token()
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '}' {
		return "", "", 0, fmt.Errorf("packages entry %q did not end", path)
	}
	if !strings.Contains(path, "node_modules/") {
		return "", "", 0, nil
	}
	if name == "" {
		name = nameFromLockPath(path)
	}
	return name, ver, line, nil
}

func nameFromLockPath(path string) string {
	const mark = "node_modules/"
	i := strings.LastIndex(path, mark)
	if i < 0 {
		return ""
	}
	return path[i+len(mark):]
}

func lockString(dec *json.Decoder, what string) (string, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	s, ok := tok.(string)
	if ok {
		return s, nil
	}
	if d, isDelim := tok.(json.Delim); isDelim && (d == '{' || d == '[') {
		if err := skipLockRest(dec, closeDelim(d)); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("%s is not a string", what)
}

func skipLock(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	if d != '{' && d != '[' {
		return fmt.Errorf("invalid JSON: bad delimiter")
	}
	return skipLockRest(dec, closeDelim(d))
}

func skipLockRest(dec *json.Decoder, end json.Delim) error {
	if end == '}' {
		for dec.More() {
			if _, err := dec.Token(); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
			if err := skipLock(dec); err != nil {
				return err
			}
		}
	} else {
		for dec.More() {
			if err := skipLock(dec); err != nil {
				return err
			}
		}
	}
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != end {
		return fmt.Errorf("invalid JSON: value did not end")
	}
	return nil
}
