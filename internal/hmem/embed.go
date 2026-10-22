package hmem

import "math"

// DIM is the width of the deterministic local embedding.
// It is a stand-in for a model embedding. No network call builds it.
const DIM = 64

// Tokenize lowercases text, keeps letters and digits, and drops tokens of one rune.
func Tokenize(text string) []string {
	lower := []rune(text)
	for i, r := range lower {
		if r >= 'A' && r <= 'Z' {
			lower[i] = r - 'A' + 'a'
		}
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 1 {
			out = append(out, string(cur))
		}
		cur = cur[:0]
	}
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur = append(cur, r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func fnv1a(s string) uint32 {
	h := uint32(0x811c9dc5)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 0x01000193
	}
	return h
}

// Embed is a pure function of the text: FNV-1a per token into a unit vector.
func Embed(text string) []float64 {
	vec := make([]float64, DIM)
	for _, token := range Tokenize(text) {
		h := fnv1a(token)
		idx := int(h % DIM)
		sign := 1.0
		if (h>>8)%2 != 0 {
			sign = -1
		}
		mag := 1 + float64((h>>16)%8)/8
		vec[idx] += sign * mag
	}
	var sum float64
	for _, v := range vec {
		sum += v * v
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return vec
	}
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

// Cosine is the dot product of two unit vectors. A length mismatch scores 0.
func Cosine(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

// Jaccard is the overlap of two token sets. An empty set scores 0.
func Jaccard(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	sb := make(map[string]struct{}, len(b))
	for _, x := range b {
		sb[x] = struct{}{}
	}
	seen := make(map[string]struct{}, len(a))
	inter := 0
	for _, x := range a {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		if _, ok := sb[x]; ok {
			inter++
		}
	}
	union := len(seen)
	for x := range sb {
		if _, ok := seen[x]; !ok {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}
