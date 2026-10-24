package cli

import (
	"bytes"
	"encoding/json"
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
		// The escape is still JSON, and it decodes to the same title and message.
		if !bytes.Contains(body, []byte(`\u007f`)) || !bytes.Contains(body, []byte(`\u009b`)) {
			t.Errorf("%s: DEL or C1 was not written as a \\u00XX escape\n%s", format, body)
		}
		var doc map[string]any
		if err := json.Unmarshal(body, &doc); err != nil {
			t.Errorf("%s: document is not JSON: %v", format, err)
			continue
		}
		title, message := decodedTitleAndMessage(t, format, doc)
		if title != "title\x7f\u009b[2J" || message != "msg\u009b[31m" {
			t.Errorf("%s: decoded title %q message %q", format, title, message)
		}
	}
}

// decodedTitleAndMessage reads the finding title and message back from a
// JSON document or a SARIF log.
func decodedTitleAndMessage(t *testing.T, format string, doc map[string]any) (string, string) {
	t.Helper()
	if format == "json" {
		findings, _ := doc["findings"].([]any)
		if len(findings) != 1 {
			t.Fatalf("json findings %#v", doc["findings"])
		}
		one, _ := findings[0].(map[string]any)
		title, _ := one["title"].(string)
		message, _ := one["message"].(string)
		return title, message
	}
	runs, _ := doc["runs"].([]any)
	if len(runs) != 1 {
		t.Fatalf("sarif runs %#v", doc["runs"])
	}
	run, _ := runs[0].(map[string]any)
	results, _ := run["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("sarif results %#v", run["results"])
	}
	one, _ := results[0].(map[string]any)
	msg, _ := one["message"].(map[string]any)
	message, _ := msg["text"].(string)
	props, _ := one["properties"].(map[string]any)
	title, _ := props["title"].(string)
	return title, message
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
	// A path that cleans back inside the repository stays a relative URI.
	if got := artifactURI("dir/../a.go"); got != "a.go" {
		t.Errorf("artifactURI(%q) = %q, want %q", "dir/../a.go", got, "a.go")
	}
}

// TestSARIFKeepsAFindingWhenThePathLeavesTheRepository proves a path that
// cannot be a relative URI is omitted and the finding stays in the log.
// The mutated field is the location path. This log sets no uriBaseId.
func TestSARIFKeepsAFindingWhenThePathLeavesTheRepository(t *testing.T) {
	body, err := render("sarif", review.Result{Findings: []review.Finding{{
		ID: "F1", Severity: review.SeverityError, Message: "kept",
		Location: review.Location{Path: "/etc/a.go", StartLine: 4, EndLine: 4},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	doc := decodeSARIF(t, body)
	results := doc.Runs[0].Results
	if len(results) != 1 || results[0].RuleID != "F1" || results[0].Message.Text != "kept" {
		t.Fatalf("finding was not kept: %+v", results)
	}
	if n := len(results[0].Locations); n != 0 {
		t.Fatalf("%d locations, want none for a path outside the repository", n)
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

// TestFinalSymlinkOutputIsRefusedAtLoad proves each output flag refuses a
// final symbolic link at load with exit 7, naming the flag and the path,
// before the review and before any write. The mutated field is that final
// component. With the load refusal removed, the review runs.
func TestFinalSymlinkOutputIsRefusedAtLoad(t *testing.T) {
	for _, f := range outputFlags {
		t.Run(f.name, func(t *testing.T) {
			dir := t.TempDir()
			victim := filepath.Join(dir, "kept.txt")
			if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(dir, "out")
			if err := os.Symlink(victim, link); err != nil {
				t.Fatal(err)
			}
			rev := &captureReviewer{}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", f.arg + link}, emptyEnv, &stdout, &stderr, openCapture(rev))
			if code != exitConfig {
				t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
			}
			for _, want := range []string{f.name, link} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("refusal does not name %q: %s", want, stderr.String())
				}
			}
			if rev.req.Target != "" {
				t.Error("the review ran despite the refusal")
			}
			b, err := os.ReadFile(victim)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != "kept\n" {
				t.Fatalf("the report was written through the link:\n%s", b)
			}
		})
	}
}

// TestRegularOutputFileIsStillOverwritten proves the symlink refusal does not
// cover a regular file. The mutated field is the final component: a regular
// file, not a link. Refusing every existing path turns this red.
func TestRegularOutputFileIsStillOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proven.sarif")
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout strings.Builder
	if err := writeOutputs(config{format: "text", sarifOutput: path}, review.Result{Class: "c", ClassRule: "r"}, &stdout); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "old\n" || !bytes.Contains(b, []byte(`"class": "c"`)) {
		t.Fatalf("regular file was not overwritten:\n%s", b)
	}
}
