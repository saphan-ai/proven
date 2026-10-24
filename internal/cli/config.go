package cli

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

// maxTimeoutSec is the largest whole number of seconds that fits in a time.Duration.
// math.MaxInt64 / time.Second is 9223372036; the next second overflows the conversion.
const maxTimeoutSec = int64(math.MaxInt64 / time.Second)

// The exit codes of ADR-0020 §2, Semgrep's own, so a pipeline that branches
// on them needs no change.
const (
	exitOK       = 0 // reviewed; no findings, or findings without --error
	exitFindings = 1 // reviewed; findings, with --error
	exitFatal    = 2 // Proven could not review
	exitConfig   = 7 // a setting was refused when it loaded
)

// KnownCLIEnv reports whether name is an environment variable ADR-0020 §1
// documents for the command line. The table names PROVEN_OUTPUT_FORMAT,
// PROVEN_OUTPUT, PROVEN_ERROR, PROVEN_BASELINE_COMMIT, SEMGREP_BASELINE_COMMIT,
// SEMGREP_BASELINE_REF, PROVEN_CONFIG, PROVEN_INCLUDE, PROVEN_EXCLUDE,
// PROVEN_SEVERITY, PROVEN_TIMEOUT and SAMPLE_MODE. That section's ci paragraph
// names GITHUB_EVENT_PATH, GITHUB_BASE_REF and GITHUB_SHA. GITHUB_WORKSPACE is
// not in the section. Adapter secrets are not in this set.
func KnownCLIEnv(name string) bool {
	switch name {
	case "GITHUB_BASE_REF",
		"GITHUB_EVENT_PATH",
		"GITHUB_SHA",
		"PROVEN_BASELINE_COMMIT",
		"PROVEN_CONFIG",
		"PROVEN_ERROR",
		"PROVEN_EXCLUDE",
		"PROVEN_INCLUDE",
		"PROVEN_OUTPUT",
		"PROVEN_OUTPUT_FORMAT",
		"PROVEN_SEVERITY",
		"PROVEN_TIMEOUT",
		"SAMPLE_MODE",
		"SEMGREP_BASELINE_COMMIT",
		"SEMGREP_BASELINE_REF":
		return true
	default:
		return false
	}
}

// config is one resolved invocation. Flag wins over variable, variable over default.
type config struct {
	command         string
	target          string
	format          string
	textOutput      string
	jsonOutput      string
	sarifOutput     string
	output          string
	errorOnFindings bool
	baseline        string
	configPath      string
	include         []string
	exclude         []string
	severity        review.Severity
	timeoutSec      int
	sample          bool
	quiet           bool
	verbose         bool
	help            bool
	version         bool
	head            string
}

// cliError is a refusal that carries its own exit code.
type cliError struct {
	code int
	msg  string
}

func (e *cliError) Error() string { return e.msg }

// boolPresence is a boolean flag that remembers whether it was given at
// all, because an absent flag must fall through to its variable while an
// explicit --sample=false must not.
type boolPresence struct {
	name string
	set  bool
	v    bool
	// bad is the text Set refused. The flag package would report that as a
	// usage error, which exits 2 and spells the flag with one dash.
	// refused is a separate bit because the text can be empty: Set("")
	// fails ParseBool, and an empty bad is also how a flag that was never
	// set looks. resolve turns a refusal into the same exit-7 answer as
	// the variable.
	bad     string
	refused bool
}

func (b *boolPresence) String() string {
	if b != nil && b.v {
		return "true"
	}
	return "false"
}

func (b *boolPresence) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		b.bad = s
		b.refused = true
		return err
	}
	b.set = true
	b.v = v
	b.bad = ""
	b.refused = false
	return nil
}

func (b *boolPresence) IsBoolFlag() bool { return true }

// setString is a string flag that remembers whether it was given, so an
// explicit empty value still wins over the variable.
type setString struct {
	set bool
	v   string
}

func (s *setString) String() string {
	if s == nil {
		return ""
	}
	return s.v
}

func (s *setString) Set(v string) error {
	s.set = true
	s.v = v
	return nil
}

// multiFlag is a repeatable flag (--include, --exclude), each use one glob,
// as Semgrep takes them.
type multiFlag struct {
	set bool
	v   []string
}

func (m *multiFlag) String() string { return strings.Join(m.v, ",") }

func (m *multiFlag) Set(v string) error {
	m.set = true
	m.v = append(m.v, v)
	return nil
}

// holders is every flag of one parse, before the variables and defaults are applied.
type holders struct {
	text, asJSON, sarif                 boolPresence
	errorOn, noError                    boolPresence
	sample, quiet, verbose, help, ver   boolPresence
	textOutput, jsonOutput, sarifOutput setString
	baseline, configPath, severity      setString
	timeout, output                     setString
	include, exclude                    multiFlag
}

// knownFlagNames is the list an unknown flag's refusal prints, so the
// reader sees what is allowed instead of a usage page.
func knownFlagNames() []string {
	return []string{
		"--baseline-commit",
		"--config",
		"--error",
		"--exclude",
		"--help",
		"--include",
		"--json",
		"--json-output",
		"--no-error",
		"--output",
		"--quiet",
		"--sample",
		"--sarif",
		"--sarif-output",
		"--severity",
		"--text",
		"--text-output",
		"--timeout",
		"--verbose",
		"--version",
		"-o",
	}
}

// defineFlags builds the flag set of ADR-0020 §1. Its output is discarded
// and its usage silenced: every refusal is written by resolve, with the
// allowed values, never by the flag package.
func defineFlags() (*flag.FlagSet, *holders) {
	h := &holders{}
	// Set is not told which flag it is setting. The name is the spelling
	// the refusal prints; the flag package's own error uses one dash.
	h.text.name = "--text"
	h.asJSON.name = "--json"
	h.sarif.name = "--sarif"
	h.errorOn.name = "--error"
	h.noError.name = "--no-error"
	h.sample.name = "--sample"
	h.quiet.name = "--quiet"
	h.verbose.name = "--verbose"
	h.help.name = "--help"
	h.ver.name = "--version"
	fs := flag.NewFlagSet("proven", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	fs.Var(&h.text, "text", "text on standard output")
	fs.Var(&h.asJSON, "json", "json on standard output")
	fs.Var(&h.sarif, "sarif", "sarif on standard output")
	fs.Var(&h.textOutput, "text-output", "also write text to this file")
	fs.Var(&h.jsonOutput, "json-output", "also write json to this file")
	fs.Var(&h.sarifOutput, "sarif-output", "also write sarif to this file")
	fs.Var(&h.output, "output", "write the chosen format to this file")
	fs.Var(&h.output, "o", "write the chosen format to this file")
	fs.Var(&h.errorOn, "error", "exit 1 when findings are reported")
	fs.Var(&h.noError, "no-error", "do not exit 1 when findings are reported")
	fs.Var(&h.baseline, "baseline-commit", "review only changes since this commit")
	fs.Var(&h.configPath, "config", "path to Proven's configuration file")
	fs.Var(&h.include, "include", "glob of changed paths to review")
	fs.Var(&h.exclude, "exclude", "glob of changed paths to skip")
	fs.Var(&h.severity, "severity", "lowest severity reported")
	fs.Var(&h.timeout, "timeout", "review time limit in seconds")
	fs.Var(&h.sample, "sample", "use the sample reviewer")
	fs.Var(&h.quiet, "quiet", "suppress the progress line")
	fs.Var(&h.verbose, "verbose", "print the request on standard error")
	fs.Var(&h.help, "help", "show help")
	fs.Var(&h.ver, "version", "show the version")
	return fs, h
}

// refusedBoolFlag names the boolean flag whose value the parse refused, if any.
func refusedBoolFlag(h *holders) (string, string, bool) {
	flags := []*boolPresence{
		&h.text, &h.asJSON, &h.sarif, &h.errorOn, &h.noError,
		&h.sample, &h.quiet, &h.verbose, &h.help, &h.ver,
	}
	for _, b := range flags {
		// An empty refusal must still count. bad == "" is also an unset flag.
		if b.refused {
			return b.name, b.bad, true
		}
	}
	return "", "", false
}

// isValueFlag reports whether a flag takes the next argument as its value,
// which arrange needs to keep a value from being read as the command.
func isValueFlag(name string) bool {
	switch name {
	case "text-output", "json-output", "sarif-output", "output", "o", "baseline-commit", "config", "include", "exclude", "severity", "timeout":
		return true
	default:
		return false
	}
}

// arrange pulls the command out of args so flags may sit before or after it.
// Flag values stay with their flags. Everything else, in order, is the command
// and then the positionals.
func arrange(args []string) (cmd string, flagArgs []string, positionals []string) {
	var pos []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		name, ok := flagToken(args[i])
		if !ok {
			pos = append(pos, args[i])
			continue
		}
		flagArgs = append(flagArgs, args[i])
		if isValueFlag(name) && !strings.Contains(args[i], "=") && i+1 < len(args) {
			i++
			flagArgs = append(flagArgs, args[i])
		}
	}
	if len(pos) > 0 {
		cmd = pos[0]
		positionals = pos[1:]
	}
	return cmd, flagArgs, positionals
}

// flagToken returns the flag name in arg, without dashes or value, and
// whether arg is a flag at all ("-" alone is a path, not a flag).
func flagToken(arg string) (string, bool) {
	if arg == "-" || !strings.HasPrefix(arg, "-") {
		return "", false
	}
	name := strings.TrimLeft(arg, "-")
	if name == "" {
		return "", false
	}
	if i := strings.IndexByte(name, '='); i >= 0 {
		name = name[:i]
	}
	return name, true
}

// resolve turns the arguments and the environment into one config, flag
// over variable over default, and refuses a value outside its allowed set
// with the allowed list (ADR-0013). Nothing here reads a secret or opens a
// connection: it only decides what the run will ask for.
func resolve(args []string, getenv func(string) string, exists func(string) bool) (config, error) {
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	if exists == nil {
		exists = func(string) bool { return false }
	}
	cmd, flagArgs, positionals := arrange(args)
	fs, h := defineFlags()
	if err := fs.Parse(flagArgs); err != nil {
		if strings.Contains(err.Error(), "not defined") {
			return config{}, &cliError{code: exitFatal, msg: err.Error() + "\nknown flags: " + strings.Join(knownFlagNames(), ", ")}
		}
		// ADR-0020 §2: a value that is not true or false is a configuration
		// refused at load, the same whether it came from the flag or the variable.
		if name, value, ok := refusedBoolFlag(h); ok {
			return config{}, refuse(name, value, "value must be true or false", []string{"true", "false"})
		}
		return config{}, &cliError{code: exitFatal, msg: err.Error()}
	}
	// --help and --version answer without a command, like any CLI.
	if (h.help.set && h.help.v) || (h.ver.set && h.ver.v) {
		return config{help: h.help.set && h.help.v, version: h.ver.set && h.ver.v}, nil
	}
	if cmd == "" {
		return config{}, &cliError{code: exitFatal, msg: "missing command"}
	}
	cfg := config{command: cmd, format: "text", severity: review.SeverityInfo}
	switch cmd {
	case "scan":
		if len(positionals) > 1 {
			return config{}, &cliError{code: exitFatal, msg: "scan takes at most one target"}
		}
		cfg.target = "."
		if len(positionals) == 1 {
			cfg.target = positionals[0]
		}
	case "ci":
		if len(positionals) != 0 {
			return config{}, &cliError{code: exitFatal, msg: "ci takes no target"}
		}
		// proven.yaml is resolved from cfg.target below. ci has no positional
		// target, so the checkout is GITHUB_WORKSPACE, including when this
		// process was started outside that directory.
		cfg.target = "."
		if ws := getenv("GITHUB_WORKSPACE"); ws != "" {
			cfg.target = ws
		}
	default:
		return config{}, &cliError{code: exitFatal, msg: fmt.Sprintf("unknown command %q", cmd)}
	}
	if err := refuseExclusive(h); err != nil {
		return config{}, err
	}
	format, err := resolveFormat(getenv, h)
	if err != nil {
		return config{}, err
	}
	cfg.format = format
	cfg.textOutput = h.textOutput.v
	cfg.jsonOutput = h.jsonOutput.v
	cfg.sarifOutput = h.sarifOutput.v
	if h.output.set {
		cfg.output = h.output.v
	} else {
		cfg.output = getenv("PROVEN_OUTPUT")
	}
	if err := refuseCollidingOutputs(cfg); err != nil {
		return config{}, err
	}
	errorOn, err := resolveError(cmd, getenv, h)
	if err != nil {
		return config{}, err
	}
	cfg.errorOnFindings = errorOn
	cfg.baseline = resolveBaseline(getenv, h)
	configPath, err := resolveConfig(cfg.target, getenv, exists, h)
	if err != nil {
		return config{}, err
	}
	cfg.configPath = configPath
	if h.include.set {
		cfg.include = append([]string{}, h.include.v...)
	} else {
		cfg.include = splitList(getenv("PROVEN_INCLUDE"))
	}
	if h.exclude.set {
		cfg.exclude = append([]string{}, h.exclude.v...)
	} else {
		cfg.exclude = splitList(getenv("PROVEN_EXCLUDE"))
	}
	severity, err := resolveSeverity(getenv, h)
	if err != nil {
		return config{}, err
	}
	cfg.severity = severity
	timeout, err := resolveTimeout(getenv, h)
	if err != nil {
		return config{}, err
	}
	cfg.timeoutSec = timeout
	sample, err := resolveSample(getenv, h)
	if err != nil {
		return config{}, err
	}
	cfg.sample = sample
	cfg.quiet = h.quiet.set && h.quiet.v
	cfg.verbose = h.verbose.set && h.verbose.v
	return cfg, nil
}

// refuseCollidingOutputs refuses two outputs that are one file, at load
// (ADR-0013). The files are written one after another, so the last write would
// silently replace the others and the run would still exit 0 (ADR-0020 §1).
//
// What counts as one file: the same path after the filesystem's own resolution
// of the raw spelling, links before ".." (outputIdentity), or two existing files for which os.SameFile is true (a
// hard link, or a link to the file itself). Links are resolved because a
// directory link such as link -> real makes "link/r.json" and "real/r.json" the
// same file under two spellings, and a check that compared spellings would let
// the second write replace the first. Only stat calls happen here; nothing is
// created. Two spellings of a file that does not exist yet and differs only in
// letter case on a case-insensitive volume are not caught: no I/O can tell them
// apart before the file is made.
func refuseCollidingOutputs(cfg config) error {
	named := []struct{ flag, path string }{
		{"-o", cfg.output},
		{"--text-output", cfg.textOutput},
		{"--json-output", cfg.jsonOutput},
		{"--sarif-output", cfg.sarifOutput},
	}
	type seenOutput struct {
		flag string
		info os.FileInfo
	}
	byIdentity := map[string]string{} // identity -> the flag that named it first
	var existing []seenOutput         // outputs whose file exists, for os.SameFile
	for _, n := range named {
		if n.path == "" {
			continue
		}
		identity, err := outputIdentity(n.path)
		if err != nil {
			return refuse(n.flag, n.path, err.Error(), nil)
		}
		if first, ok := byIdentity[identity]; ok {
			return refuse(n.flag, identity, "same file as "+first+"; each output needs its own path", nil)
		}
		byIdentity[identity] = n.flag
		info, err := os.Stat(n.path)
		if err != nil {
			continue
		}
		for _, other := range existing {
			if os.SameFile(info, other.info) {
				return refuse(n.flag, identity, "same file as "+other.flag+"; each output needs its own path", nil)
			}
		}
		existing = append(existing, seenOutput{flag: n.flag, info: info})
	}
	return nil
}

// refuseSymlinkOutputs refuses an output whose final component is a symbolic
// link, before the review and before any file is written (exit 7).
// WHY the link is not followed: os.WriteFile would replace the file the link
// names. Run calls this after refuseCollidingOutputs and
// refuseStdoutOnAnOutput, so a link that is the same file as another output,
// or as standard output, is still refused as that same file, with both sides
// named. A regular file is not a link and is overwritten at the write. A
// symbolic link earlier in the path is a directory on the way and is still
// resolved by that same-file check.
func refuseSymlinkOutputs(cfg config) error {
	named := []struct{ flag, path string }{
		{"-o", cfg.output},
		{"--text-output", cfg.textOutput},
		{"--json-output", cfg.jsonOutput},
		{"--sarif-output", cfg.sarifOutput},
	}
	for _, n := range named {
		if n.path == "" {
			continue
		}
		info, err := os.Lstat(n.path)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return refuse(n.flag, n.path, "final component is a symbolic link; the file it names is not written", nil)
		}
	}
	return nil
}

// regularFileInfo is the FileInfo of w when w is an *os.File on a regular
// file, and nil for a terminal, a pipe, a buffer or anything that cannot be
// compared with an output path.
func regularFileInfo(w io.Writer) os.FileInfo {
	f, ok := w.(*os.File)
	if !ok {
		return nil
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}
	return info
}

// refuseStdoutOnAnOutput refuses an output file that is the file standard
// output is already redirected onto (ADR-0013), at load, before the review
// runs. The shell creates the redirect target before Proven starts, so an
// output that is that file exists under some spelling and os.Stat, which
// follows links, finds it; os.SameFile then compares the two. A terminal or a
// pipe on standard output is not a regular file and is never a collision.
func refuseStdoutOnAnOutput(cfg config, stdout io.Writer) error {
	std := regularFileInfo(stdout)
	if std == nil {
		return nil
	}
	named := []struct{ flag, path string }{
		{"-o", cfg.output},
		{"--text-output", cfg.textOutput},
		{"--json-output", cfg.jsonOutput},
		{"--sarif-output", cfg.sarifOutput},
	}
	for _, n := range named {
		if n.path == "" {
			continue
		}
		if info, err := os.Stat(n.path); err == nil && os.SameFile(info, std) {
			return refuse(n.flag, n.path, "same file as standard output; each output needs its own path", nil)
		}
	}
	return nil
}

// maxOutputLinkHops bounds the symbolic links followed from one output path.
// The kernel stops at 40 as well; a longer chain or a loop is refused at load
// instead of being left to fail at the write.
const maxOutputLinkHops = 40

// unreachableOutputPrefix marks a path the kernel cannot name: a ".."
// follows a component that does not exist, so the walk cannot reach the
// parent a lexical collapse would invent. An identity resolved with a
// readable working directory is absolute, so it does not start with this
// prefix, and the marked spelling is not that file. The write of the raw
// path fails later.
const unreachableOutputPrefix = "unreachable:"

// outputIdentity is the path the file would have if the filesystem resolved
// the raw path, whether or not the file exists.
//
// THE RULE: walk the path as the kernel does, one component at a time, from
// the root (or from the working directory for a relative path). A symbolic
// link is resolved when it is met, before the next component is looked at, so
// ".." applies to the directory the link led to and not to the directory that
// holds the link: with link -> elsewhere/dir, "link/../report" is
// "elsewhere/report". The path is never cleaned or made absolute first,
// because filepath.Clean and filepath.Abs remove "link/.." textually, before
// link is read, and name a different file. A relative link target is taken
// against the directory that holds the link, and at most maxOutputLinkHops
// links are followed. At the first component that does not exist the rest is
// appended as written, until a "..". The kernel cannot traverse the missing
// component, so that ".." does not name the parent a lexical collapse would
// reach. The path is unwritable. Its identity is the raw spelling under
// unreachableOutputPrefix, and the write fails.
//
// WHY a dangling final link is followed: os.WriteFile on alias.json ->
// actual.json creates actual.json, so a not-yet-existing target is the same
// file as the link's name.
//
// WHY the write-time os.SameFile check in run.go stays: this runs once, at
// load, and the disk can change before the write; that check is the second
// line and is never the one meant to fire for an alias this function can see.
func outputIdentity(path string) (string, error) {
	root, rest := splitRoot(path)
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			// The working directory is unreadable; the cleaned path is
			// still a usable key for the comparison.
			return filepath.Clean(path), nil
		}
		root, _ = splitRoot(cwd)
		rest = cwd[len(root):] + string(filepath.Separator) + path
	}
	pending := splitComponents(rest)
	resolved := root
	missing := false // true once a component did not exist. A later ".." is unreachable.
	for hops := 0; len(pending) > 0; {
		name := pending[0]
		pending = pending[1:]
		switch name {
		case ".":
			continue
		case "..":
			if missing {
				return unreachableOutputPrefix + path, nil
			}
			if resolved != root {
				resolved = filepath.Dir(resolved)
			}
			continue
		}
		next := filepath.Join(resolved, name)
		if missing {
			resolved = next
			continue
		}
		info, err := os.Lstat(next)
		if err != nil {
			missing = true
			resolved = next
			continue
		}
		if info.Mode()&os.ModeSymlink == 0 {
			resolved = next
			continue
		}
		if hops == maxOutputLinkHops {
			return "", fmt.Errorf("more than %d symbolic links, or a link loop", maxOutputLinkHops)
		}
		hops++
		target, err := os.Readlink(next)
		if err != nil {
			resolved = next
			missing = true
			continue
		}
		if tRoot, tRest := splitRoot(target); tRoot != "" {
			resolved = tRoot
			target = tRest
		}
		pending = append(splitComponents(target), pending...)
	}
	return resolved, nil
}

// splitRoot splits a path into its root ("/" or `C:\`, "" when relative) and
// the remainder, without cleaning either.
func splitRoot(path string) (root, rest string) {
	vol := filepath.VolumeName(path)
	rest = path[len(vol):]
	if rest != "" && os.IsPathSeparator(rest[0]) {
		return vol + string(filepath.Separator), rest
	}
	if vol != "" {
		return vol, rest
	}
	return "", path
}

// splitComponents splits a path on its separators and drops empty parts, so a
// trailing or doubled separator names the same component.
func splitComponents(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool { return os.IsPathSeparator(byte(r)) && r < 0x80 })
}

// refuseExclusive refuses pairs that cannot both hold: two output formats,
// --error with --no-error, --quiet with --verbose.
func refuseExclusive(h *holders) error {
	formats := 0
	if h.text.set {
		formats++
	}
	if h.asJSON.set {
		formats++
	}
	if h.sarif.set {
		formats++
	}
	if formats > 1 {
		return &cliError{code: exitFatal, msg: "choose one of --text, --json, --sarif"}
	}
	if h.errorOn.set && h.noError.set {
		return &cliError{code: exitFatal, msg: "choose one of --error, --no-error"}
	}
	if h.quiet.set && h.verbose.set {
		return &cliError{code: exitFatal, msg: "choose one of --quiet, --verbose"}
	}
	return nil
}

// resolveFormat picks the standard-output format: a format flag, else
// PROVEN_OUTPUT_FORMAT, else text.
func resolveFormat(getenv func(string) string, h *holders) (string, error) {
	if h.text.set && h.text.v {
		return "text", nil
	}
	if h.asJSON.set && h.asJSON.v {
		return "json", nil
	}
	if h.sarif.set && h.sarif.v {
		return "sarif", nil
	}
	if v := getenv("PROVEN_OUTPUT_FORMAT"); v != "" {
		return parseFormat(v)
	}
	return "text", nil
}

func parseFormat(v string) (string, error) {
	switch strings.ToLower(v) {
	case "text", "json", "sarif":
		return strings.ToLower(v), nil
	default:
		return "", refuse("PROVEN_OUTPUT_FORMAT", v, "output format is not one of the allowed values", []string{"text", "json", "sarif"})
	}
}

// resolveError decides whether findings fail the run. ci defaults to yes,
// because a pull-request check that cannot fail is not a check; scan
// defaults to no, as Semgrep's does.
func resolveError(cmd string, getenv func(string) string, h *holders) (bool, error) {
	if h.errorOn.set {
		return h.errorOn.v, nil
	}
	if h.noError.set {
		return !h.noError.v, nil
	}
	if v := getenv("PROVEN_ERROR"); v != "" {
		return parseBoolSetting("PROVEN_ERROR", v)
	}
	return cmd == "ci", nil
}

// resolveBaseline reads the baseline commit. Semgrep's two variables are
// read after Proven's own, so a pipeline that already sets them keeps
// working unchanged (ADR-0016).
func resolveBaseline(getenv func(string) string, h *holders) string {
	switch {
	case h.baseline.set:
		return h.baseline.v
	case getenv("PROVEN_BASELINE_COMMIT") != "":
		return getenv("PROVEN_BASELINE_COMMIT")
	case getenv("SEMGREP_BASELINE_COMMIT") != "":
		return getenv("SEMGREP_BASELINE_COMMIT")
	case getenv("SEMGREP_BASELINE_REF") != "":
		return getenv("SEMGREP_BASELINE_REF")
	default:
		return ""
	}
}

// resolveConfig finds Proven's configuration file: --config, else
// PROVEN_CONFIG, else proven.yaml in the target, else none (the defaults).
func resolveConfig(target string, getenv func(string) string, exists func(string) bool, h *holders) (string, error) {
	var path string
	switch {
	case h.configPath.set:
		path = h.configPath.v
	case getenv("PROVEN_CONFIG") != "":
		path = getenv("PROVEN_CONFIG")
	default:
		candidate := filepath.Join(target, "proven.yaml")
		if exists(candidate) {
			path = candidate
		}
	}
	if path == "" {
		return "", nil
	}
	// A Semgrep registry name (auto, p/…, r/…) is refused by name rather than
	// read as a missing file: Proven has no rule registry to fetch.
	if isRegistry(path) {
		return "", refuse("--config", path, "--config takes a path to a Proven configuration file", []string{"a file path", "proven.yaml if present", "omit to use the defaults"})
	}
	if !exists(path) {
		return "", refuse("--config", path, "configuration file not found", []string{"a file path", "proven.yaml if present", "omit to use the defaults"})
	}
	return path, nil
}

// isRegistry reports whether v is a Semgrep registry reference.
func isRegistry(v string) bool {
	if strings.EqualFold(v, "auto") {
		return true
	}
	return strings.HasPrefix(v, "p/") || strings.HasPrefix(v, "r/")
}

// resolveSeverity reads the lowest severity reported. Unlike Semgrep's
// --severity, which selects exact levels, Proven's is a floor: WARNING
// keeps WARNING and ERROR.
func resolveSeverity(getenv func(string) string, h *holders) (review.Severity, error) {
	if h.severity.set {
		return parseSeverity("--severity", h.severity.v)
	}
	if v := getenv("PROVEN_SEVERITY"); v != "" {
		return parseSeverity("PROVEN_SEVERITY", v)
	}
	return review.SeverityInfo, nil
}

// parseSeverity accepts the three levels in any case.
func parseSeverity(field, v string) (review.Severity, error) {
	switch strings.ToUpper(v) {
	case "INFO", "WARNING", "ERROR":
		return review.Severity(strings.ToUpper(v)), nil
	default:
		return "", refuse(field, v, "severity is not one of the allowed levels", []string{"INFO", "WARNING", "ERROR"})
	}
}

// resolveTimeout reads the review's time limit in seconds; zero is no limit.
func resolveTimeout(getenv func(string) string, h *holders) (int, error) {
	if h.timeout.set {
		return parseTimeout("--timeout", h.timeout.v)
	}
	if v := getenv("PROVEN_TIMEOUT"); v != "" {
		return parseTimeout("PROVEN_TIMEOUT", v)
	}
	return 0, nil
}

// parseTimeout refuses a value that is not a whole number of seconds, and
// one too large to become a time.Duration, rather than letting it wrap.
func parseTimeout(field, v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, refuse(field, v, "timeout is a whole number of seconds, zero or more", []string{"0", "a positive number of seconds"})
	}
	if int64(n) > maxTimeoutSec {
		return 0, refuseTimeoutBound(field, v)
	}
	return n, nil
}

// refuseTimeoutBound refuses a timeout above maxTimeoutSec and names the range.
func refuseTimeoutBound(field, v string) error {
	limit := strconv.FormatInt(maxTimeoutSec, 10)
	return refuse(field, v, "timeout is a whole number of seconds from 0 through "+limit, []string{"0", limit})
}

// resolveSample decides sample mode: the sample reviewer, with no network
// and no keys (ADR-0001). SAMPLE_MODE set to the empty string counts as unset.
func resolveSample(getenv func(string) string, h *holders) (bool, error) {
	if h.sample.set {
		return h.sample.v, nil
	}
	if v := getenv("SAMPLE_MODE"); v != "" {
		return parseBoolSetting("SAMPLE_MODE", v)
	}
	return false, nil
}

// parseBoolSetting reads a boolean variable and refuses anything else with
// true, false as the allowed values: the same answer the flag gives.
func parseBoolSetting(field, v string) (bool, error) {
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return false, refuse(field, v, "value must be true or false", []string{"true", "false"})
	}
	return parsed, nil
}

// splitList reads a comma-separated variable, the variable form of a
// repeatable flag, dropping empty items.
func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// refuse builds an exit-7 refusal in the reviewer's own ConfigError form,
// so a setting refused by the shell and one refused by the reviewer read alike.
func refuse(field, value, reason string, allowed []string) error {
	ce := &review.ConfigError{Field: field, Value: value, Reason: reason, Allowed: allowed}
	return &cliError{code: exitConfig, msg: ce.Error()}
}
