package review_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestSampleOpenRefusesAWrongNetworkModeByName: sample mode loads PROVEN_NETWORK to decide whether git
// may fetch, so a value it cannot read stops Open and names the setting.
// Mutated field: PROVEN_NETWORK (a value that is neither public nor local).
func TestSampleOpenRefusesAWrongNetworkModeByName(t *testing.T) {
	env := map[string]string{"PROVEN_NETWORK": "lokal"}
	_, err := review.Open(review.Options{Sample: true, LogDir: t.TempDir(), Env: func(k string) string { return env[k] }})
	var ce *review.ConfigError
	if !errors.As(err, &ce) || ce.Field != "PROVEN_NETWORK" || !strings.Contains(ce.Error(), "lokal") {
		t.Fatalf("sample mode opened with PROVEN_NETWORK=lokal, or refused it without the name: %v", err)
	}
}
