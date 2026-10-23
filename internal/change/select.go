package change

import (
	"fmt"
	"path"
	"path/filepath"
)

// Select keeps files that match an include pattern and no exclude pattern.
// An empty include list matches every file. Exclude wins. A pattern matches
// the slash path or its base name, using path.Match. out holds every file that
// was not selected, so the caller can say what the review left out.
func Select(files, include, exclude []string) (selected, out []string, err error) {
	for _, f := range files {
		in, err := anyMatch(include, f, true)
		if err != nil {
			return nil, nil, err
		}
		ex, err := anyMatch(exclude, f, false)
		if err != nil {
			return nil, nil, err
		}
		if in && !ex {
			selected = append(selected, f)
			continue
		}
		out = append(out, f)
	}
	return selected, out, nil
}

// anyMatch reports whether file matches one of patterns. empty is the answer
// for an empty list: true for include (everything is in), false for exclude
// (nothing is out).
func anyMatch(patterns []string, file string, empty bool) (bool, error) {
	if len(patterns) == 0 {
		return empty, nil
	}
	for _, p := range patterns {
		ok, err := match(p, file)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// match tries pattern on the whole slash path, then on the base name, so a
// pattern such as "go.mod" also matches a go.mod in a subdirectory. A bad
// pattern is an error that names it, not a silent miss.
func match(pattern, file string) (bool, error) {
	file = filepath.ToSlash(file)
	for _, candidate := range []string{file, path.Base(file)} {
		ok, err := path.Match(pattern, candidate)
		if err != nil {
			return false, fmt.Errorf("glob %q: %w", pattern, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
