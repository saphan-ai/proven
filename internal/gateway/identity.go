package gateway

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// The classes of an identity. A judge signs in with an e-mail one-time PIN; the
// runner is the GitHub workflow, which presents an Access service token.
const (
	ClassJudge  = "judge"
	ClassRunner = "runner"
)

// The refusal codes of the identity check.
const (
	CodeIdentityNotListed = "identity_not_listed"
	CodeIdentityExpired   = "identity_expired"
	CodeIdentityRevoked   = "identity_revoked"
)

var (
	emailFormat    = regexp.MustCompile(`^[a-z0-9._%+-]{1,64}@[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)
	domainFormat   = regexp.MustCompile(`^\*@[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)
	clientIDFormat = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
)

// The three shapes a Match takes.
const (
	kindEmail  = "email"
	kindDomain = "domain"
	kindClient = "client"
)

// Identity is one row of the allowlist: who may pass the Access check and be
// served. Match is an e-mail, an e-mail domain written "*@domain", or a service
// token's client id. A domain row admits every address of that domain, each
// limited on its own; an exact row beats a domain row, so one address of a
// domain is cut off by an exact row that is revoked.
type Identity struct {
	Match     string    `json:"match"`
	Class     string    `json:"class"`
	Expires   time.Time `json:"expires"`
	PerMinute int       `json:"per_minute"`
	PerDay    int       `json:"per_day"`
	RevokedAt string    `json:"revoked_at,omitempty"`
}

// Resolved is the identity of one call: the name the limits and the log use,
// its class and its limits.
type Resolved struct {
	Name      string // the e-mail or client id of the caller, never a secret
	Class     string
	PerMinute int
	PerDay    int
}

// MatchKind says which shape a Match has, or "" when it has none.
func MatchKind(m string) string {
	switch {
	case emailFormat.MatchString(m):
		return kindEmail
	case domainFormat.MatchString(m):
		return kindDomain
	case !strings.Contains(m, "@") && clientIDFormat.MatchString(m):
		return kindClient
	}
	return ""
}

// ClassOf is the only class a Match may carry: an address or a domain is a
// judge, a client id is the runner.
func ClassOf(m string) string {
	if MatchKind(m) == kindClient {
		return ClassRunner
	}
	return ClassJudge
}

func (f *file) validateIdentities() error {
	seen := map[string]bool{}
	for i, r := range f.Identities {
		kind := MatchKind(r.Match)
		switch {
		case kind == "":
			return fmt.Errorf("identity %d: match is not a lower-case e-mail, a *@domain or a service token's client id", i)
		case seen[r.Match]:
			return fmt.Errorf("identity %q: used twice", r.Match)
		case r.Class != ClassOf(r.Match):
			return fmt.Errorf("identity %q: class must be %q", r.Match, ClassOf(r.Match))
		case r.Expires.IsZero():
			return fmt.Errorf("identity %q: expires is required; no identity lasts for ever", r.Match)
		case r.PerMinute < 1 || r.PerDay < 1:
			return fmt.Errorf("identity %q: per_minute and per_day must be at least 1", r.Match)
		}
		if r.RevokedAt != "" {
			if _, err := time.Parse(time.RFC3339, r.RevokedAt); err != nil {
				return fmt.Errorf("identity %q: revoked_at is not an RFC 3339 time", r.Match)
			}
		}
		seen[r.Match] = true
	}
	return nil
}

// LookupIdentity finds the row for a subject the Access check named, and checks
// it: not revoked, not expired. An exact row wins over a domain row. A subject
// with no row is refused by name even though Access let it through.
func (s *Store) LookupIdentity(sub Subject, now time.Time) (Resolved, error) {
	if err := s.refresh(); err != nil {
		return Resolved{}, &Refusal{Status: 503, Code: CodeSettingsDown, Detail: "the gateway cannot read its settings; tell the owner"}
	}
	s.mu.Lock()
	rows := s.ids
	s.mu.Unlock()

	name := sub.Name()
	res := Resolved{Name: name}
	var hit *Identity
	switch {
	case sub.Email != "":
		if MatchKind(name) != kindEmail {
			return res, notListed()
		}
		domain := "*@" + name[strings.LastIndex(name, "@")+1:]
		for i := range rows {
			switch rows[i].Match {
			case name:
				hit = &rows[i]
			case domain:
				if hit == nil || hit.Match != name {
					hit = &rows[i]
				}
			}
		}
	default:
		if MatchKind(name) != kindClient {
			return res, notListed()
		}
		for i := range rows {
			if rows[i].Match == name {
				hit = &rows[i]
			}
		}
	}
	if hit == nil {
		return res, notListed()
	}
	res.Class, res.PerMinute, res.PerDay = hit.Class, hit.PerMinute, hit.PerDay
	if hit.RevokedAt != "" {
		return res, &Refusal{Status: 401, Code: CodeIdentityRevoked, Detail: "this identity was revoked"}
	}
	if !now.Before(hit.Expires) {
		return res, &Refusal{Status: 401, Code: CodeIdentityExpired, Detail: "this identity expired at " + hit.Expires.UTC().Format(time.RFC3339)}
	}
	return res, nil
}

func notListed() *Refusal {
	return &Refusal{Status: 403, Code: CodeIdentityNotListed, Detail: "you passed Access, but the gateway does not list you; ask the owner"}
}

// ---- the owner's side: add, revoke and list identities on the file ----

// AddIdentity adds one row. The match is lower-cased unless it is a client id.
func AddIdentity(path, match string, expires time.Time, perMinute, perDay int, now time.Time) error {
	if kind := MatchKind(strings.ToLower(match)); kind == kindEmail || kind == kindDomain {
		match = strings.ToLower(match)
	}
	f, err := load(path, true)
	if err != nil {
		return err
	}
	if !expires.After(now) {
		return errors.New("expires must be in the future")
	}
	f.Identities = append(f.Identities, Identity{Match: match, Class: ClassOf(match), Expires: expires.UTC(), PerMinute: perMinute, PerDay: perDay})
	if err := f.validate(); err != nil {
		return err
	}
	return save(path, f)
}

// RevokeIdentity marks one row revoked; the others are untouched. To cut one
// address of a domain row, add an exact row for it and revoke that.
func RevokeIdentity(path, match string, now time.Time) error {
	f, err := load(path, false)
	if err != nil {
		return err
	}
	if kind := MatchKind(strings.ToLower(match)); kind == kindEmail || kind == kindDomain {
		match = strings.ToLower(match)
	}
	for i := range f.Identities {
		if f.Identities[i].Match == match {
			if f.Identities[i].RevokedAt == "" {
				f.Identities[i].RevokedAt = now.UTC().Format(time.RFC3339)
			}
			return save(path, f)
		}
	}
	return fmt.Errorf("no identity with match %q", match)
}

// ListIdentities returns the rows.
func ListIdentities(path string) ([]Identity, error) {
	f, err := load(path, false)
	if err != nil {
		return nil, err
	}
	return f.Identities, nil
}
