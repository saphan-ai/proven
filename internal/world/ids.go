package world

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

// NormalizeName folds case, diacritics and whitespace, so two spellings of
// one name give one entity id. The meterless-conform command exposes it to
// the World Model conformance suite.
// The fold covers Latin-1 letters and combining marks, which is the set the
// conformance vector uses. It does not abbreviate or stem.
func NormalizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		// Each listed Unicode space becomes a plain space, and zero-width characters
		// are dropped, so invisible differences do not split one name in two.
		switch r {
		case '\u00a0', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005',
			'\u2006', '\u2007', '\u2008', '\u2009', '\u200a', '\u202f', '\u205f', '\u3000':
			b.WriteByte(' ')
			continue
		}
		if r >= 0x200b && r <= 0x200d || r == 0xfeff {
			continue
		}
		for _, part := range decompose(r) {
			if unicode.Is(unicode.Mn, part) {
				continue
			}
			b.WriteRune(part)
		}
	}
	return strings.Join(strings.Fields(strings.ToLower(b.String())), " ")
}

// decompose maps a Latin-1 letter to its base and combining mark.
// A rune with no entry is itself.
func decompose(r rune) []rune {
	if base, ok := latin1[r]; ok {
		return base
	}
	return []rune{r}
}

// EntityIDFromName is ent_ plus the first 16 hex digits of sha256(type + ":" + normalized name).
// The type is part of the hash, so a person and an organisation of the same name stay apart.
func EntityIDFromName(typ, name string) string {
	sum := sha256.Sum256([]byte(typ + ":" + NormalizeName(name)))
	return "ent_" + hex.EncodeToString(sum[:])[:16]
}

// ExternalID is the literal grammar type:system:externalId. The world store
// names a finding entity this way, with the system "review".
func ExternalID(typ, system, id string) string {
	return typ + ":" + system + ":" + id
}

// Decide bands a similarity. 0.85 and above auto-merges. 0.82 up to 0.85 queues.
// Below 0.82 is a new entity. The bands come from the World Model
// specification and its conformance vectors, not from a setting.
func Decide(similarity float64) string {
	if similarity >= 0.85 {
		return "auto-merge"
	}
	if similarity >= 0.82 {
		return "review"
	}
	return "new"
}

// latin1 is the NFKD base of the Latin-1 supplement letters the runner uses.
var latin1 = map[rune][]rune{
	'À': {'a', '\u0300'}, 'Á': {'a', '\u0301'}, 'Â': {'a', '\u0302'}, 'Ã': {'a', '\u0303'},
	'Ä': {'a', '\u0308'}, 'Å': {'a', '\u030a'}, 'Ç': {'c', '\u0327'},
	'È': {'e', '\u0300'}, 'É': {'e', '\u0301'}, 'Ê': {'e', '\u0302'}, 'Ë': {'e', '\u0308'},
	'Ì': {'i', '\u0300'}, 'Í': {'i', '\u0301'}, 'Î': {'i', '\u0302'}, 'Ï': {'i', '\u0308'},
	'Ñ': {'n', '\u0303'},
	'Ò': {'o', '\u0300'}, 'Ó': {'o', '\u0301'}, 'Ô': {'o', '\u0302'}, 'Õ': {'o', '\u0303'}, 'Ö': {'o', '\u0308'},
	'Ù': {'u', '\u0300'}, 'Ú': {'u', '\u0301'}, 'Û': {'u', '\u0302'}, 'Ü': {'u', '\u0308'},
	'Ý': {'y', '\u0301'},
	'à': {'a', '\u0300'}, 'á': {'a', '\u0301'}, 'â': {'a', '\u0302'}, 'ã': {'a', '\u0303'},
	'ä': {'a', '\u0308'}, 'å': {'a', '\u030a'}, 'ç': {'c', '\u0327'},
	'è': {'e', '\u0300'}, 'é': {'e', '\u0301'}, 'ê': {'e', '\u0302'}, 'ë': {'e', '\u0308'},
	'ì': {'i', '\u0300'}, 'í': {'i', '\u0301'}, 'î': {'i', '\u0302'}, 'ï': {'i', '\u0308'},
	'ñ': {'n', '\u0303'},
	'ò': {'o', '\u0300'}, 'ó': {'o', '\u0301'}, 'ô': {'o', '\u0302'}, 'õ': {'o', '\u0303'}, 'ö': {'o', '\u0308'},
	'ù': {'u', '\u0300'}, 'ú': {'u', '\u0301'}, 'û': {'u', '\u0302'}, 'ü': {'u', '\u0308'},
	'ý': {'y', '\u0301'}, 'ÿ': {'y', '\u0308'},
}
