package generator

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStreamMetadataAndUnknownUsage(t *testing.T) {
	sse := "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"żółw\",\"reasoning\":\"duplicate\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"reasoning\":\"later\",\"content\":\"answer\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: {\"usage\":{\"completion_tokens\":0}}\n\ndata: [DONE]\n\n"
	m, err := parseStreamOptions(strings.NewReader(sse), 1)
	if err != nil || m.text != "answer" || m.reasoningBytes != len("żółwlater") || m.reasoning != "" || !m.reasoningClipped || !utf8.ValidString(m.reasoning) {
		t.Fatalf("%#v %v", m, err)
	}
	if m.usage == nil || m.usage.CompletionTokens == nil || *m.usage.CompletionTokens != 0 || m.usage.TotalTokens != nil || m.usage.ReasoningTokens != nil {
		t.Fatalf("unknown values invented: %#v", m.usage)
	}
	for _, raw := range []string{`null`, `{}`, `{"total_tokens":-1}`, `{"total_tokens":"unknown"}`} {
		if u := readUsage([]byte(raw)); u != nil {
			t.Fatalf("%s: %#v", raw, u)
		}
	}
}
