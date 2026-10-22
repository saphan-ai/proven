package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		if k == "GITHUB_EVENT_PATH" {
			return event
		}
		return ""
	}, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") || !strings.Contains(stderr.String(), "merge base") {
		t.Fatalf("stderr %q", stderr.String())
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
	withRefusingGlobalHook(t)
	dir, fork, baseTip, head := advancedBaseRepo(t, "release")
	event := filepath.Join(dir, "event.json")
	body := fmt.Sprintf(`{"pull_request":{"base":{"ref":"main","sha":""},"head":{"ref":"feature","sha":"%s"}}}`, head)
	if err := os.WriteFile(event, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
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
	if cap.req.BaselineCommit != fork {
		t.Fatalf("baseline %q, fork %q, base tip %q", cap.req.BaselineCommit, fork, baseTip)
	}
	if cap.req.BaselineCommit == baseTip {
		t.Fatalf("baseline is the base tip %s", baseTip)
	}
	if cap.req.HeadCommit != head {
		t.Fatalf("head %q", cap.req.HeadCommit)
	}
}

func TestCIWorkspaceIsTheTarget(t *testing.T) {
	withRefusingGlobalHook(t)
	dir, sha := singleCommitRepo(t)
	event := writePullRequestEvent(t, dir, sha, sha)
	cap := &captureReviewer{}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		switch k {
		case "GITHUB_EVENT_PATH":
			return event
		case "GITHUB_WORKSPACE":
			return dir
		default:
			return ""
		}
	}, &stdout, &stderr, openCapture(cap))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if cap.req.Target != dir {
		t.Fatalf("target %q", cap.req.Target)
	}
}

func TestCIDefaultConfigIsTheWorkspaceFile(t *testing.T) {
	outside := t.TempDir()
	workspace := t.TempDir()
	want := filepath.Join(workspace, "proven.yaml")
	if err := os.WriteFile(want, []byte("timeout: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	flagPath := filepath.Join(outside, "flag.yaml")
	if err := os.WriteFile(flagPath, []byte("timeout: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(outside, "env.yaml")
	if err := os.WriteFile(envPath, []byte("timeout: 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(outside)

	t.Run("default", func(t *testing.T) {
		got := runCIOpen(t, []string{"ci"}, workspace, "")
		if got != want {
			t.Fatalf("ConfigPath %q, workspace file %q", got, want)
		}
	})
	t.Run("flag", func(t *testing.T) {
		got := runCIOpen(t, []string{"ci", "--config=" + flagPath}, workspace, "")
		if got != flagPath {
			t.Fatalf("ConfigPath %q, flag %q", got, flagPath)
		}
	})
	t.Run("variable", func(t *testing.T) {
		got := runCIOpen(t, []string{"ci"}, workspace, envPath)
		if got != envPath {
			t.Fatalf("ConfigPath %q, variable %q", got, envPath)
		}
	})
}

func runCIOpen(t *testing.T, args []string, workspace, provenConfig string) string {
	t.Helper()
	var got review.Options
	var stdout, stderr bytes.Buffer
	code := Run(args, func(k string) string {
		switch k {
		case "GITHUB_WORKSPACE":
			return workspace
		case "PROVEN_CONFIG":
			return provenConfig
		default:
			return ""
		}
	}, &stdout, &stderr, func(opts review.Options) (review.Reviewer, error) {
		got = opts
		return &captureReviewer{}, nil
	})
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	return got.ConfigPath
}

func TestOpenerPanicExits2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
		panic("open-boom")
	})
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "open-boom") {
		t.Fatalf("stderr %q", stderr.String())
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

func TestReviewConfigErrorExits7(t *testing.T) {
	t.Run("config", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
			return &captureReviewer{err: &review.ConfigError{
				Field:   "timeout",
				Value:   "nope",
				Reason:  "not a duration",
				Allowed: []string{"a Go duration"},
			}}, nil
		})
		if code != exitConfig {
			t.Fatalf("exit %d stderr %s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "timeout") || !strings.Contains(stderr.String(), "nope") {
			t.Fatalf("stderr %q", stderr.String())
		}
	})
	t.Run("other", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"scan"}, emptyEnv, &stdout, &stderr, func(review.Options) (review.Reviewer, error) {
			return &captureReviewer{err: fmt.Errorf("review broke")}, nil
		})
		if code != exitFatal {
			t.Fatalf("exit %d stderr %s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "review broke") {
			t.Fatalf("stderr %q", stderr.String())
		}
	})
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

// 9223372036 is the largest second count that fits in a time.Duration.
// The next second overflows the conversion and used to look like a finished review.
func TestTimeoutDurationBound(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	const (
		largest = "9223372036"
		next    = "9223372037"
	)
	cases := []struct {
		name string
		args []string
		env  map[string]string
		want int
	}{
		{
			name: "flag largest",
			args: []string{"scan", "--sample", "--timeout=" + largest, target},
			want: exitOK,
		},
		{
			name: "flag next",
			args: []string{"scan", "--sample", "--timeout=" + next, target},
			want: exitConfig,
		},
		{
			name: "variable largest",
			args: []string{"scan", "--sample", target},
			env:  map[string]string{"PROVEN_TIMEOUT": largest},
			want: exitOK,
		},
		{
			name: "variable next",
			args: []string{"scan", "--sample", target},
			env:  map[string]string{"PROVEN_TIMEOUT": next},
			want: exitConfig,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			getenv := emptyEnv
			if tt.env != nil {
				getenv = envOf(tt.env)
			}
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, getenv, &stdout, &stderr, nil)
			if code != tt.want {
				t.Fatalf("exit %d, want %d\nstderr: %s\nstdout: %s", code, tt.want, stderr.String(), stdout.String())
			}
			if tt.want == exitConfig {
				if !strings.Contains(stderr.String(), largest) {
					t.Fatalf("stderr does not name %s: %q", largest, stderr.String())
				}
				return
			}
			assertDiscountText(t, stdout.String())
		})
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
	code := Run([]string{"scan", "somewhere"}, func(k string) string {
		switch k {
		case "METERLESS_RUNTIME_TOKEN":
			return "test-token"
		case "PROVEN_GUARDRAIL_MODE":
			return "off"
		case "ZETARIS_REST_URL":
			return "http://127.0.0.1:9"
		case "ZETARIS_API_KEY":
			return "test-key"
		case "ZETARIS_ORG_ID":
			return "1"
		default:
			return ""
		}
	}, &stdout, &stderr, nil)
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), review.ErrNotImplemented.Error()) {
		t.Fatalf("stderr still says the reviewer is missing\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "somewhere/base") {
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

func TestValidFlagBeatsInvalidVariable(t *testing.T) {
	finding := review.Result{Findings: []review.Finding{{
		RuleID:   "kept.rule",
		Severity: review.SeverityError,
		Message:  "kept",
		Location: review.Location{Path: "a.go", StartLine: 1},
	}}}
	cases := []struct {
		name     string
		args     []string
		env      map[string]string
		want     int
		deadline bool
		finding  bool
	}{
		{name: "text", args: []string{"scan", "--text"}, env: map[string]string{"PROVEN_OUTPUT_FORMAT": "xml"}, want: exitOK},
		{name: "json", args: []string{"scan", "--json"}, env: map[string]string{"PROVEN_OUTPUT_FORMAT": "xml"}, want: exitFatal},
		{name: "sarif", args: []string{"scan", "--sarif"}, env: map[string]string{"PROVEN_OUTPUT_FORMAT": "xml"}, want: exitFatal},
		{name: "error", args: []string{"scan", "--error"}, env: map[string]string{"PROVEN_ERROR": "maybe"}, want: exitFindings, finding: true},
		{name: "no-error", args: []string{"scan", "--no-error"}, env: map[string]string{"PROVEN_ERROR": "maybe"}, want: exitOK, finding: true},
		{name: "severity", args: []string{"scan", "--severity=INFO"}, env: map[string]string{"PROVEN_SEVERITY": "invalid"}, want: exitOK},
		{name: "timeout", args: []string{"scan", "--timeout=5"}, env: map[string]string{"PROVEN_TIMEOUT": "nope"}, want: exitOK, deadline: true},
		{name: "sample-off", args: []string{"scan", "--sample=false"}, env: map[string]string{"SAMPLE_MODE": "not-a-bool"}, want: exitOK},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rev := &captureReviewer{}
			if tt.finding {
				rev.res = finding
			}
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, envOf(tt.env), &stdout, &stderr, openCapture(rev))
			if code != tt.want {
				t.Fatalf("exit %d, want %d\nstderr: %s", code, tt.want, stderr.String())
			}
			for key := range tt.env {
				if strings.Contains(stderr.String(), key) {
					t.Fatalf("stderr names %s: %q", key, stderr.String())
				}
			}
			if tt.deadline && !rev.hasDeadline {
				t.Fatal("no deadline")
			}
		})
	}
}

func TestInvalidVariableAloneExits7(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "format", key: "PROVEN_OUTPUT_FORMAT", value: "xml"},
		{name: "error", key: "PROVEN_ERROR", value: "maybe"},
		{name: "severity", key: "PROVEN_SEVERITY", value: "invalid"},
		{name: "timeout", key: "PROVEN_TIMEOUT", value: "nope"},
		{name: "sample", key: "SAMPLE_MODE", value: "not-a-bool"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan"}, envOf(map[string]string{tt.key: tt.value}), &stdout, &stderr, nil)
			if code != exitConfig {
				t.Fatalf("exit %d stderr %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.key) {
				t.Fatalf("stderr %q", stderr.String())
			}
		})
	}
}

func TestSampleFlagIgnoresInvalidSampleMode(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", target}, envOf(map[string]string{"SAMPLE_MODE": "not-a-bool"}), &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("exit %d stderr %s\nstdout: %s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stderr.String(), "SAMPLE_MODE") {
		t.Fatalf("stderr %q", stderr.String())
	}
	assertDiscountText(t, stdout.String())
}

func TestConfigFlagIgnoresInvalidVariable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proven.yaml")
	if err := os.WriteFile(path, []byte("timeout: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--config=" + path}, envOf(map[string]string{"PROVEN_CONFIG": "auto"}), &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitOK {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "PROVEN_CONFIG") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

func TestCIMergeBaseIsTheForkPoint(t *testing.T) {
	withRefusingGlobalHook(t)
	dir, fork, baseTip, head := advancedBaseRepo(t, "main")
	event := writePullRequestEvent(t, dir, baseTip, head)
	t.Chdir(dir)
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
	if cap.req.BaselineCommit != fork {
		t.Fatalf("baseline %q, fork %q, base tip %q", cap.req.BaselineCommit, fork, baseTip)
	}
	if cap.req.BaselineCommit == baseTip {
		t.Fatalf("baseline is the base tip %s", baseTip)
	}
	if cap.req.HeadCommit != head {
		t.Fatalf("head %q", cap.req.HeadCommit)
	}
}

func TestCIUnresolvableBaseExits2(t *testing.T) {
	withRefusingGlobalHook(t)
	dir, head := singleCommitRepo(t)
	event := writePullRequestEvent(t, dir, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", head)
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"ci"}, func(k string) string {
		if k == "GITHUB_EVENT_PATH" {
			return event
		}
		return ""
	}, &stdout, &stderr, openCapture(&captureReviewer{}))
	if code != exitFatal {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "merge base") || !strings.Contains(stderr.String(), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatalf("stderr %q", stderr.String())
	}
}

// withRefusingGlobalHook points GIT_CONFIG_GLOBAL at a config whose
// core.hooksPath is a commit-msg hook that exits 1. The seat's own
// hooksPath is passed as command configuration and wins over that file,
// so the seat hook still judges fixture subjects.
func withRefusingGlobalHook(t *testing.T) {
	t.Helper()
	hooks := t.TempDir()
	script := filepath.Join(hooks, "commit-msg")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	body := "[core]\n\thooksPath = " + hooks + "\n"
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
}

func advancedBaseRepo(t *testing.T, baseBranch string) (dir, fork, baseTip, head string) {
	t.Helper()
	dir = t.TempDir()
	runGit(t, dir, "init", "-b", baseBranch, "--template=")
	fork = commitFile(t, dir, "f.txt", "root\n", "[cli-fixture] repo: root")
	runGit(t, dir, "branch", "feature")
	baseTip = commitFile(t, dir, "f.txt", "base advanced\n", "[cli-fixture] repo: advance base")
	runGit(t, dir, "checkout", "feature")
	head = commitFile(t, dir, "f.txt", "feature work\n", "[cli-fixture] repo: feature work")
	runGit(t, dir, "checkout", baseBranch)
	return dir, fork, baseTip, head
}

func singleCommitRepo(t *testing.T) (dir, sha string) {
	t.Helper()
	dir = t.TempDir()
	runGit(t, dir, "init", "-b", "main", "--template=")
	sha = commitFile(t, dir, "f.txt", "only\n", "[cli-fixture] repo: only")
	return dir, sha
}

func commitFile(t *testing.T, dir, name, body, message string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", name)
	runGit(t, dir, "-c", "commit.gpgsign=false", "-c", "user.name=Proven Test", "-c", "user.email=proven-test@example.com", "commit", "-m", message)
	return runGit(t, dir, "rev-parse", "HEAD")
}

func writePullRequestEvent(t *testing.T, dir, baseSHA, headSHA string) string {
	t.Helper()
	path := filepath.Join(dir, "event.json")
	body := fmt.Sprintf(`{"pull_request":{"base":{"ref":"main","sha":"%s"},"head":{"ref":"feature","sha":"%s"}}}`, baseSHA, headSHA)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = fixtureGitEnv(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// fixtureGitEnv keeps the process environment and replaces the host
// global and system git files. The seat's command configuration is kept,
// including core.hooksPath, so the commit-msg hook still judges fixture
// subjects.
func fixtureGitEnv(t *testing.T) []string {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(cfg, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	env := make([]string, 0, len(os.Environ())+8)
	for _, entry := range os.Environ() {
		if dropFixtureGitEnv(entry) {
			continue
		}
		env = append(env, entry)
	}
	return append(env,
		"GIT_CONFIG_GLOBAL="+cfg,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME=Proven Test",
		"GIT_AUTHOR_EMAIL=proven-test@example.com",
		"GIT_COMMITTER_NAME=Proven Test",
		"GIT_COMMITTER_EMAIL=proven-test@example.com",
	)
}

func dropFixtureGitEnv(entry string) bool {
	key, _, ok := strings.Cut(entry, "=")
	if !ok {
		return false
	}
	switch key {
	case "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_NOSYSTEM",
		"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL",
		"GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL":
		return true
	}
	return false
}
