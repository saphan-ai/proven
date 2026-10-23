package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

const versionLine = "proven development\n"

// opener loads a reviewer for a non-sample run under test.
// nil, and every sample run, uses review.Open.
type opener func(review.Options) (review.Reviewer, error)

// Run executes one proven invocation and returns the process exit code.
func Run(args []string, getenv func(string) string, stdout, stderr io.Writer, open opener) int {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	cfg, err := resolve(args, getenv, fileExists)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errorCode(err)
	}
	if cfg.help {
		fmt.Fprint(stdout, helpText())
		return exitOK
	}
	if cfg.version {
		fmt.Fprint(stdout, versionLine)
		return exitOK
	}
	if err := applyCI(&cfg, getenv); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errorCode(err)
	}
	if err := writable(cfg); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return exitFatal
	}
	rev, err := openCall(cfg, getenv, open)
	if err != nil {
		var conf *review.ConfigError
		if errors.As(err, &conf) {
			fmt.Fprintln(stderr, conf.Error())
			return exitConfig
		}
		fmt.Fprintln(stderr, err.Error())
		return exitFatal
	}
	ctx := context.Background()
	if cfg.timeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(cfg.timeoutSec)*time.Second)
		defer cancel()
	}
	result, err := reviewCall(ctx, rev, review.Request{
		Target:         cfg.target,
		BaselineCommit: cfg.baseline,
		HeadCommit:     cfg.head,
		Include:        cfg.include,
		Exclude:        cfg.exclude,
		MinSeverity:    cfg.severity,
	})
	if err != nil {
		var conf *review.ConfigError
		if errors.As(err, &conf) {
			fmt.Fprintln(stderr, conf.Error())
			return exitConfig
		}
		fmt.Fprintln(stderr, err.Error())
		return exitFatal
	}
	result.Findings = filterFindings(result.Findings, cfg.severity)
	if cfg.verbose && !cfg.quiet {
		fmt.Fprintf(stderr, "target: %s\nbaseline: %s\nhead: %s\n", cfg.target, cfg.baseline, cfg.head)
	}
	if err := writeOutputs(cfg, result, stdout); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return exitFatal
	}
	code := exitOK
	if cfg.errorOnFindings && len(result.Findings) > 0 {
		code = exitFindings
	}
	return code
}

func errorCode(err error) int {
	var ce *cliError
	if errors.As(err, &ce) {
		return ce.code
	}
	return exitFatal
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func writable(cfg config) error {
	names := []string{cfg.format}
	if cfg.textOutput != "" {
		names = append(names, "text")
	}
	if cfg.jsonOutput != "" {
		names = append(names, "json")
	}
	if cfg.sarifOutput != "" {
		names = append(names, "sarif")
	}
	for _, name := range names {
		if _, err := FormatterFor(name); err != nil {
			return err
		}
	}
	return nil
}

func openCall(cfg config, getenv func(string) string, open opener) (rev review.Reviewer, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("reviewer panicked: %v", rec)
		}
	}()
	return openReviewer(cfg, getenv, open)
}

func openReviewer(cfg config, getenv func(string) string, open opener) (review.Reviewer, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	opts := review.Options{
		ConfigPath: cfg.configPath,
		Sample:     cfg.sample,
		LogDir:     filepath.Join(wd, "logs"),
		Env:        getenv,
	}
	if cfg.sample || open == nil {
		return review.Open(opts)
	}
	return open(opts)
}

func reviewCall(ctx context.Context, rev review.Reviewer, req review.Request) (res review.Result, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("reviewer panicked: %v", rec)
		}
	}()
	return rev.Review(ctx, req)
}

func writeOutputs(cfg config, result review.Result, stdout io.Writer) error {
	primary, err := render(cfg.format, result)
	if err != nil {
		return err
	}
	type extraFile struct {
		path string
		body []byte
	}
	var extras []extraFile
	if cfg.textOutput != "" {
		body, err := render("text", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{cfg.textOutput, body})
	}
	if cfg.jsonOutput != "" {
		body, err := render("json", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{cfg.jsonOutput, body})
	}
	if cfg.sarifOutput != "" {
		body, err := render("sarif", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{cfg.sarifOutput, body})
	}
	if cfg.output == "" {
		if _, err := stdout.Write(primary); err != nil {
			return err
		}
	} else if err := os.WriteFile(cfg.output, primary, 0o644); err != nil {
		return err
	}
	for _, one := range extras {
		if err := os.WriteFile(one.path, one.body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

type pullRequestEvent struct {
	PullRequest struct {
		Base struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"base"`
		Head struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
	} `json:"pull_request"`
}

func applyCI(cfg *config, getenv func(string) string) error {
	if cfg.command != "ci" {
		return nil
	}
	var ev pullRequestEvent
	have := false
	if path := getenv("GITHUB_EVENT_PATH"); path != "" {
		body, err := os.ReadFile(path)
		if err != nil {
			return &cliError{code: exitFatal, msg: fmt.Sprintf("GITHUB_EVENT_PATH: %v", err)}
		}
		if err := json.Unmarshal(body, &ev); err != nil {
			return &cliError{code: exitFatal, msg: fmt.Sprintf("GITHUB_EVENT_PATH: %v", err)}
		}
		have = true
	}
	explicitBaseline := cfg.baseline != ""
	if cfg.baseline == "" {
		if have && ev.PullRequest.Base.SHA != "" {
			cfg.baseline = ev.PullRequest.Base.SHA
		} else if ref := getenv("GITHUB_BASE_REF"); ref != "" {
			cfg.baseline = ref
		}
	}
	if sha := getenv("GITHUB_SHA"); sha != "" {
		cfg.head = sha
	} else if have {
		cfg.head = ev.PullRequest.Head.SHA
	}
	if cfg.baseline == "" {
		// An empty baseline reviews the whole target. ci requires a pull-request baseline.
		if have {
			return &cliError{code: exitFatal, msg: "ci: the event has no base SHA, and --baseline-commit and GITHUB_BASE_REF are unset"}
		}
		return &cliError{code: exitFatal, msg: "ci: no pull-request baseline: set --baseline-commit, or GITHUB_EVENT_PATH with a base SHA, or GITHUB_BASE_REF"}
	}
	if cfg.head == "" {
		// An empty head is the working tree. ci would then review a different
		// checkout from the pull request it was asked for, including when the
		// baseline was passed explicitly and the merge base is not computed.
		// Exit 2, not 7. ADR-0020 §2 reserves 7 for a configuration refused at load.
		return &cliError{code: exitFatal, msg: "ci: no head commit: set GITHUB_SHA, or GITHUB_EVENT_PATH with a head SHA"}
	}
	if explicitBaseline {
		return nil
	}
	// Exit 2, not 7. ADR-0020 §2 reserves 7 for a configuration refused at load.
	// A base git cannot resolve means Proven could not review.
	base, err := gitMergeBase(cfg.target, cfg.baseline, cfg.head)
	if err != nil {
		return &cliError{code: exitFatal, msg: fmt.Sprintf("ci: cannot resolve the merge base of %s and %s: %s", cfg.baseline, cfg.head, err.Error())}
	}
	cfg.baseline = base
	return nil
}

func gitMergeBase(repo, base, head string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "merge-base", base, head)
	out, err := cmd.Output()
	if err != nil {
		detail := err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if msg := strings.TrimSpace(string(exitErr.Stderr)); msg != "" {
				detail = msg
			}
		}
		return "", errors.New(detail)
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("git merge-base returned no commit")
	}
	return sha, nil
}

func helpText() string {
	return `proven — review a change

usage:
  proven scan [flags] [TARGET]
  proven ci [flags]
  proven --help
  proven --version

TARGET defaults to ".". proven ci reads GITHUB_EVENT_PATH, GITHUB_SHA and
GITHUB_BASE_REF. When no baseline flag or variable is set, the baseline is
the merge base of the pull request base and the head. With no baseline
commit, no event base SHA and no GITHUB_BASE_REF, ci exits 2. A base that
cannot be resolved is exit 2: Proven could not review.

flags:
  --text, --json, --sarif
        format on standard output (PROVEN_OUTPUT_FORMAT, default text).
        This build writes text. --json and --sarif stop with exit 2 until their writers land.
  --text-output, --json-output, --sarif-output
        also write that format to a file
  -o, --output
        write the chosen format to a file instead of standard output (PROVEN_OUTPUT)
  --error, --no-error
        exit 1 when findings are reported (PROVEN_ERROR; scan default off, ci default on)
  --baseline-commit
        review only what changed since that commit (PROVEN_BASELINE_COMMIT, then SEMGREP_BASELINE_COMMIT, then SEMGREP_BASELINE_REF)
  --config
        a path to a Proven configuration file (PROVEN_CONFIG; proven.yaml in the target if present, else the defaults)
  --include, --exclude
        globs over changed paths (PROVEN_INCLUDE, PROVEN_EXCLUDE; comma-separated in the variable)
  --severity
        lowest severity reported: INFO, WARNING, ERROR (PROVEN_SEVERITY, default INFO)
  --timeout
        the whole review's time limit in seconds (PROVEN_TIMEOUT; default no deadline)
  --sample
        sample reviewer, no network and no keys (SAMPLE_MODE=true)
  --quiet, --verbose
  --help, --version

exit codes:
  0  review done; no findings, or findings without --error
  1  review done; findings, with --error (the default in ci)
  2  fatal error: proven could not review
  7  the configuration was refused when it loaded
`
}
