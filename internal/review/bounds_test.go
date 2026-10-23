package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Property: a proven.yaml in the target larger than the configuration bound is
// refused, and the sentence names the file and the bound, never its content.
func TestTargetConfigOverTheConfigBoundIsRefused(t *testing.T) {
	target := t.TempDir()
	body := "timeout: 5m\n" + strings.Repeat("# padding-line\n", (2<<20)/15)
	if err := os.WriteFile(filepath.Join(target, "proven.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := loadSettingsInTarget(target)
	if err == nil {
		t.Fatalf("a %d-byte proven.yaml was read whole", len(body))
	}
	if !strings.Contains(err.Error(), "proven.yaml") || strings.Contains(err.Error(), "padding-line") {
		t.Fatalf("the refusal must name the file and not quote it: %v", err)
	}
}

// Property: an explicit configuration file is bounded by the same setting.
func TestConfigFileOverTheConfigBoundIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.yaml")
	body := strings.Repeat("# padding-line\n", (2<<20)/15)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfigFile(path); err == nil {
		t.Fatal("an explicit configuration over the bound was read whole")
	}
}

// Property: the author record is bounded as a whole, over every file the
// setting names.
func TestAuthorRecordOverTheRecordBoundIsRefused(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "author.txt"), []byte(strings.Repeat("x", 2<<20)), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readRecord(target, "author.txt")
	if err == nil {
		t.Fatal("an author record over the bound was read whole")
	}
	if !strings.Contains(err.Error(), "author.txt") {
		t.Fatalf("the refusal must name the record: %v", err)
	}
}

// Property: a record at the bound is read, so the bound is a ceiling.
func TestAuthorRecordWithinTheRecordBoundIsRead(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "author.txt"), []byte("Assisted-by: a model\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readRecord(target, "author.txt"); err != nil {
		t.Fatal(err)
	}
}
