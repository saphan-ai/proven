// Package gateway is the judges' gateway: a small proxy on our side that lets
// a named judge's lab reach our model service and our evidence store with the
// judge's own expiring token (ADR-0051). Our two credentials never leave the
// gateway; the judge's token never goes upstream.
package gateway

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TokenPrefix starts every judge token, so a leaked one is recognisable.
const TokenPrefix = "pjt_"

// settingsVersion is the only version of the token settings file.
const settingsVersion = 1

// The refusal codes of the token check. They are what a judge sees, by name.
const (
	CodeMissingToken = "missing_token"
	CodeUnknownToken = "unknown_token"
	CodeTokenExpired = "token_expired"
	CodeTokenRevoked = "token_revoked"
	CodeSettingsDown = "gateway_settings_unreadable"
)

var idFormat = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// Refusal is a call the gateway will not serve. Code is the stable name a
// judge and a watcher can match; Detail is a fixed sentence, never the
// caller's own text.
type Refusal struct {
	Status int
	Code   string
	Detail string
}

// Error reads "<code>: <detail>".
func (r *Refusal) Error() string { return r.Code + ": " + r.Detail }

// Record is one judge token as the settings file holds it. Only the hash of
// the token is stored; the token itself is shown once, when it is made.
type Record struct {
	ID        string    `json:"id"`
	SHA256    string    `json:"sha256"`
	Expires   time.Time `json:"expires"`
	PerMinute int       `json:"per_minute"`
	PerDay    int       `json:"per_day"`
	RevokedAt string    `json:"revoked_at,omitempty"`
}

// file is the settings file: one JSON object.
type file struct {
	Version int      `json:"version"`
	Tokens  []Record `json:"tokens"`
}

// Hash is the SHA-256 of a token, in lower-case hex.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewToken returns a new random token: 32 bytes from the system source.
func NewToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// validate holds the whole file to its rules. A file that breaks one is
// refused whole: a half-read file could leave a revoked token alive.
func (f *file) validate() error {
	if f.Version != settingsVersion {
		return fmt.Errorf("version must be %d", settingsVersion)
	}
	seen := map[string]bool{}
	for i, r := range f.Tokens {
		switch {
		case !idFormat.MatchString(r.ID):
			return fmt.Errorf("token %d: id is not 1 to 40 lower-case letters, digits or hyphens", i)
		case seen[r.ID]:
			return fmt.Errorf("token %q: id is used twice", r.ID)
		case len(r.SHA256) != 64 || strings.Trim(r.SHA256, "0123456789abcdef") != "":
			return fmt.Errorf("token %q: sha256 is not 64 lower-case hex digits", r.ID)
		case r.Expires.IsZero():
			return fmt.Errorf("token %q: expires is required; no token lasts for ever", r.ID)
		case r.PerMinute < 1 || r.PerDay < 1:
			return fmt.Errorf("token %q: per_minute and per_day must be at least 1", r.ID)
		}
		if r.RevokedAt != "" {
			if _, err := time.Parse(time.RFC3339, r.RevokedAt); err != nil {
				return fmt.Errorf("token %q: revoked_at is not an RFC 3339 time", r.ID)
			}
		}
		seen[r.ID] = true
	}
	return nil
}

// Store reads the settings file and reloads it when it changes, so a
// revocation takes effect on the next call without a restart. A file that
// cannot be read or fails its rules closes the gateway: every call is refused
// by name until the file is right again.
type Store struct {
	path string

	mu    sync.Mutex
	stamp time.Time
	size  int64
	have  bool
	recs  []Record
	err   error
}

// NewStore opens the settings file. The first read must succeed: a gateway
// that cannot read its tokens does not start.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.reload(); err != nil {
		return nil, err
	}
	return s, nil
}

// reload reads the file now. The caller holds no lock.
func (s *Store) reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *Store) readLocked() error {
	st, err := os.Stat(s.path)
	if err != nil {
		s.recs, s.err, s.have = nil, fmt.Errorf("settings file: %w", pathless(err)), false
		return s.err
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		s.recs, s.err, s.have = nil, fmt.Errorf("settings file: %w", pathless(err)), false
		return s.err
	}
	f, err := parse(raw)
	if err != nil {
		s.recs, s.err, s.have = nil, err, false
		return err
	}
	s.recs, s.err, s.have = f.Tokens, nil, true
	s.stamp, s.size = st.ModTime(), st.Size()
	return nil
}

// pathless drops the file path from an operating-system error.
func pathless(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

func parse(raw []byte) (*file, error) {
	var f file
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("settings file is not the token settings format: %v", err)
	}
	if dec.More() {
		return nil, errors.New("settings file holds more than one JSON value")
	}
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("settings file refused: %v", err)
	}
	return &f, nil
}

// refresh reloads when the file's time or size changed since the last read,
// and reports the standing read error.
func (s *Store) refresh() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := os.Stat(s.path)
	if err != nil {
		return s.readLocked()
	}
	if !s.have || !st.ModTime().Equal(s.stamp) || st.Size() != s.size {
		return s.readLocked()
	}
	return s.err
}

// Lookup finds the record whose hash matches the token and checks it: not
// revoked, not expired. It never says which part of a wrong token was wrong.
func (s *Store) Lookup(token string, now time.Time) (Record, error) {
	if err := s.refresh(); err != nil {
		return Record{}, &Refusal{Status: 503, Code: CodeSettingsDown, Detail: "the gateway cannot read its token settings; tell the owner"}
	}
	s.mu.Lock()
	recs := s.recs
	s.mu.Unlock()
	want := []byte(Hash(token))
	var hit *Record
	// Every record is compared, in constant time, so the answer's timing does
	// not say how many hash characters matched.
	for i := range recs {
		if subtle.ConstantTimeCompare(want, []byte(recs[i].SHA256)) == 1 {
			hit = &recs[i]
		}
	}
	if hit == nil {
		return Record{}, &Refusal{Status: 401, Code: CodeUnknownToken, Detail: "this token is not one the gateway issued"}
	}
	if hit.RevokedAt != "" {
		return *hit, &Refusal{Status: 401, Code: CodeTokenRevoked, Detail: "this token was revoked"}
	}
	if !now.Before(hit.Expires) {
		return *hit, &Refusal{Status: 401, Code: CodeTokenExpired, Detail: "this token expired at " + hit.Expires.UTC().Format(time.RFC3339)}
	}
	return *hit, nil
}

// ---- the owner's side: make, revoke and list, on the file itself ----

// Add makes a token and stores its hash. The token is returned once and is
// not kept anywhere.
func Add(path, id string, expires time.Time, perMinute, perDay int, now time.Time) (string, error) {
	f, err := load(path, true)
	if err != nil {
		return "", err
	}
	if !expires.After(now) {
		return "", errors.New("expires must be in the future")
	}
	tok, err := NewToken()
	if err != nil {
		return "", err
	}
	f.Tokens = append(f.Tokens, Record{ID: id, SHA256: Hash(tok), Expires: expires.UTC(), PerMinute: perMinute, PerDay: perDay})
	if err := f.validate(); err != nil {
		return "", err
	}
	return tok, save(path, f)
}

// Revoke marks one token revoked. The others are untouched. A token already
// revoked keeps its first revocation time.
func Revoke(path, id string, now time.Time) error {
	f, err := load(path, false)
	if err != nil {
		return err
	}
	for i := range f.Tokens {
		if f.Tokens[i].ID == id {
			if f.Tokens[i].RevokedAt == "" {
				f.Tokens[i].RevokedAt = now.UTC().Format(time.RFC3339)
			}
			return save(path, f)
		}
	}
	return fmt.Errorf("no token with id %q", id)
}

// List returns the records. It holds no token and no secret: the hash is not
// a token.
func List(path string) ([]Record, error) {
	f, err := load(path, false)
	if err != nil {
		return nil, err
	}
	return f.Tokens, nil
}

func load(path string, create bool) (*file, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && create {
		return &file{Version: settingsVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("settings file: %w", pathless(err))
	}
	return parse(raw)
}

// save writes the file whole to a new name and renames it over the old one,
// so a running gateway never reads a half-written file. Mode 0600.
func save(path string, f *file) error {
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".new"
	if err := os.WriteFile(tmp, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("settings file: %w", pathless(err))
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("settings file: %w", pathless(err))
	}
	return nil
}
