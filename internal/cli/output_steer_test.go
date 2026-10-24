package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// rawControlIn reports the first raw control byte (newline apart) or raw C1 rune in s.
func rawControlIn(s string) (int, bool) {
	for i, r := range s {
		if r == '\n' {
			continue
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return i, true
		}
	}
	return 0, false
}

// TestJSONAndSARIFCarryNoRawDELOrC1 proves a title or message that holds DEL
// or a C1 rune is not written as that raw rune in JSON or SARIF. The mutated
// field is the bytes of the encoded document.
func TestJSONAndSARIFCarryNoRawDELOrC1(t *testing.T) {
	result := review.Result{Findings: []review.Finding{{
		ID: "F1", Severity: review.SeverityError, Title: "title\x7f\u009b[2J", Message: "msg\u009b[31m",
		Location: review.Location{Path: "a.go", StartLine: 1, EndLine: 1},
	}}}
	for _, format := range []string{"json", "sarif"} {
		body, err := render(format, result)
		if err != nil {
			t.Fatal(err)
		}
		if i, ok := rawControlIn(string(body)); ok {
			t.Errorf("%s: raw control byte at offset %d:\n%s", format, i, body)
		}
	}
}

// TestSARIFArtifactURIStaysARelativePathInTheRepository proves artifactURI
// does not yield a leading slash or a ".." segment. The mutated field is the
// path passed to artifactURI.
func TestSARIFArtifactURIStaysARelativePathInTheRepository(t *testing.T) {
	for _, path := range []string{"//example.invalid/a.go", "/etc/a.go", "../outside/a.go", "a/../../b.go"} {
		uri := artifactURI(path)
		if strings.HasPrefix(uri, "/") {
			t.Errorf("artifactURI(%q) = %q starts with a slash (an authority or a root)", path, uri)
		}
		for _, seg := range strings.Split(uri, "/") {
			if seg == ".." {
				t.Errorf("artifactURI(%q) = %q climbs out with a \"..\" segment", path, uri)
			}
		}
	}
}

// TestOutputFileIsNotWrittenThroughALinkInTheCheckout proves an output path
// whose final component is a symbolic link does not replace the file the link
// names. The mutated field is that final component: a link, not a regular file.
func TestOutputFileIsNotWrittenThroughALinkInTheCheckout(t *testing.T) {
	dir := t.TempDir()
	checkout := filepath.Join(dir, "checkout")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(outside, "kept.txt")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(checkout, "proven.sarif")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	cfg := config{format: "text", sarifOutput: link}
	var stdout strings.Builder
	_ = writeOutputs(cfg, review.Result{Class: "c", ClassRule: "r"}, &stdout)
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
	}
}
