package osv

import "testing"

// TestLoadIDViewEmptyIsTheDefaultAndABadNameIsRefused proves an empty setting
// is the default view, a plain name is taken, and a name that is not one is
// refused rather than replaced by the default.
func TestLoadIDViewEmptyIsTheDefaultAndABadNameIsRefused(t *testing.T) {
	got, err := LoadIDView(nil)
	if err != nil || got != IDView {
		t.Fatalf("nil env: %q, %v", got, err)
	}
	got, err = LoadIDView(func(string) string { return "  " })
	if err != nil || got != IDView {
		t.Fatalf("blank: %q, %v", got, err)
	}
	got, err = LoadIDView(func(k string) string {
		if k != EnvIDView {
			t.Fatalf("read %s, want %s", k, EnvIDView)
		}
		return "px_usl.px_advisory"
	})
	if err != nil || got != "px_usl.px_advisory" {
		t.Fatalf("plain name: %q, %v", got, err)
	}
	// Mutated field: PROVEN_OSV_ID_VIEW.
	_, err = LoadIDView(func(string) string { return "px_usl.bad;drop" })
	cfg, ok := err.(*ConfigError)
	if !ok || cfg.Field != EnvIDView || cfg.Value != "px_usl.bad;drop" {
		t.Fatalf("a bad name: %v", err)
	}
}
