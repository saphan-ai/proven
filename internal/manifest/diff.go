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
		// The same string is no change. "v1.2.3" against "1.2.3" goes on to
		// the comparison below, which treats them as equal.
		if old.version == now.version {
			continue
		}
		cmp, ordered := order(now.version, old.version)
		// Only a lowered or equal version is dropped. A pair that does not
		// order (a range such as ^1.2.3, or "latest") is kept: the OSV check
		// then writes why it could not compare it, instead of staying silent.
		if ordered && cmp <= 0 {
			continue
		}
		out = append(out, Change{Package: path, Version: now.version, Line: now.line, Added: false})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Package < out[j].Package })
	return out
}

// versions is package → version → line. A lock file can pin one package
// at more than one version. Each new version is its own change.
type versions map[string]map[string]int

// addVersion records that pkg is pinned at ver on line.
func addVersion(dst versions, pkg, ver string, line int) {
	if dst[pkg] == nil {
		dst[pkg] = map[string]int{}
	}
	dst[pkg][ver] = line
}

// diffVersions returns versions after introduces.
// A package that is new reports every version as added.
// A package that was already pinned reports a version only when it is
// higher than every version on the base, or when the two do not order.
// A lower or equal version is not a raise.
func diffVersions(before, after versions, order func(a, b string) (int, bool)) []Change {
	var out []Change
	for pkg, now := range after {
		old := before[pkg]
		for ver, line := range now {
			if _, ok := old[ver]; ok {
				continue
			}
			if len(old) == 0 {
				out = append(out, Change{Package: pkg, Version: ver, Line: line, Added: true})
				continue
			}
			if versionRaised(ver, old, order) {
				out = append(out, Change{Package: pkg, Version: ver, Line: line, Added: false})
			}
		}
	}
	// Map order is random in Go; the sort makes the output stable.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Package != out[j].Package {
			return out[i].Package < out[j].Package
		}
		if out[i].Version != out[j].Version {
			return out[i].Version < out[j].Version
		}
		return out[i].Line < out[j].Line
	})
	return out
}

// versionRaised reports whether ver is higher than every base version it
// can be ordered against. A base version that does not order is passed
// over, so one odd pin on the base does not hide a real raise.
func versionRaised(ver string, old map[string]int, order func(a, b string) (int, bool)) bool {
	for prev := range old {
		cmp, ordered := order(ver, prev)
		if !ordered {
			continue
		}
		if cmp <= 0 {
			return false
		}
	}
	return true
}
