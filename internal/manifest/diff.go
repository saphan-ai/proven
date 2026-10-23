package manifest

import "sort"

// diffNamed returns packages that after adds, or whose version went up.
// A pair order cannot rank is still returned: the caller traces the reason
// and does not guess which one is newer.
func diffNamed(before, after map[string]req, order func(a, b string) (int, bool)) []Change {
	var out []Change
	for path, now := range after {
		old, ok := before[path]
		if !ok {
			out = append(out, Change{Package: path, Version: now.version, Line: now.line, Added: true})
			continue
		}
		if old.version == now.version {
			continue
		}
		cmp, ordered := order(now.version, old.version)
		if ordered && cmp <= 0 {
			continue
		}
		out = append(out, Change{Package: path, Version: now.version, Line: now.line, Added: false})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Package < out[j].Package })
	return out
}
