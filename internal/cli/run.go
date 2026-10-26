package cli

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/network"
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
	// verify checks a pack and runs no review (ADR-0035 point 3).
	if len(args) > 0 && args[0] == "verify" {
		return runVerify(args[1:], getenv, stdout, stderr)
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
	// The directory of each output inside the working directory is opened
	// here, after the checks and before the review runs, and the write goes
	// through that handle (openOutputDirectories).
	cfg.outputDirs = openOutputDirectories(cfg)
	defer cfg.outputDirs.close()
	// Every requested format must have a writer before the review runs: a
	// review that cannot be written is time spent for an exit 2 anyway.
	if err := writable(cfg); err != nil {
		printErr(stderr, err)
		return exitFatal
	}
	// A pack that cannot be signed is refused before the review, not after it.
	var packKey ed25519.PrivateKey
	if cfg.evidencePack != "" {
		packKey, err = loadPackKey(getenv)
		if err != nil {
			fmt.Fprintln(stderr, err.Error())
			return errorCode(err)
		}
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
	// The trace file takes every run's lines; the pack keeps this run's, which begin here.
	traceStart := int64(0)
	// The trace is in the log directory resolved at load (PROVEN_LOG_DIR), not
	// under the working directory.
	if cfg.evidencePack != "" && cfg.logDir != "" {
		traceStart = fileSize(filepath.Join(cfg.logDir, traceFile))
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
	// The pack records counts of the findings it holds, so they are made after the floor.
	result.ProofCounts = review.CountProofs(result.Findings)
	if cfg.verbose && !cfg.quiet {
		fmt.Fprintf(stderr, "target: %s\nbaseline: %s\nhead: %s\n", visible(cfg.target), visible(cfg.baseline), visible(cfg.head))
	}
	if cfg.evidencePack != "" {
		if err := writePack(cfg, getenv, packKey, result, traceStart); err != nil {
			fmt.Fprintln(stderr, "evidence pack not written: "+err.Error())
			return exitFatal
		}
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

// openReviewer builds the reviewer's options. The trace directory was chosen
// and refused, when it had to be, while the configuration loaded.
func openReviewer(cfg config, getenv func(string) string, open opener) (review.Reviewer, error) {
	opts := review.Options{
		ConfigPath: cfg.configPath,
		PolicyPath: cfg.policyPath,
		Sample:     cfg.sample,
		LogDir:     cfg.logDir,
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
	// The signed pack was written before this point; it is the first file of the
	// run, so a link made after the load check cannot lead an output onto it.
	if cfg.evidencePack != "" {
		if info, err := os.Stat(cfg.evidencePack); err == nil {
			written = append(written, writtenOutput{"--evidence-pack", info})
		}
	}
	if cfg.output == "" {
		// Standard output was written above; when it is a regular file it is
		// the first file of the run.
		if info := regularFileInfo(stdout); info != nil {
			written = append(written, writtenOutput{"standard output", info})
		}
	}
	for _, one := range files {
		dir := cfg.outputDirs[one.path]
		if info, err := statOutput(one.path, dir); err == nil {
			for _, w := range written {
				if os.SameFile(info, w.info) {
					return fmt.Errorf("%s and %s are the same file (%s); %s is left as it was written", w.flag, one.flag, one.path, w.flag)
				}
			}
		}
		if err := writeOutputFile(one.flag, one.path, one.body, dir); err != nil {
			return err
		}
		info, err := statOutput(one.path, dir)
		if err != nil {
			return fmt.Errorf("%s %s: %w", one.flag, one.path, err)
		}
		written = append(written, writtenOutput{one.flag, info})
	}
	return nil
}

// writeOutputFile writes body to path.
// WHY a final symbolic link is not followed: the open would replace the file
// the link names. The link is left in place and the write fails. A regular
// file is not a link and is overwritten as before.
//
// A path inside the working directory is written as a new file in the same
// directory, through that directory opened as a root, and the new file is
// renamed over the final name. The final name is replaced, never written
// through. A directory component swapped for a link after the load check
// cannot then name a file outside the directory: the root refuses the link
// and this returns an error that names the path. A path outside the working
// directory is the operator's and is opened as before.
func writeOutputFile(flag, path string, body []byte, dir *os.Root) error {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s %s: final component is a symbolic link; the file it names is not written", flag, path)
	}
	_, inside, err := outputRelativeToWorkingDirectory(path)
	if err != nil {
		return err
	}
	if inside && dir != nil {
		return writeOutputThroughDirectory(flag, path, dir, body)
	}
	if inside {
		// WHY no second resolution: a directory that did not exist when the
		// run started can be created during the review as a link to another
		// directory of the checkout, and a root follows that link.
		return fmt.Errorf("%s %s: its directory could not be opened when the run started; the report is not written", flag, path)
	}
	return writeOutputOutsideWorkingDirectory(flag, path, body)
}

// beforeRootedOutputCreate, when set, runs after the final-component
// check and before the rooted create. Tests replace that component in
// the window a rooted open would follow. Production leaves it nil.
var beforeRootedOutputCreate func(path string)

// statOutput reads the file at path through dir, the directory opened at
// load, when there is one: the path's spelling may name another directory by
// now, and the report is where the handle put it.
func statOutput(path string, dir *os.Root) (os.FileInfo, error) {
	if dir != nil {
		return dir.Stat(filepath.Base(path))
	}
	return os.Stat(path)
}

// outputDirectories holds, per output path inside the working directory,
// that path's directory opened at load.
type outputDirectories map[string]*os.Root

func (d outputDirectories) close() {
	for _, r := range d {
		r.Close()
	}
}

// openOutputDirectories opens the directory of each output that lies inside
// the working directory, through the working directory opened as a root.
//
// WHY at load: the review runs between the load check and the write, and
// the code under review can replace a directory of the output path with a
// link to another directory of the checkout. A root follows a link that
// stays inside it, so a write that resolves the path again would land in
// the directory the link names. A directory opened here is the directory
// the load check saw: a later rename or link at its name does not move the
// handle, and a directory removed meanwhile fails the write with its path.
// A directory that cannot be opened here (absent, or not a directory) is
// left out, and the write of that output fails with exit 2 naming the path.
func openOutputDirectories(cfg config) outputDirectories {
	dirs := outputDirectories{}
	wd, err := os.Getwd()
	if err != nil {
		return dirs
	}
	var wdRoot *os.Root
	for _, path := range []string{cfg.output, cfg.textOutput, cfg.jsonOutput, cfg.sarifOutput} {
		if path == "" || dirs[path] != nil {
			continue
		}
		rel, inside, err := outputRelativeToWorkingDirectory(path)
		if err != nil || !inside {
			continue
		}
		if wdRoot == nil {
			if wdRoot, err = os.OpenRoot(wd); err != nil {
				return dirs
			}
			defer wdRoot.Close()
		}
		// The directory part as spelled, not cleaned: "missing/.." must be
		// walked by the filesystem, which cannot pass through an absent name.
		dirPart, _ := filepath.Split(rel)
		dirPart = strings.TrimRight(dirPart, string(filepath.Separator))
		if dirPart == "" {
			dirPart = "."
		}
		r, err := wdRoot.OpenRoot(dirPart)
		if err != nil {
			continue
		}
		dirs[path] = r
	}
	return dirs
}

// writeOutputThroughDirectory writes body as the final name of path in dir,
// the directory opened at load: a new file created there with O_EXCL, then
// renamed over the final name.
//
// WHY the final name is renamed over, not opened: a root follows a symbolic
// link that stays inside it, so opening the final name would replace the
// file a link there names, including a link swapped in after the check.
// Rename replaces the directory entry, so a link there is not written
// through. The new file is removed if the write or the rename fails.
func writeOutputThroughDirectory(flag, path string, dir *os.Root, body []byte) error {
	if beforeRootedOutputCreate != nil {
		beforeRootedOutputCreate(path)
	}
	name := filepath.Base(path)
	tmp, f, err := createExclusiveOutput(dir, name)
	if err != nil {
		return fmt.Errorf("%s %s: %w", flag, path, err)
	}
	_, werr := f.Write(body)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		werr = discardOutputTemp(dir, tmp, werr)
		return fmt.Errorf("%s %s: %w", flag, path, werr)
	}
	if err := dir.Rename(tmp, name); err != nil {
		err = discardOutputTemp(dir, tmp, err)
		return fmt.Errorf("%s %s: %w", flag, path, err)
	}
	return nil
}

// outputTempAttempts bounds how many exclusive names are tried in the
// output's directory before the write fails.
const outputTempAttempts = 8

// createExclusiveOutput creates a new file in the same directory as rel.
// The name is not rel: O_EXCL refuses a name that already exists, and the
// caller renames this file over rel.
func createExclusiveOutput(root *os.Root, rel string) (string, *os.File, error) {
	dir, _ := filepath.Split(rel)
	var last error
	for range outputTempAttempts {
		name, err := exclusiveOutputName(dir)
		if err != nil {
			return "", nil, err
		}
		f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			return name, f, nil
		}
		if !os.IsExist(err) {
			return "", nil, err
		}
		last = err
	}
	return "", nil, last
}

// exclusiveOutputName is a relative path in dir that is not the final output
// name. dir is empty or ends with a separator, as filepath.Split returns it.
func exclusiveOutputName(dir string) (string, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return dir + ".proven-output." + hex.EncodeToString(buf[:]), nil
}

// discardOutputTemp removes the new file after a failed write or rename.
// A failed removal is part of the error, so the name does not disappear
// from the report.
func discardOutputTemp(root *os.Root, name string, err error) error {
	if rerr := root.Remove(name); rerr != nil {
		return fmt.Errorf("%w (removing temporary file %s: %v)", err, name, rerr)
	}
	return err
}

// writeOutputOutsideWorkingDirectory writes a path that does not lie in the
// working directory. The operator chose it. A final symbolic link is not
// followed where the system can refuse one. An earlier component is followed.
func writeOutputOutsideWorkingDirectory(flag, path string, body []byte) error {
	f, err := openOutputFile(path)
	if err != nil {
		// ELOOP from a final link swapped in after the Lstat. An ELOOP from
		// an intermediate loop is not that case: Lstat of the final component
		// would not show a symlink, and the open error stands.
		if openFailedOnFinalLink(err) {
			again, lerr := os.Lstat(path)
			if lerr == nil && again.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("%s %s: final component is a symbolic link; the file it names is not written", flag, path)
			}
		}
		return err
	}
	_, werr := f.Write(body)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	return werr
}

// outputRelativeToWorkingDirectory reports path relative to the working
// directory when the path lies inside it. inside is false when the path
// leaves the working directory; the caller writes that path as the operator's.
//
// A relative path lies inside when ".." never climbs out. An absolute path
// lies inside when a prefix resolves to the working directory. Components
// after that prefix are not resolved here: a link among them stays in the
// relative path, and the root open refuses one that leaves the directory.
func outputRelativeToWorkingDirectory(path string) (rel string, inside bool, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("working directory: %w", err)
	}
	if !filepath.IsAbs(path) {
		if !relativePathStaysInWorkingDirectory(path) {
			return "", false, nil
		}
		return path, true, nil
	}
	resolvedWD := resolvedPath(wd)
	root, rest := splitRoot(path)
	parts := splitComponents(rest)
	cur := root
	for i, name := range parts {
		if name == "." {
			continue
		}
		if name == ".." {
			parent := filepath.Dir(cur)
			if parent != cur {
				cur = parent
			}
			continue
		}
		if resolvedPath(cur) == resolvedWD {
			rel = filepath.Join(append([]string{name}, parts[i+1:]...)...)
			return rel, rel != "", nil
		}
		cur = filepath.Join(cur, name)
	}
	return "", false, nil
}

// relativePathStaysInWorkingDirectory reports whether a relative path climbs
// out of the working directory with "..". The path is not cleaned first:
// "link/../file" still names link, so the root open can refuse it.
func relativePathStaysInWorkingDirectory(path string) bool {
	depth := 0
	for _, name := range splitComponents(path) {
		switch name {
		case ".":
		case "..":
			if depth == 0 {
				return false
			}
			depth--
		default:
			depth++
		}
	}
	return depth > 0
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
	// The merge base is read before review.Open loads the mode, so git is told here.
	// A wrong network setting is not judged here: Open refuses it by name. Until
	// then git runs offline, so a setting that does not load cannot open a remote.
	pol, err := network.Load(getenv)
	defer change.UseOffline(change.UseOffline(err != nil || pol.Local()))
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
	// Property: no key reaches a subprocess (ADR-0025). GitCommand builds the
	// minimal environment and, in local mode, a git that cannot fetch.
	cmd := change.GitCommand("-C", repo, "merge-base", base, head)
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
  proven verify [--public-key FILE] PACK
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
  --evidence-pack
        write one signed evidence pack (a tar) for this review; the private key is the file PROVEN_PACK_KEY names.
        proven verify checks it with the public key in PROVEN_PACK_PUBLIC_KEY, or --public-key
  PROVEN_LOG_DIR
        trace directory (default: the user cache, proven/logs, mode 0700).
        A directory that resolves inside the review target is refused.
  --sample
        sample reviewer, no network and no keys (SAMPLE_MODE=true)
  --quiet, --verbose
  --help, --version

environment only:
  PROVEN_POLICY
        a path to the organisation policy file, read before proven.yaml (docs/POLICY.md)

exit codes:
  0  review done; no findings, or findings without --error
  1  review done; findings, with --error (the default in ci)
  2  fatal error: proven could not review
  7  the configuration was refused when it loaded
`
}
