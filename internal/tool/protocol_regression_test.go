package tool

import (
	"encoding/json"
	"testing"
)

func TestGovulncheckAcceptsIndentedValueStream(t *testing.T) {
	compact := `{"config":{"scanner_name":"govulncheck"}}` + "\n" + `{"finding":{"trace":[{"module":"example.com/lib","package":"example.com/lib","function":"Bad"},{"module":"example.com/app","function":"main"}]}}`
	var pretty string
	for _, raw := range []string{`{"config":{"scanner_name":"govulncheck"}}`, `{"finding":{"trace":[{"module":"example.com/lib","package":"example.com/lib","function":"Bad"},{"module":"example.com/app","function":"main"}]}}`} {
		var value any
		if err := json.Unmarshal([]byte(raw), &value); err != nil {
			t.Fatal(err)
		}
		b, _ := json.MarshalIndent(value, "", "  ")
		pretty += string(b) + "\n"
	}
	for _, text := range []string{compact, pretty} {
		if !ScanStarted(text) || len(Symbols(text)) != 1 {
			t.Fatalf("valid stream lost: %s", text)
		}
	}
	if ScanStarted(pretty+`{"finding":`) || len(Symbols(pretty+`{"finding":`)) != 0 {
		t.Fatal("truncated stream counted as measurement")
	}
}
