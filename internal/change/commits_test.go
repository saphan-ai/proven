package change

import (
	"testing"
)

func TestCommitMessagesKeepEmptyAndMultilineSourcesWithTheirSHAs(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	git(t, dir, "commit", "--allow-empty", "-m", "Create repository baseline")
	base := rev(t, dir)
	git(t, dir, "commit", "--allow-empty", "--allow-empty-message", "-m", "")
	empty := rev(t, dir)
	message := "Reject invalid payment amounts\n\nKeep repeated requests idempotent."
	git(t, dir, "commit", "--allow-empty", "-m", message)
	head := rev(t, dir)
	got, err := NewPolicy(false, DefaultLimits()).CommitMessages(dir, base, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].SHA != empty || got[0].Message != "" || got[1].SHA != head || got[1].Message != message {
		t.Fatalf("lost source identity or content: %#v", got)
	}
}
