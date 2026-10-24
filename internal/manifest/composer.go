package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/composerver"
)

// Composer reads composer.json require and require-dev.
// The version string is kept as written. Comparison is the Composer subset
// in composerver. A platform package is not a Packagist package: it is
// named in Notes and not checked. The package name is stored in lower case,
// which is the form Packagist and the OSV table share.
func Composer() Reader {
	return Reader{
		Ecosystem: "Packagist",
		Match: func(path string) bool {
			return path == "composer.json" || strings.HasSuffix(path, "/composer.json")
		},
		Diff:    diffComposer,
		Notes:   composerNotes,
		Compare: composerver.Compare,
		Explain: composerver.Explain,
	}
}

// diffComposer compares the requirements of two composer.json files.
func diffComposer(base, change []byte) ([]Change, error) {
	before, _, err := parseComposer(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseComposer(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, composerver.Compare), nil
}

// composerNotes names the entries of the change side that were not packages.
func composerNotes(change []byte) []string {
	_, notes, err := parseComposer(change)
	if err != nil {
		return nil
	}
	return notes
}

// parseComposer reads require and require-dev. The later line wins when a
// package is named twice. Other keys are not requirements.
func parseComposer(b []byte) (map[string]req, []string, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return map[string]req{}, nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("composer.json: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, nil, fmt.Errorf("composer.json: not an object")
	}
	out := map[string]req{}
	var notes []string
	for dec.More() {
		key, _, err := jsonKey(dec, b, "composer.json")
		if err != nil {
			return nil, nil, err
		}
		// Only the two requirement objects are packages. The name of the
		// project itself, and scripts, are not dependencies.
		if key != "require" && key != "require-dev" {
			if err := jsonSkip(dec, "composer.json"); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := composerSection(dec, b, key, out, &notes); err != nil {
			return nil, nil, err
		}
	}
	if _, err := endObject(dec, "composer.json"); err != nil {
		return nil, nil, err
	}
	if dec.More() {
		return nil, nil, fmt.Errorf("composer.json: trailing data")
	}
	return out, notes, nil
}

// composerSection reads one require object. A value that is not a string,
// and a platform package, are notes: the other requirements are still checked.
func composerSection(dec *json.Decoder, b []byte, section string, out map[string]req, notes *[]string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("composer.json: %w", err)
	}
	d, ok := tok.(json.Delim)
	if !ok || d != '{' {
		if ok && d == '[' {
			if err := jsonSkipRest(dec, ']', "composer.json"); err != nil {
				return err
			}
		}
		*notes = append(*notes, section+" is not an object")
		return nil
	}
	for dec.More() {
		name, line, err := jsonKey(dec, b, "composer.json")
		if err != nil {
			return err
		}
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("composer.json: %w", err)
		}
		ver, isString := tok.(string)
		if !isString {
			if d, ok := tok.(json.Delim); ok && (d == '{' || d == '[') {
				if err := jsonSkipRest(dec, closeDelim(d), "composer.json"); err != nil {
					return err
				}
			}
			*notes = append(*notes, section+"."+name+" is not a string")
			continue
		}
		// php, hhvm, ext-*, lib-* and the composer platform APIs are not
		// Packagist packages. Querying them would look up the wrong name.
		if platformPackage(name) {
			*notes = append(*notes, section+"."+name+" is a platform package, not a Packagist package")
			continue
		}
		pkg := strings.ToLower(strings.TrimSpace(name))
		if pkg == "" {
			*notes = append(*notes, section+" has an empty package name")
			continue
		}
		out[pkg] = req{version: ver, line: line}
	}
	_, err = endObject(dec, "composer.json")
	return err
}

// platformPackage reports a Composer platform requirement.
// These names are satisfied by the PHP runtime, not by Packagist.
func platformPackage(name string) bool {
	low := strings.ToLower(strings.TrimSpace(name))
	switch low {
	case "php", "hhvm", "composer-plugin-api", "composer-runtime-api":
		return true
	}
	return strings.HasPrefix(low, "ext-") || strings.HasPrefix(low, "lib-") || strings.HasPrefix(low, "php-")
}

// jsonKey reads one object key and the line it ends on.
func jsonKey(dec *json.Decoder, b []byte, file string) (string, int, error) {
	tok, err := dec.Token()
	if err != nil {
		return "", 0, fmt.Errorf("%s: %w", file, err)
	}
	key, ok := tok.(string)
	if !ok {
		return "", 0, fmt.Errorf("%s: object key is not a string", file)
	}
	return key, lineAt(b, dec.InputOffset()), nil
}

// endObject reads the closing brace of the object the decoder is inside.
func endObject(dec *json.Decoder, file string) (json.Token, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '}' {
		return nil, fmt.Errorf("%s: object did not end", file)
	}
	return tok, nil
}

// jsonSkip consumes one JSON value of any shape.
func jsonSkip(dec *json.Decoder, file string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	if d != '{' && d != '[' {
		return fmt.Errorf("%s: bad delimiter", file)
	}
	return jsonSkipRest(dec, closeDelim(d), file)
}

// jsonSkipRest consumes the rest of an object or array whose opening
// delimiter was already read, up to and including end.
func jsonSkipRest(dec *json.Decoder, end json.Delim, file string) error {
	if end == '}' {
		for dec.More() {
			if _, err := dec.Token(); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			if err := jsonSkip(dec, file); err != nil {
				return err
			}
		}
	} else {
		for dec.More() {
			if err := jsonSkip(dec, file); err != nil {
				return err
			}
		}
	}
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	if d, ok := tok.(json.Delim); !ok || d != end {
		return fmt.Errorf("%s: value did not end", file)
	}
	return nil
}
