package cli

import (
	"flag"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/review"
)

// maxTimeoutSec is the largest whole number of seconds that fits in a time.Duration.
// math.MaxInt64 / time.Second is 9223372036; the next second overflows the conversion.
const maxTimeoutSec = int64(math.MaxInt64 / time.Second)

const (
	exitOK       = 0
	exitFindings = 1
	exitFatal    = 2
	exitConfig   = 7
)

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

type cliError struct {
	code int
	msg  string
}

func (e *cliError) Error() string { return e.msg }

type boolPresence struct {
	set bool
	v   bool
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
		return err
	}
	b.set = true
	b.v = v
	return nil
}

func (b *boolPresence) IsBoolFlag() bool { return true }

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

type holders struct {
	text, asJSON, sarif                 boolPresence
	errorOn, noError                    boolPresence
	sample, quiet, verbose, help, ver   boolPresence
	textOutput, jsonOutput, sarifOutput setString
	baseline, configPath, severity      setString
	timeout, output                     setString
	include, exclude                    multiFlag
}

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

func defineFlags() (*flag.FlagSet, *holders) {
	h := &holders{}
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
		return config{}, &cliError{code: exitFatal, msg: err.Error()}
	}
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
	if isRegistry(path) {
		return "", refuse("--config", path, "--config takes a path to a Proven configuration file", []string{"a file path", "proven.yaml if present", "omit to use the defaults"})
	}
	if !exists(path) {
		return "", refuse("--config", path, "configuration file not found", []string{"a file path", "proven.yaml if present", "omit to use the defaults"})
	}
	return path, nil
}

func isRegistry(v string) bool {
	if strings.EqualFold(v, "auto") {
		return true
	}
	return strings.HasPrefix(v, "p/") || strings.HasPrefix(v, "r/")
}

func resolveSeverity(getenv func(string) string, h *holders) (review.Severity, error) {
	if h.severity.set {
		return parseSeverity("--severity", h.severity.v)
	}
	if v := getenv("PROVEN_SEVERITY"); v != "" {
		return parseSeverity("PROVEN_SEVERITY", v)
	}
	return review.SeverityInfo, nil
}

func parseSeverity(field, v string) (review.Severity, error) {
	switch strings.ToUpper(v) {
	case "INFO", "WARNING", "ERROR":
		return review.Severity(strings.ToUpper(v)), nil
	default:
		return "", refuse(field, v, "severity is not one of the allowed levels", []string{"INFO", "WARNING", "ERROR"})
	}
}

func resolveTimeout(getenv func(string) string, h *holders) (int, error) {
	if h.timeout.set {
		return parseTimeout("--timeout", h.timeout.v)
	}
	if v := getenv("PROVEN_TIMEOUT"); v != "" {
		return parseTimeout("PROVEN_TIMEOUT", v)
	}
	return 0, nil
}

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

func refuseTimeoutBound(field, v string) error {
	limit := strconv.FormatInt(maxTimeoutSec, 10)
	return refuse(field, v, "timeout is a whole number of seconds from 0 through "+limit, []string{"0", limit})
}

func resolveSample(getenv func(string) string, h *holders) (bool, error) {
	if h.sample.set {
		return h.sample.v, nil
	}
	if v := getenv("SAMPLE_MODE"); v != "" {
		return parseBoolSetting("SAMPLE_MODE", v)
	}
	return false, nil
}

func parseBoolSetting(field, v string) (bool, error) {
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return false, refuse(field, v, "value must be true or false", []string{"true", "false"})
	}
	return parsed, nil
}

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

func refuse(field, value, reason string, allowed []string) error {
	ce := &review.ConfigError{Field: field, Value: value, Reason: reason, Allowed: allowed}
	return &cliError{code: exitConfig, msg: ce.Error()}
}
