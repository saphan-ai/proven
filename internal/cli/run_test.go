package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

type captureReviewer struct {
	req         review.Request
	res         review.Result
	err         error
	hasDeadline bool
}

func (c *captureReviewer) Review(ctx context.Context, req review.Request) (review.Result, error) {
	c.req = req
	_, c.hasDeadline = ctx.Deadline()
	return c.res, c.err
}

type panicReviewer struct{}

func (panicReviewer) Review(context.Context, review.Request) (review.Result, error) {
	panic("boom")
}

func openCapture(c *captureReviewer) opener {
	return func(review.Options) (review.Reviewer, error) { return c, nil }
}

func TestSampleFixtureExitCodes(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	open := func(review.Options) (review.Reviewer, error) {
		t.Error("sample mode called the injected opener")
		return nil, review.ErrNotImplemented
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", "--error", target}, emptyEnv, &stdout, &stderr, open)
	if code != exitFindings {
		t.Fatalf("with --error: exit %d\nstderr: %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	assertDiscountText(t, stdout.String())
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"scan", "--sample", target}, emptyEnv, &stdout, &stderr, open)
	if code != exitOK {
		t.Fatalf("without --error: exit %d\nstderr: %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	assertDiscountText(t, stdout.String())
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"scan", "--error", target}, func(k string) string {
		if k == "SAMPLE_MODE" {
			return "true"
		}
		return ""
	}, &stdout, &stderr, open)
	if code != exitFindings {
		t.Fatalf("SAMPLE_MODE: exit %d\nstderr: %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	assertDiscountText(t, stdout.String())
}

func assertDiscountText(t *testing.T, body string) {
	t.Helper()
	for _, want := range []string{
		"finding F1 ERROR Discount is applied twice",
		"evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true",
		"finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list",
		"evidence query local-sample snapshot=sample-2026-10-22",
		"row: evil.example/leftpad | 1.2.3 | CVE-2026-0001 | true",
		"failed services: none",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "finding F3") {
		t.Fatalf("F3 was printed\n%s", body)
	}
}

func TestCICommitsComeFromTheFixtureEvent(t *testing.T) {
	event := filepath.Join(moduleRoot(t), "tests", "fixtures", "pull_request.json")
	cap := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		if k == "GITHUB_EVENT_PATH" {
			return event
		}
		return ""
	}, &stdout, &stderr, openCapture(cap))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if cap.req.BaselineCommit != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("baseline %q", cap.req.BaselineCommit)
	}
	if cap.req.HeadCommit != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("head %q", cap.req.HeadCommit)
	}
}

func TestCIFlagAndSHAOverrideTheEvent(t *testing.T) {
	event := filepath.Join(moduleRoot(t), "tests", "fixtures", "pull_request.json")
	cap := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci", "--baseline-commit=from-flag"}, func(k string) string {
		switch k {
		case "GITHUB_EVENT_PATH":
			return event
		case "GITHUB_SHA":
			return "cccccccccccccccccccccccccccccccccccccccc"
		case "GITHUB_BASE_REF":
			return "main"
		default:
			return ""
		}
	}, &stdout, &stderr, openCapture(cap))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if cap.req.BaselineCommit != "from-flag" {
		t.Fatalf("baseline %q", cap.req.BaselineCommit)
	}
	if cap.req.HeadCommit != "cccccccccccccccccccccccccccccccccccccccc" {
		t.Fatalf("head %q", cap.req.HeadCommit)
	}
}

func TestCIBaseRefWhenTheEventHasNoSHA(t *testing.T) {
	dir := t.TempDir()
	event := filepath.Join(dir, "event.json")
	body := []byte(`{"pull_request":{"base":{"ref":"main","sha":""},"head":{"ref":"feature","sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}`)
	if err := os.WriteFile(event, body, 0o644); err != nil {
		t.Fatal(err)
	}
	cap := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		switch k {
		case "GITHUB_EVENT_PATH":
			return event
		case "GITHUB_BASE_REF":
			return "release"
		default:
			return ""
		}
	}, &stdout, &stderr, openCapture(cap))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if cap.req.BaselineCommit != "release" {
		t.Fatalf("baseline %q", cap.req.BaselineCommit)
	}
	if cap.req.HeadCommit != "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("head %q", cap.req.HeadCommit)
	}
}

func TestCIWorkspaceIsTheTarget(t *testing.T) {
	event := filepath.Join(moduleRoot(t), "tests", "fixtures", "pull_request.json")
	cap := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		switch k {
		case "GITHUB_EVENT_PATH":
			return event
		case "GITHUB_WORKSPACE":
			return "/work"
		default:
			return ""
		}
	}, &stdout, &stderr, openCapture(cap))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if cap.req.Target != "/work" {
		t.Fatalf("target %q", cap.req.Target)
	}
}

func TestReviewerPanicExits2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		return panicReviewer{}, nil
	})
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestFailedServiceIsNamedAndDoesNotExit2(t *testing.T) {
	rev := &captureReviewer{res: review.Result{FailedServices: []review.ServiceFailure{{
		Service: "nightshift",
		Reason:  "unreachable",
	}}}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s stdout %s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "nightshift") {
		t.Fatalf("stdout %q", stdout.String())
	}

	rev.res.Findings = []review.Finding{{
		RuleID:   "sample.finding",
		Severity: review.SeverityError,
		Message:  "still named",
		Location: review.Location{Path: "a.go", StartLine: 1},
	}}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"scan", "--error"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitFindings {
		t.Fatalf("exit %d stderr %s stdout %s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "nightshift") {
		t.Fatalf("stdout %q", stdout.String())
	}
}

func TestQuietStillNamesTheFailedService(t *testing.T) {
	rev := &captureReviewer{res: review.Result{FailedServices: []review.ServiceFailure{{
		Service: "zetaris",
		Reason:  "down",
	}}}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--quiet"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout.String(), "zetaris") {
		t.Fatalf("stdout %q", stdout.String())
	}
}

func TestSeverityFilterCountsReportedFindings(t *testing.T) {
	rev := &captureReviewer{res: review.Result{Findings: []review.Finding{
		{RuleID: "info.rule", Severity: review.SeverityInfo, Message: "low", Location: review.Location{Path: "a.go", StartLine: 1}},
		{RuleID: "error.rule", Severity: review.SeverityError, Message: "high", Location: review.Location{Path: "b.go", StartLine: 2}},
	}}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--severity=ERROR", "--error"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitFindings {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "info.rule") || !strings.Contains(stdout.String(), "error.rule") {
		t.Fatalf("stdout %q", stdout.String())
	}
}

func TestTimeoutSetsADeadline(t *testing.T) {
	rev := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--timeout=5"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !rev.hasDeadline {
		t.Fatal("no deadline")
	}
	rev = &captureReviewer{}
	code = Run([]string{"scan"}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if rev.hasDeadline {
		t.Fatal("deadline without --timeout")
	}
}

func TestJSONFormatHasNoWriter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--json"}, emptyEnv, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "json") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestTextOutputFileReceivesTheFinding(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")
	rev := &captureReviewer{res: review.Result{Findings: []review.Finding{{
		RuleID:   "file.rule",
		Severity: review.SeverityWarning,
		Message:  "written",
		Location: review.Location{Path: "a.go", StartLine: 3},
	}}}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--text-output=" + out}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "file.rule") || !strings.Contains(stdout.String(), "file.rule") {
		t.Fatalf("file %q stdout %q", body, stdout.String())
	}
}

func TestOutputFlagReplacesStdout(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "chosen.txt")
	rev := &captureReviewer{res: review.Result{Findings: []review.Finding{{
		RuleID:   "chosen.rule",
		Severity: review.SeverityInfo,
		Message:  "in file",
		Location: review.Location{Path: "a.go", StartLine: 1},
	}}}}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "-o", out}, emptyEnv, &stdout, &stderr, openCapture(rev))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout %q", stdout.String())
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "chosen.rule") {
		t.Fatalf("file %q", body)
	}
}

func TestHelpAndVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--help"}, emptyEnv, &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("help exit %d", code)
	}
	if !strings.Contains(stdout.String(), "--severity") || !strings.Contains(stdout.String(), "configuration was refused") {
		t.Fatalf("help %q", stdout.String())
	}
	stdout.Reset()
	code = Run([]string{"--version"}, emptyEnv, &stdout, &stderr, nil)
	if code != exitOK || !strings.Contains(stdout.String(), "development") {
		t.Fatalf("version exit %d %q", code, stdout.String())
	}
}

func TestScanWithoutSampleOnBareTargetExits2(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "somewhere"}, emptyEnv, &stdout, &stderr, nil)
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), review.ErrNotImplemented.Error()) {
		t.Fatalf("stderr still says the reviewer is missing\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "fixture.json") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestRunUnknownSeverityExits7(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan"}, envOf(map[string]string{"PROVEN_SEVERITY": "CRITICAL"}), &stdout, &stderr, nil)
	if code != exitConfig {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "INFO, WARNING, ERROR") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestRunConfigAutoExits7(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--config=auto"}, emptyEnv, &stdout, &stderr, nil)
	if code != exitConfig {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "configuration file") {
		t.Fatalf("stderr %q", stderr.String())
	}
}
