package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/semver"
)

// NPM reads package.json dependencies and devDependencies.
// The version string is kept as written. Comparison is semantic version.
func NPM() Reader {
	return Reader{
		Ecosystem: "npm",
		Match: func(path string) bool {
			return path == "package.json" || strings.HasSuffix(path, "/package.json")
		},
		Diff:  diffNPM,
		Notes: npmNotes,
	}
}

func diffNPM(base, change []byte) ([]Change, error) {
	before, _, err := parseNPM(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseNPM(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, semver.Compare), nil
}

func npmNotes(change []byte) []string {
	_, notes, err := parseNPM(change)
	if err != nil {
		return nil
	}
	return notes
}

func parseNPM(b []byte) (map[string]req, []string, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return map[string]req{}, nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("package.json: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, nil, fmt.Errorf("package.json: not an object")
	}
	out := map[string]req{}
	var notes []string
	if err := readObject(dec, b, func(key string, _ int) error {
		if key != "dependencies" && key != "devDependencies" {
			return skipValue(dec)
		}
		return readSection(dec, b, key, out, &notes)
	}); err != nil {
		return nil, nil, err
	}
	if dec.More() {
		return nil, nil, fmt.Errorf("package.json: trailing data")
	}
	return out, notes, nil
}

func readSection(dec *json.Decoder, b []byte, section string, out map[string]req, notes *[]string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '{' {
		if ok && d == '[' {
			if err := skipRest(dec, ']'); err != nil {
				return err
			}
		}
		*notes = append(*notes, section+" is not an object")
		return nil
	}
	return readObject(dec, b, func(pkg string, line int) error {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("package.json: %w", err)
		}
		if s, ok := tok.(string); ok {
			out[pkg] = req{version: s, line: line}
			return nil
		}
		if d, ok := tok.(json.Delim); ok && (d == '{' || d == '[') {
			if err := skipRest(dec, closeDelim(d)); err != nil {
				return err
			}
		}
		*notes = append(*notes, section+"."+pkg+" is not a string")
		return nil
	})
}

func readObject(dec *json.Decoder, b []byte, fn func(key string, line int) error) error {
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("package.json: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return fmt.Errorf("package.json: object key is not a string")
		}
		if err := fn(key, lineAt(b, dec.InputOffset())); err != nil {
			return err
		}
	}
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '}' {
		return fmt.Errorf("package.json: object did not end")
	}
	return nil
}

func skipValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	if d != '{' && d != '[' {
		return fmt.Errorf("package.json: bad delimiter")
	}
	return skipRest(dec, closeDelim(d))
}

func skipRest(dec *json.Decoder, end json.Delim) error {
	if end == '}' {
		for dec.More() {
			if _, err := dec.Token(); err != nil {
				return fmt.Errorf("package.json: %w", err)
			}
			if err := skipValue(dec); err != nil {
				return err
			}
		}
	} else {
		for dec.More() {
			if err := skipValue(dec); err != nil {
				return err
			}
		}
	}
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("package.json: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != end {
		return fmt.Errorf("package.json: value did not end")
	}
	return nil
}

func closeDelim(d json.Delim) json.Delim {
	if d == '[' {
		return ']'
	}
	return '}'
}

func lineAt(b []byte, offset int64) int {
	if offset < 0 {
		return 1
	}
	n := int(offset)
	if n > len(b) {
		n = len(b)
	}
	return 1 + bytes.Count(b[:n], []byte("\n"))
}
