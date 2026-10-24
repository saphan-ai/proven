package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// hostedLineKey is a stand-in for an NVIDIA key. No real key is used.
const hostedLineKey = "nvapi-line-key-5b2e90c1"

// hostedTransport serves one recorded hosted answer whose chunks name model.
type hostedTransport struct{ model string }

// RoundTrip answers every request with a hosted stream, so no port is opened.
func (h hostedTransport) RoundTrip(*http.Request) (*http.Response, error) {
	body := `data: {"model":"` + h.model + `","choices":[{"delta":{"content":"ok"}}]}` + "\n\n" + "data: [DONE]\n\n"
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
}

// runHosted runs livecheck in nvidia mode over a recorded answer and returns
// its output. Only the generator is reachable: the other services have no
// settings, so none of them dials.
func runHosted(t *testing.T, model string) string {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
	was := newLive
	newLive = func(cfg generator.Settings) *generator.Live {
		return generator.NewLive(cfg, &http.Client{Transport: hostedTransport{model: model}})
	}
	t.Cleanup(func() { newLive = was })
	env := func(k string) string {
		switch k {
		case generator.EnvMode:
			return "nvidia"
		case generator.EnvKey:
			return hostedLineKey
		}
		return ""
	}
	var stdout, stderr bytes.Buffer
	run(nil, &stdout, &stderr, env)
	return stdout.String() + stderr.String()
}

// Proves the generator line in nvidia mode names the mode, the model the answer
// named, "no receipt", and the host the answer came from. Mutated field: the
// mode in the line, which a fixed "nightshift" would print.
func TestNVIDIAModeNamesModelReceiptAndHost(t *testing.T) {
	out := runHosted(t, "nvidia/line-test")
	want := "generator: ready: mode nvidia, model nvidia/line-test, receipt no receipt, host integrate.api.nvidia.com"
	if !strings.Contains(out, want) {
		t.Fatalf("want %q in\n%s", want, out)
	}
}

// Proves the NVIDIA key is redacted from the generator line when a hosted
// answer echoes it in its model field. Mutated field: the key's entry in
// redact.Names, and the model field of the answer, which carries the key.
func TestNVIDIAKeyEchoedInTheModelFieldIsNotPrinted(t *testing.T) {
	out := runHosted(t, "nvidia/"+hostedLineKey)
	if strings.Contains(out, hostedLineKey) {
		t.Fatalf("the key reached terminal output\n%s", out)
	}
	if !strings.Contains(out, "model nvidia/[redacted: "+generator.EnvKey+"]") {
		t.Fatalf("the line does not carry the marker\n%s", out)
	}
}
