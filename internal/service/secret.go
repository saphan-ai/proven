package service

import "strings"

// MinSecretBytes is the shortest configured secret a settings loader accepts
// (ADR-0013: a wrong non-empty value is refused). The redactor replaces every
// secret it is given with no length floor, so a secret shorter than this is
// refused at load rather than left to be redacted out of ordinary text.
const MinSecretBytes = 8

// TooShortSecret reports a non-empty secret, after trimming white space, that
// is shorter than MinSecretBytes. An empty value is not this case: an empty
// secret is a service that is down, not a wrong value.
func TooShortSecret(v string) bool {
	t := strings.TrimSpace(v)
	return t != "" && len(t) < MinSecretBytes
}
