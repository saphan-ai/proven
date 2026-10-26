package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAddStoresTheHashAndNeverTheToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	tok, err := Add(path, "judge-ada", t0.Add(time.Hour), 5, 50, t0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tok, TokenPrefix) || len(tok) < 40 {
		t.Errorf("token shape: %q", tok)
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), tok) {
		t.Error("the settings file holds the token")
	}
	if !strings.Contains(string(raw), Hash(tok)) {
		t.Error("the settings file lacks the token's hash")
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", st.Mode().Perm())
	}
	recs, err := List(path)
	if err != nil || len(recs) != 1 || recs[0].ID != "judge-ada" {
		t.Fatalf("list: %v %v", recs, err)
	}
}

func TestAddRefusals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	if _, err := Add(path, "judge-ada", t0.Add(time.Hour), 5, 50, t0); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		id      string
		exp     time.Time
		pm, pd  int
		wantErr string
	}{
		"expiry in the past":  {"judge-bob", t0.Add(-time.Second), 5, 50, "future"},
		"duplicate id":        {"judge-ada", t0.Add(time.Hour), 5, 50, "twice"},
		"upper-case id":       {"Judge", t0.Add(time.Hour), 5, 50, "id"},
		"zero per_minute":     {"judge-bob", t0.Add(time.Hour), 0, 50, "per_minute"},
		"zero per_day":        {"judge-bob", t0.Add(time.Hour), 5, 0, "per_day"},
		"id with a separator": {"a/b", t0.Add(time.Hour), 5, 50, "id"},
	} {
		_, err := Add(path, tc.id, tc.exp, tc.pm, tc.pd, t0)
		if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: %v, want an error naming %q", name, err, tc.wantErr)
		}
	}
	if recs, _ := List(path); len(recs) != 1 {
		t.Errorf("a refused token was stored: %d records", len(recs))
	}
}

func TestSettingsFileIsRefusedWholeWhenOneRuleBreaks(t *testing.T) {
	good := `{"version":1,"tokens":[{"id":"a","sha256":"` + strings.Repeat("a", 64) + `","expires":"2026-10-28T00:00:00Z","per_minute":1,"per_day":1}]}`
	if _, err := parse([]byte(good)); err != nil {
		t.Fatalf("the control file is refused: %v", err)
	}
	// Each case mutates exactly one field of the control file.
	for name, bad := range map[string]string{
		"sha256 too short":  strings.Replace(good, strings.Repeat("a", 64), "abc", 1),
		"sha256 upper case": strings.Replace(good, strings.Repeat("a", 64), strings.Repeat("A", 64), 1),
		"expires missing":   strings.Replace(good, `"expires":"2026-10-28T00:00:00Z",`, "", 1),
		"per_minute zero":   strings.Replace(good, `"per_minute":1`, `"per_minute":0`, 1),
		"version":           strings.Replace(good, `"version":1`, `"version":2`, 1),
		"unknown field":     strings.Replace(good, `"id":"a",`, `"id":"a","admin":true,`, 1),
		"revoked_at":        strings.Replace(good, `"per_day":1}`, `"per_day":1,"revoked_at":"yesterday"}`, 1),
	} {
		if _, err := parse([]byte(bad)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestRevokeNamesAnUnknownID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	if _, err := Add(path, "judge-ada", t0.Add(time.Hour), 5, 50, t0); err != nil {
		t.Fatal(err)
	}
	if err := Revoke(path, "judge-nobody", t0); err == nil {
		t.Error("revoking an unknown id succeeded")
	}
}

func TestLookupHandlesTheExpiryInstant(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tokens.json")
	exp := t0.Add(time.Hour)
	tok, _ := Add(path, "judge-ada", exp, 5, 50, t0)
	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(tok, exp.Add(-time.Nanosecond)); err != nil {
		t.Errorf("one tick before expiry: %v", err)
	}
	if _, err := s.Lookup(tok, exp); err == nil || err.(*Refusal).Code != CodeTokenExpired {
		t.Errorf("at the expiry instant: %v", err)
	}
}

func TestAlarmThresholdIsEightyPercentRoundedUp(t *testing.T) {
	for limit, wantAt := range map[int]int{1: 1, 5: 4, 10: 8, 100: 80, 3: 3} {
		l := NewLimiter()
		rec := Record{ID: "j", PerMinute: limit, PerDay: 1000}
		var at, total int
		for i := 1; i <= limit; i++ {
			alarms, ref := l.Allow(rec, t0)
			if ref != nil {
				t.Fatalf("limit %d refused call %d", limit, i)
			}
			total += len(alarms)
			if len(alarms) > 0 && at == 0 {
				at = i
			}
		}
		if at != wantAt || total != 1 {
			t.Errorf("limit %d: alarm at call %d (%d alarms), want call %d once", limit, at, total, wantAt)
		}
	}
}
