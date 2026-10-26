package telemetry

import (
	"context"
	"testing"
)

// TestSetupPrintsTheBuiltinTransportRefusal proves a direct caller of Setup
// still reads the sentence the package printed before the shared error type.
// The allowed list stays on the error and off this line. Swapping the
// sentence back to the seam line makes this test fail.
func TestSetupPrintsTheBuiltinTransportRefusal(t *testing.T) {
	_, err := Setup(context.Background(), func(k string) string {
		switch k {
		case EnvEndpoint:
			return "http://127.0.0.1:4318"
		case EnvProtocol:
			return "grpc"
		default:
			return ""
		}
	})
	const want = `telemetry: OTEL_EXPORTER_OTLP_PROTOCOL="grpc": only the HTTP transport is built in`
	if err == nil || err.Error() != want {
		got := ""
		if err != nil {
			got = err.Error()
		}
		t.Fatalf("got %q want %q", got, want)
	}
}
