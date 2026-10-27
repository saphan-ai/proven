package generator

import "testing"

// GLM serves its OpenAI-compatible API under /api/paas/v4, so a base whose path ends in a
// version is accepted and is the API root; a path that does not is still refused.
// Mutated field: PROVEN_GENERATOR_BASE_URL, the path's last segment.
func TestABaseEndingInAVersionIsTheAPIRoot(t *testing.T) {
	for base, want := range map[string]string{
		"https://api.z.ai/api/paas/v4/": "https://api.z.ai/api/paas/v4/chat/completions",
		"http://127.0.0.1:8088/v1":      "http://127.0.0.1:8088/v1/chat/completions",
		"http://127.0.0.1:8088":         "http://127.0.0.1:8088/v1/chat/completions",
	} {
		got, err := checkBase(EnvBaseURL, base, false)
		if err != nil {
			t.Fatalf("%s: %v", base, err)
		}
		if e := endpoint(got, "/chat/completions"); e != want {
			t.Errorf("%s: endpoint %s, want %s", base, e, want)
		}
	}
	for _, base := range []string{"https://api.z.ai/api/paas/v4x", "https://api.z.ai/api/paas", "http://127.0.0.1:8088/v"} {
		if _, err := checkBase(EnvBaseURL, base, false); err == nil {
			t.Errorf("%s was accepted", base)
		}
	}
}
