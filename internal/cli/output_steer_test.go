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

// TestIntermediateDirectoryLinkInsideTheWorkingDirectoryIsRefusedAtLoad proves
// an output path whose directory component lies inside the working directory
// and is a symbolic link is refused at load with exit 7. The refusal names
// the flag, the path and that linked component, and nothing is written
// outside. The mutated field is that directory component: a link, not a real
// directory. With the walk removed the write follows the link.
func TestIntermediateDirectoryLinkInsideTheWorkingDirectoryIsRefusedAtLoad(t *testing.T) {
	for _, f := range outputFlags {
		t.Run(f.name, func(t *testing.T) {
			_, link, victim := intermediateLinkCheckout(t, "reports")
			rel := "reports/proven.sarif"
			assertIntermediateLinkRefused(t, f, rel, link, victim)
		})
	}
	sarif := outputFlag{"--sarif-output", "--sarif-output="}
	t.Run("nested directory", func(t *testing.T) {
		_, link, victim := intermediateLinkCheckout(t, filepath.Join("sub", "reports"))
		rel := "sub/reports/proven.sarif"
		assertIntermediateLinkRefused(t, sarif, rel, link, victim)
	})
	t.Run("absolute path under the working directory", func(t *testing.T) {
		checkout, link, victim := intermediateLinkCheckout(t, "reports")
		abs := filepath.Join(checkout, "reports", "proven.sarif")
		assertIntermediateLinkRefused(t, sarif, abs, link, victim)
	})
}

// intermediateLinkCheckout makes a working directory whose named directory
// component is a symbolic link to a directory outside it. It returns that
// working directory as the process reports it, the link joined onto it, and
// victim, the file the link would steer the report onto, seeded so a write
// is visible.
func intermediateLinkCheckout(t *testing.T, linkRel string) (checkout, link, victim string) {
	t.Helper()
	dir := t.TempDir()
	checkout = filepath.Join(dir, "checkout")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	link = filepath.Join(checkout, linkRel)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	victim = filepath.Join(outside, "proven.sarif")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	// The refusal names the component the walk Lstats, which is joined onto
	// the working directory the process reports, not the path TempDir returned.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	link = filepath.Join(wd, linkRel)
	return wd, link, victim
}

// assertIntermediateLinkRefused runs one output flag and requires exit 7,
// the flag, the path and the linked component in the refusal, no review,
// and the outside file left as it was.
func assertIntermediateLinkRefused(t *testing.T, f outputFlag, path, link, victim string) {
	t.Helper()
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", f.arg + path}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitConfig {
		t.Errorf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	for _, want := range []string{f.name, path, link, "inside the working directory"} {
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
		t.Errorf("the report was written through the link onto %s:\n%s", victim, b)
	}
}

// TestOutputInARealDirectoryInsideTheWorkingDirectoryIsWritten proves a
// directory component inside the working directory is written through when
// it is a real directory. The mutated field is that component: a directory,
// not a link. Refusing every directory component turns this red.
func TestOutputInARealDirectoryInsideTheWorkingDirectoryIsWritten(t *testing.T) {
	checkout := t.TempDir()
	if err := os.Mkdir(filepath.Join(checkout, "reports"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	rev := &captureReviewer{res: review.Result{Class: "c", ClassRule: "r"}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=reports/proven.sarif"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d, want 0; stderr %s", code, stderr.String())
	}
	b, err := os.ReadFile(filepath.Join("reports", "proven.sarif"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"class": "c"`)) {
		t.Fatalf("report was not written:\n%s", b)
	}
}

// TestDirectoryLinkOutsideTheWorkingDirectoryIsStillWritten proves an
// intermediate symbolic link that does not lie inside the working directory
// is not refused. The mutated field is that directory: it sits on an absolute
// path outside the working directory. Refusing it turns this red.
func TestDirectoryLinkOutsideTheWorkingDirectoryIsStillWritten(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	t.Chdir(t.TempDir())
	out := filepath.Join(link, "proven.sarif")
	rev := &captureReviewer{res: review.Result{Class: "c", ClassRule: "r"}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=" + out}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d, want 0; stderr %s", code, stderr.String())
	}
	b, err := os.ReadFile(filepath.Join(real, "proven.sarif"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"class": "c"`)) {
		t.Fatalf("report was not written through the outside link:\n%s", b)
	}
}

// TestOpenRefusesAFinalLinkSwappedInAfterTheCheck proves the open itself
// refuses a symbolic link as the final component, so a link that replaces
// the path after the Lstat is not followed. The mutated field is the open
// flag O_NOFOLLOW. Without it the file the link names is replaced. On a
// system whose open has no such flag the test does not run: the Lstat before
// the write is the only guard, and a swap between the two can still steer it.
func TestOpenRefusesAFinalLinkSwappedInAfterTheCheck(t *testing.T) {
	if !outputOpenRefusesFinalLink {
		t.Skip("this OS open follows a final symbolic link; the Lstat check is the only guard")
	}
	dir := t.TempDir()
	victim := filepath.Join(dir, "kept.txt")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "out")
	if err := os.Symlink(victim, link); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	f, err := openOutputFile(link)
	if err == nil {
		f.Close()
		t.Error("open followed the final symbolic link")
	}
	b, rerr := os.ReadFile(victim)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
	}
}

// TestFinalLinkAppearingAfterLoadExits2 proves a final symbolic link that
// appears after the load check is refused by the write with exit 2, and the
// file it names is left as it was. The mutated field is that final component
// during the review: a link, where load saw a regular file.
func TestFinalLinkAppearingAfterLoadExits2(t *testing.T) {
	dir := t.TempDir()
	victim := filepath.Join(dir, "kept.txt")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	if err := os.WriteFile(out, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			if err := os.Remove(out); err != nil {
				t.Error(err)
			}
			if err := os.Symlink(victim, out); err != nil {
				t.Error(err)
			}
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=" + out}, emptyEnv, &stdout, &stderr, open)
	if code != exitFatal {
		t.Fatalf("exit %d, want %d; stderr %s", code, exitFatal, stderr.String())
	}
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
	}
}

// TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile proves a
// report does not replace another file of the checkout when the output's
// final component becomes a symbolic link to that file after the
// final-component check and before the rooted create. The mutated field is
// that component. The link target is a relative name inside the checkout.
// Opening the final name with OpenFile follows the link and the other file
// changes. The review runs before the check, so a swap there is refused and
// never reaches this window.
func TestFinalComponentSwappedAfterTheCheckDoesNotChangeTheOtherFile(t *testing.T) {
	dir := t.TempDir()
	checkout := filepath.Join(dir, "checkout")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(checkout, "kept.txt")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(checkout, "proven.sarif")
	if err := os.WriteFile(out, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	beforeRootedOutputCreate = func(path string) {
		if err := os.Remove(path); err != nil {
			t.Error(err)
			return
		}
		if err := os.Symlink("kept.txt", path); err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(func() { beforeRootedOutputCreate = nil })
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=proven.sarif"}, emptyEnv, &stdout, &stderr, open)
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
	}
	if code != exitOK {
		t.Fatalf("exit %d, want 0; stderr %s", code, stderr.String())
	}
	info, err := os.Lstat(out)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("final name is still a symbolic link")
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(written, []byte(`"class": "c"`)) {
		t.Fatalf("report was not written at the final name:\n%s", written)
	}
	entries, err := os.ReadDir(checkout)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "kept.txt" && e.Name() != "proven.sarif" {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}
}

// TestDirectoryLinkSwappedAfterLoadIsNotWrittenThrough proves a directory of
// the output path that is a real directory at load, and a symbolic link to a
// directory outside the working directory by the time of the write, does not
// receive the report. The review is the seam between the load check and the
// write: it replaces the directory. The mutated field is that directory
// component. Writing the file with os.WriteFile follows the link, the run
// exits 0, and the outside file changes.
func TestDirectoryLinkSwappedAfterLoadIsNotWrittenThrough(t *testing.T) {
	rel := "reports/proven.sarif"
	t.Run("relative path", func(t *testing.T) {
		assertSwappedDirectoryDoesNotReceiveTheReport(t, func(checkout string) string { return rel }, "")
	})
	t.Run("absolute path under the working directory", func(t *testing.T) {
		assertSwappedDirectoryDoesNotReceiveTheReport(t, func(checkout string) string {
			return filepath.Join(checkout, rel)
		}, "")
	})
	t.Run("relative link target", func(t *testing.T) {
		assertSwappedDirectoryDoesNotReceiveTheReport(t, func(checkout string) string { return rel }, "../outside")
	})
}

// assertSwappedDirectoryDoesNotReceiveTheReport builds a checkout whose
// reports directory is real at load, swaps it for a link to an outside
// directory during the review, and requires a non-zero exit that names the
// output path. The outside file must still hold what was seeded.
func assertSwappedDirectoryDoesNotReceiveTheReport(t *testing.T, outputPath func(checkout string) string, linkTarget string) {
	t.Helper()
	dir := t.TempDir()
	checkout := filepath.Join(dir, "checkout")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(filepath.Join(checkout, "reports"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(outside, "proven.sarif")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := outputPath(wd)
	reports := filepath.Join(wd, "reports")
	swapped := false
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			if err := os.Remove(reports); err != nil {
				t.Error(err)
			}
			target := outside
			if linkTarget != "" {
				target = linkTarget
			}
			if err := os.Symlink(target, reports); err != nil {
				t.Error(err)
			}
			swapped = true
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=" + path}, emptyEnv, &stdout, &stderr, open)
	if !swapped {
		t.Fatal("the load check refused the path, so the swap never ran")
	}
	if code == exitOK {
		t.Errorf("exit 0, want non-zero; stderr %s", stderr.String())
	} else if !strings.Contains(stderr.String(), path) {
		t.Errorf("refusal does not name %q: %s", path, stderr.String())
	}
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
	}
}

// TestAliasedWorkingDirectorySpellingRefusesTheInsideLink proves an absolute
// output path that reaches the working directory through a different spelling
// is still checked for a directory link inside that directory. The mutated
// field is the comparison: the working directory and the walked prefix are
// resolved before they are compared. Comparing the spellings leaves the
// outside file overwritten and the run exits 0.
func TestAliasedWorkingDirectorySpellingRefusesTheInsideLink(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	checkout := filepath.Join(real, "checkout")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	link := filepath.Join(checkout, "reports")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(outside, "proven.sarif")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	path := filepath.Join(alias, "checkout", "reports", "proven.sarif")
	named := filepath.Join(alias, "checkout", "reports")
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=" + path}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitConfig {
		t.Errorf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
	}
	for _, want := range []string{"--sarif-output", path, named, "inside the working directory"} {
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
		t.Errorf("the report was written through the link onto %s:\n%s", victim, b)
	}
}

// TestDirectoryLinkSwappedAfterLoadThroughAnAliasedSpellingIsNotWrittenThrough
// proves the write, not only the load check, treats an aliased spelling of the
// working directory as inside it. The directory is real at load and becomes a
// link during the review. The mutated field is that classification. Treating
// the aliased spelling as outside follows the link.
func TestDirectoryLinkSwappedAfterLoadThroughAnAliasedSpellingIsNotWrittenThrough(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	checkout := filepath.Join(real, "checkout")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(filepath.Join(checkout, "reports"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
	victim := filepath.Join(outside, "proven.sarif")
	if err := os.WriteFile(victim, []byte("kept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(alias, "checkout", "reports", "proven.sarif")
	reports := filepath.Join(wd, "reports")
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			if err := os.Remove(reports); err != nil {
				t.Error(err)
			}
			if err := os.Symlink(outside, reports); err != nil {
				t.Error(err)
			}
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=" + path}, emptyEnv, &stdout, &stderr, open)
	if code == exitOK {
		t.Errorf("exit 0, want non-zero; stderr %s", stderr.String())
	} else if !strings.Contains(stderr.String(), path) {
		t.Errorf("refusal does not name %q: %s", path, stderr.String())
	}
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "kept\n" {
		t.Fatalf("the report was written through the link onto %s:\n%s", victim, b)
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

// TestDirectoryRenamedAndLinkedInsideTheCheckoutDoesNotRedirectTheReport
// proves a directory of the output path that the review renames away and
// replaces with a relative link to another directory of the checkout does
// not receive the report under that other directory. The mutated field is
// the directory component, swapped for an in-root link the root would
// follow. Resolving the path again at the write replaces
// redirected/proven.sarif and the run exits 0.
func TestDirectoryRenamedAndLinkedInsideTheCheckoutDoesNotRedirectTheReport(t *testing.T) {
	dir := t.TempDir()
	checkout := filepath.Join(dir, "checkout")
	for _, d := range []string{"reports", "redirected"} {
		if err := os.MkdirAll(filepath.Join(checkout, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	victim := filepath.Join(checkout, "redirected", "proven.sarif")
	if err := os.WriteFile(victim, []byte("committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	swapped := false
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			if err := os.Rename("reports", "reports.before"); err != nil {
				t.Error(err)
			}
			if err := os.Symlink("redirected", "reports"); err != nil {
				t.Skipf("symlinks are not available here: %v", err)
			}
			swapped = true
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=reports/proven.sarif"}, emptyEnv, &stdout, &stderr, open)
	if !swapped {
		t.Fatal("the load check refused the path, so the swap never ran")
	}
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "committed\n" {
		t.Fatalf("exit %d; the report replaced %s:\n%s", code, victim, b)
	}
	if code != exitOK {
		t.Fatalf("exit %d, want 0; stderr %s", code, stderr.String())
	}
	written, err := os.ReadFile(filepath.Join(checkout, "reports.before", "proven.sarif"))
	if err != nil {
		t.Fatalf("the report is not in the directory the load check saw: %v", err)
	}
	if !bytes.Contains(written, []byte(`"class": "c"`)) {
		t.Fatalf("report was not written in the directory opened at load:\n%s", written)
	}
}

// TestAbsentDirectoryLinkedDuringTheReviewDoesNotReceiveTheReport proves an
// output whose directory does not exist at load is not written when the
// review creates that directory as a relative link to another directory of
// the checkout. The mutated field is the absent directory. Resolving the
// path again at the write follows the link inside the root, replaces
// redirected/proven.sarif and exits 0.
func TestAbsentDirectoryLinkedDuringTheReviewDoesNotReceiveTheReport(t *testing.T) {
	dir := t.TempDir()
	checkout := filepath.Join(dir, "checkout")
	if err := os.MkdirAll(filepath.Join(checkout, "redirected"), 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(checkout, "redirected", "proven.sarif")
	if err := os.WriteFile(victim, []byte("committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(checkout)
	open := func(review.Options) (review.Reviewer, error) {
		return funcReviewer(func() review.Result {
			if err := os.Symlink("redirected", "reports"); err != nil {
				t.Skipf("symlinks are not available here: %v", err)
			}
			return review.Result{Class: "c", ClassRule: "r"}
		}), nil
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sarif-output=reports/proven.sarif"}, emptyEnv, &stdout, &stderr, open)
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "committed\n" {
		t.Fatalf("exit %d; the report replaced %s:\n%s", code, victim, b)
	}
	if code == exitOK {
		t.Fatalf("exit 0, want non-zero; stderr %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "reports/proven.sarif") {
		t.Fatalf("the error does not name the path: %s", stderr.String())
	}
}
