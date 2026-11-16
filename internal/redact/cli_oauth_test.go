package redact

import "testing"

func TestCLIOAuthIsRedactedEvenInCommandMode(t *testing.T) {
	r := FromEnv(func(k string) string {
		if k == "CLAUDE_CODE_OAUTH_TOKEN" {
			return "subscription-oauth-value"
		}
		return ""
	})
	if got := r.Text("echo subscription-oauth-value"); got != "echo [redacted: CLAUDE_CODE_OAUTH_TOKEN]" {
		t.Fatal(got)
	}
}
