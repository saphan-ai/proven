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

	"github.com/saphan-ai/proven/internal/childenv"
	"github.com/saphan-ai/proven/internal/review"
)

// versionLine is what --version prints until releases are stamped at build time.
const versionLine = "proven development\n"

// opener loads a reviewer for a non-sample run under test.
// nil, and every sample run, uses review.Open.
type opener func(review.Options) (review.Reviewer, error)

// Run executes one proven invocation and returns the process exit code
// (ADR-0020 §2: 0 reviewed, 1 findings under --error, 2 could not review,
// 7 configuration refused). Every input arrives as a parameter, never from
// the process, so a test drives exactly what the binary would.
func Run(args []string, getenv func(string) string, stdout, stderr io.Writer, open opener) int {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	// Settings are resolved and refused before anything is opened, so a bad
	// value costs nothing and its exit code is the one ADR-0020 §2 names.
	cfg, err := resolve(args, getenv, fileExists)
	if err != nil {
		printErr(stderr, err)
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
		printErr(stderr, err)
		return errorCode(err)
	}
	// Standard output is compared with the output files here, not in resolve,
	// because only Run holds the writer it is open on.
	if err := refuseStdoutOnAnOutput(cfg, stdout); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errorCode(err)
	}
	// After the same-file checks: a final symlink that is not already the
	// same file as another output, or as standard output, is refused here.
	// Those checks name both sides. This one names the flag and the path.
	if err := refuseSymlinkOutputs(cfg); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return errorCode(err)
	}
	// Every requested format must have a writer before the review runs: a
	// review that cannot be written is time spent for an exit 2 anyway.
	if err := writable(cfg); err != nil {
		printErr(stderr, err)
		return exitFatal
	}
	rev, err := openCall(cfg, getenv, open)
	if err != nil {
		// A setting the reviewer refuses (its own keys, its proven.yaml) is a
		// configuration refused at load, exit 7; anything else is exit 2.
		var conf *review.ConfigError
		if errors.As(err, &conf) {
			printErr(stderr, conf)
			return exitConfig
		}
		printErr(stderr, err)
		return exitFatal
	}
	// --timeout bounds the whole review; zero means no deadline (ADR-0020 §1).
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
			printErr(stderr, conf)
			return exitConfig
		}
		printErr(stderr, err)
		return exitFatal
	}
	// The severity floor is applied here as well as in the request, so the
	// outputs and the exit code agree whatever a reviewer returns.
	result.Findings = filterFindings(result.Findings, cfg.severity)
	if cfg.verbose && !cfg.quiet {
		fmt.Fprintf(stderr, "target: %s\nbaseline: %s\nhead: %s\n", visible(cfg.target), visible(cfg.baseline), visible(cfg.head))
	}
	if err := writeOutputs(cfg, result, stdout); err != nil {
		printErr(stderr, err)
		return exitFatal
	}
	// A failed outside service is named in the output and is not exit 2:
	// the review still ran on what it could reach (ADR-0020 §2).
	code := exitOK
	if cfg.errorOnFindings && len(result.Findings) > 0 {
		code = exitFindings
	}
	return code
}

// printErr writes an error sentence to stderr through the encoder the text
// output uses: errors name files from the change, and stderr is a terminal too.
func printErr(stderr io.Writer, err error) {
	fmt.Fprintln(stderr, visible(err.Error()))
}

// errorCode is the exit code an error carries; an error that carries none
// means Proven could not review, exit 2.
func errorCode(err error) int {
	var ce *cliError
	if errors.As(err, &ce) {
		return ce.code
	}
	return exitFatal
}

// fileExists reports whether path is a regular file; resolve takes it as a
// parameter so tests can decide which configuration files exist.
func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// writable refuses before the review when any requested format has no writer.
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

// openCall opens the reviewer and turns a panic into an error, so a broken
// reviewer is exit 2 with a sentence, never a Go stack trace (ADR-0020 §2).
func openCall(cfg config, getenv func(string) string, open opener) (rev review.Reviewer, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("reviewer panicked: %v", rec)
		}
	}()
	return openReviewer(cfg, getenv, open)
}

// openReviewer builds the reviewer's options. The trace goes to logs/ in
// the directory the command starts from, which is where docs/SAMPLE_MODE.md
// tells a reader to look.
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
	// Sample mode always takes the real review.Open, never a test's opener:
	// the sample reviewer is the thing under test, with no network and no
	// keys (ADR-0001).
	if cfg.sample || open == nil {
		return review.Open(opts)
	}
	return open(opts)
}

// reviewCall runs the review and turns a panic into an error, as openCall does.
func reviewCall(ctx context.Context, rev review.Reviewer, req review.Request) (res review.Result, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("reviewer panicked: %v", rec)
		}
	}()
	return rev.Review(ctx, req)
}

// writeOutputs writes the chosen format to standard output or -o, and each
// --*-output file besides, all from the one result (ADR-0020 §1). Every body
// is rendered before any file is written, so a format that fails leaves no
// half-written set behind.
func writeOutputs(cfg config, result review.Result, stdout io.Writer) error {
	primary, err := render(cfg.format, result)
	if err != nil {
		return err
	}
	type extraFile struct {
		flag string
		path string
		body []byte
	}
	var extras []extraFile
	if cfg.textOutput != "" {
		body, err := render("text", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{"--text-output", cfg.textOutput, body})
	}
	if cfg.jsonOutput != "" {
		body, err := render("json", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{"--json-output", cfg.jsonOutput, body})
	}
	if cfg.sarifOutput != "" {
		body, err := render("sarif", result)
		if err != nil {
			return err
		}
		extras = append(extras, extraFile{"--sarif-output", cfg.sarifOutput, body})
	}
	files := extras
	if cfg.output == "" {
		if _, err := stdout.Write(primary); err != nil {
			return err
		}
	} else {
		files = append([]extraFile{{"-o", cfg.output, primary}}, extras...)
	}
	// WHY a second check beside refuseCollidingOutputs: that one runs at load,
	// and the disk can change before the write (a link made while the review
	// ran). Two outputs must never share one file, so each file is compared
	// with the files already written in this run before it is written, and the
	// run stops with the first output intact.
	var written []writtenOutput
	if cfg.output == "" {
		// Standard output was written above; when it is a regular file it is
		// the first file of the run.
		if info := regularFileInfo(stdout); info != nil {
			written = append(written, writtenOutput{"standard output", info})
		}
	}
	for _, one := range files {
		if info, err := os.Stat(one.path); err == nil {
			for _, w := range written {
				if os.SameFile(info, w.info) {
					return fmt.Errorf("%s and %s are the same file (%s); %s is left as it was written", w.flag, one.flag, one.path, w.flag)
				}
			}
		}
		if err := writeOutputFile(one.flag, one.path, one.body); err != nil {
			return err
		}
		info, err := os.Stat(one.path)
		if err != nil {
			return err
		}
		written = append(written, writtenOutput{one.flag, info})
	}
	return nil
}

// writeOutputFile writes body to path.
// WHY a final symbolic link is not followed: os.WriteFile replaces the file
// the link names. In proven ci the working directory is the checkout, so a
// link the change committed would steer the report onto that other file.
// The link is left in place and the write fails. A regular file is not a
// link, and os.WriteFile overwrites it as before. A link that is only a
// directory on the way to the file is not the final component, so Lstat
// does not see it as a link and the file is created inside the target.
func writeOutputFile(flag, path string, body []byte) error {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s %s: final component is a symbolic link; the file it names is not written", flag, path)
	}
	return os.WriteFile(path, body, 0o644)
}

// writtenOutput is a file this run has written, kept for os.SameFile.
type writtenOutput struct {
	flag string
	info os.FileInfo
}

// pullRequestEvent is the part of GitHub's pull_request event payload ci
// reads. Gitea is an equal target (ADR-0022); that its payload carries the
// same fields is still to be measured on the lab's Gitea.
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

// applyCI fills the baseline and head of a `proven ci` run from the pull
// request it runs for. The baseline comes from --baseline-commit, else the
// event's base SHA, else GITHUB_BASE_REF; the head from GITHUB_SHA, else
// the event's head SHA. ci never falls back to reviewing the whole
// checkout: without a baseline or a head it exits 2 and names what is
// missing.
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
	// A baseline the caller named is used as given. One found in the event
	// or GITHUB_BASE_REF is the branch the pull request targets, so the
	// review starts from where the change forked from it, not its tip.
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

// gitMergeBase asks git for the fork point of base and head in repo and
// returns git's own sentence when it cannot answer.
func gitMergeBase(repo, base, head string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "merge-base", base, head)
	// Property: no key reaches a subprocess. An inherited environment would
	// hand git, and any hook or alias it runs, every key Proven holds; git
	// needs PATH and HOME only (ADR-0025).
	cmd.Env = childenv.Minimal()
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

// helpText is the --help page: the flags of ADR-0020 §1 with the variable
// beside each, and the exit codes of §2.
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
commit, no event base SHA and no GITHUB_BASE_REF, ci exits 2.
With no head commit, ci exits 2: set GITHUB_SHA, or GITHUB_EVENT_PATH
with a head SHA. A base that cannot be resolved is exit 2: Proven could
not review.

flags:
  --text, --json, --sarif
        format on standard output (PROVEN_OUTPUT_FORMAT, default text).
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
