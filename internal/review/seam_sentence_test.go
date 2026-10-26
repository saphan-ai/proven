package review

import (
	"testing"
)

// TestOpenPrintsTheSeamRefusalForTheBuiltinTransport proves the shell still
// reads the seam sentence when Setup refuses grpc. The direct Setup sentence
// is a different line and must not leak through this copy.
func TestOpenPrintsTheSeamRefusalForTheBuiltinTransport(t *testing.T) {
	_, err := Open(Options{
		Sample: true,
		LogDir: t.TempDir(),
		Env: func(k string) string {
			switch k {
			case "OTEL_EXPORTER_OTLP_ENDPOINT":
				return "http://127.0.0.1:4318"
			case "OTEL_EXPORTER_OTLP_PROTOCOL":
				return "grpc"
			default:
				return ""
			}
		},
	})
	const want = `configuration refused: OTEL_EXPORTER_OTLP_PROTOCOL="grpc": only the HTTP transport is built in; allowed: http/protobuf`
	if err == nil || err.Error() != want {
		got := ""
		if err != nil {
			got = err.Error()
		}
		t.Fatalf("got %q want %q", got, want)
	}
}
