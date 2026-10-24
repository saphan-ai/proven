package network_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

// refusedField reads the setting and the text of a refusal of any of the four packages.
func refusedField(t *testing.T, err error) (field, text string) {
	t.Helper()
	var g *generator.ConfigError
	var r *guardrail.ConfigError
	var q *query.ConfigError
	var c *testrun.ConfigError
	switch {
	case errors.As(err, &g):
		return g.Field, g.Error()
	case errors.As(err, &r):
		return r.Field, r.Error()
	case errors.As(err, &q):
		return q.Field, q.Error()
	case errors.As(err, &c):
		return c.Field, c.Error()
	}
	t.Fatalf("not a configuration refusal: %v", err)
	return "", ""
}

// TestLocalRefusesAPublicAddressInEachAddressSetting proves ADR-0034 point 1: with PROVEN_NETWORK=local, a
// public address in each of the three address settings is refused at load, and each refusal names the
// setting and the host. The same addresses load in public mode, so the mode is what refuses them.
// Mutated field: the host of the one address under test.
func TestLocalRefusesAPublicAddressInEachAddressSetting(t *testing.T) {
	local := map[string]string{
		"PROVEN_NETWORK":          "local",
		"METERLESS_RUNTIME_TOKEN": "t",
		"PROVEN_GUARDRAIL_KEY":    "k",
		"ZETARIS_API_KEY":         "z",
		"ZETARIS_ORG_ID":          "1",
		"ZETARIS_REST_URL":        "http://10.0.0.7:8080",
	}
	cases := []struct {
		setting, host string
		load          func(func(string) string) error
	}{
		{"PROVEN_GENERATOR_BASE_URL", "nightshift.example.com", func(e func(string) string) error { _, err := generator.Load(e); return err }},
		{"PROVEN_GUARDRAIL_BASE_URL", "guard.example.com", func(e func(string) string) error { _, err := guardrail.Load(e); return err }},
		{"ZETARIS_REST_URL", "zetaris.example.com", func(e func(string) string) error { _, err := query.Load(e); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.setting, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range local {
				env[k] = v
			}
			env[tc.setting] = "https://" + tc.host + "/v1"
			err := tc.load(func(k string) string { return env[k] })
			field, text := refusedField(t, err)
			if field != tc.setting || !strings.Contains(text, tc.host) {
				t.Fatalf("refusal %q", text)
			}
			env["PROVEN_NETWORK"] = "public"
			if err := tc.load(func(k string) string { return env[k] }); err != nil {
				t.Fatalf("public mode refused it too: %v", err)
			}
		})
	}
}

// TestLocalRefusesTheHostedGuardrailByName proves ADR-0034 point 2: the guardrail's nvidia mode is refused
// in local mode even with a local address, and the refusal names PROVEN_GUARDRAIL_MODE.
// Mutated field: PROVEN_GUARDRAIL_MODE.
func TestLocalRefusesTheHostedGuardrailByName(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK": "local", "PROVEN_GUARDRAIL_MODE": "nvidia",
		"PROVEN_GUARDRAIL_BASE_URL": "http://127.0.0.1:8081", "PROVEN_GUARDRAIL_KEY": "k",
	}
	_, err := guardrail.Load(func(k string) string { return env[k] })
	field, text := refusedField(t, err)
	if field != "PROVEN_GUARDRAIL_MODE" || !strings.Contains(text, "nvidia") {
		t.Fatalf("refusal %q", text)
	}
	// Mode off still checks an address that is set, so a hosted one cannot wait there to be switched on.
	env["PROVEN_GUARDRAIL_MODE"] = "off"
	env["PROVEN_GUARDRAIL_BASE_URL"] = "https://integrate.api.nvidia.com/v1"
	_, err = guardrail.Load(func(k string) string { return env[k] })
	if field, text := refusedField(t, err); field != "PROVEN_GUARDRAIL_BASE_URL" || !strings.Contains(text, "integrate.api.nvidia.com") {
		t.Fatalf("refusal %q", text)
	}
}

// TestLocalRefusesTheHostedGeneratorByName proves ADR-0034 for the generator: PROVEN_GENERATOR_MODE=nvidia
// is refused in local mode even with a local address, and the refusal names the setting and the value.
// Mutated field: the refusal of PROVEN_GENERATOR_MODE in generator.Load.
func TestLocalRefusesTheHostedGeneratorByName(t *testing.T) {
	env := map[string]string{
		"PROVEN_NETWORK": "local", "PROVEN_GENERATOR_MODE": "nvidia",
		"PROVEN_GENERATOR_KEY": "k", "PROVEN_NVIDIA_BASE_URL": "https://127.0.0.1:8443",
	}
	_, err := generator.Load(func(k string) string { return env[k] })
	field, text := refusedField(t, err)
	if field != "PROVEN_GENERATOR_MODE" || !strings.Contains(text, "nvidia") {
		t.Fatalf("refusal %q", text)
	}
}

// TestLocalNamesAWrongNetworkModeFromEveryLoader proves a wrong PROVEN_NETWORK is refused by name
// wherever it is read, so no loader runs in a mode nobody chose.
func TestLocalNamesAWrongNetworkModeFromEveryLoader(t *testing.T) {
	env := func(k string) string {
		if k == "PROVEN_NETWORK" {
			return "airgap"
		}
		return ""
	}
	_, e1 := generator.Load(env)
	_, e2 := guardrail.Load(env)
	_, e3 := query.Load(env)
	_, e4 := testrun.Load(env)
	for _, err := range []error{e1, e2, e3, e4} {
		if field, text := refusedField(t, err); field != "PROVEN_NETWORK" || !strings.Contains(text, "airgap") {
			t.Fatalf("refusal %q", text)
		}
	}
}
