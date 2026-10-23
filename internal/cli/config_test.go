package cli

import (
	"errors"
	"flag"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

func emptyEnv(string) string { return "" }

func noFile(string) bool { return false }

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func filesOf(paths []string) func(string) bool {
	return func(p string) bool {
		for _, one := range paths {
			if one == p {
				return true
			}
		}
		return false
	}
}

func mustResolve(t *testing.T, args []string, getenv func(string) string, exists func(string) bool) config {
	t.Helper()
	if getenv == nil {
		getenv = emptyEnv
	}
	if exists == nil {
		exists = noFile
	}
	cfg, err := resolve(args, getenv, exists)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	return cfg
}

func mustRefuse(t *testing.T, args []string, getenv func(string) string, exists func(string) bool, code int) string {
	t.Helper()
	if getenv == nil {
		getenv = emptyEnv
	}
	if exists == nil {
		exists = noFile
	}
	_, err := resolve(args, getenv, exists)
	var ce *cliError
	if !errors.As(err, &ce) {
		t.Fatalf("got %v, want cliError code %d", err, code)
	}
	if ce.code != code {
		t.Fatalf("code %d, want %d: %s", ce.code, code, ce.msg)
	}
	return ce.msg
}

func TestPrecedence(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		env   map[string]string
		files []string
		check func(*testing.T, config)
	}{
		{
			name: "format defaults to text",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.format != "text" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "PROVEN_OUTPUT_FORMAT overrides the default",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_OUTPUT_FORMAT": "json"},
			check: func(t *testing.T, c config) {
				if c.format != "json" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "--text overrides PROVEN_OUTPUT_FORMAT",
			args: []string{"scan", "--text"},
			env:  map[string]string{"PROVEN_OUTPUT_FORMAT": "sarif"},
			check: func(t *testing.T, c config) {
				if c.format != "text" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "--json overrides PROVEN_OUTPUT_FORMAT",
			args: []string{"scan", "--json"},
			env:  map[string]string{"PROVEN_OUTPUT_FORMAT": "text"},
			check: func(t *testing.T, c config) {
				if c.format != "json" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "--sarif overrides the default",
			args: []string{"scan", "--sarif"},
			check: func(t *testing.T, c config) {
				if c.format != "sarif" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "PROVEN_OUTPUT_FORMAT sarif",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_OUTPUT_FORMAT": "sarif"},
			check: func(t *testing.T, c config) {
				if c.format != "sarif" {
					t.Fatalf("format %q", c.format)
				}
			},
		},
		{
			name: "output files default to empty",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.textOutput != "" || c.jsonOutput != "" || c.sarifOutput != "" || c.output != "" {
					t.Fatalf("outputs %+v", c)
				}
			},
		},
		{
			name: "--text-output sets the text file",
			args: []string{"scan", "--text-output=out.txt"},
			check: func(t *testing.T, c config) {
				if c.textOutput != "out.txt" {
					t.Fatalf("text output %q", c.textOutput)
				}
			},
		},
		{
			name: "--json-output sets the json file",
			args: []string{"scan", "--json-output=out.json"},
			check: func(t *testing.T, c config) {
				if c.jsonOutput != "out.json" {
					t.Fatalf("json output %q", c.jsonOutput)
				}
			},
		},
		{
			name: "--sarif-output sets the sarif file",
			args: []string{"scan", "--sarif-output=out.sarif"},
			check: func(t *testing.T, c config) {
				if c.sarifOutput != "out.sarif" {
					t.Fatalf("sarif output %q", c.sarifOutput)
				}
			},
		},
		{
			name: "PROVEN_OUTPUT sets the output file",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_OUTPUT": "env.txt"},
			check: func(t *testing.T, c config) {
				if c.output != "env.txt" {
					t.Fatalf("output %q", c.output)
				}
			},
		},
		{
			name: "--output overrides PROVEN_OUTPUT",
			args: []string{"scan", "--output=flag.txt"},
			env:  map[string]string{"PROVEN_OUTPUT": "env.txt"},
			check: func(t *testing.T, c config) {
				if c.output != "flag.txt" {
					t.Fatalf("output %q", c.output)
				}
			},
		},
		{
			name: "-o overrides PROVEN_OUTPUT",
			args: []string{"scan", "-o", "short.txt"},
			env:  map[string]string{"PROVEN_OUTPUT": "env.txt"},
			check: func(t *testing.T, c config) {
				if c.output != "short.txt" {
					t.Fatalf("output %q", c.output)
				}
			},
		},
		{
			name: "scan does not exit on findings by default",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.errorOnFindings {
					t.Fatal("error on findings")
				}
			},
		},
		{
			name: "ci exits on findings by default",
			args: []string{"ci"},
			check: func(t *testing.T, c config) {
				if !c.errorOnFindings {
					t.Fatal("error on findings is off")
				}
			},
		},
		{
			name: "PROVEN_ERROR overrides the scan default",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_ERROR": "true"},
			check: func(t *testing.T, c config) {
				if !c.errorOnFindings {
					t.Fatal("error on findings is off")
				}
			},
		},
		{
			name: "PROVEN_ERROR overrides the ci default",
			args: []string{"ci"},
			env:  map[string]string{"PROVEN_ERROR": "false"},
			check: func(t *testing.T, c config) {
				if c.errorOnFindings {
					t.Fatal("error on findings")
				}
			},
		},
		{
			name: "--error overrides PROVEN_ERROR",
			args: []string{"scan", "--error"},
			env:  map[string]string{"PROVEN_ERROR": "false"},
			check: func(t *testing.T, c config) {
				if !c.errorOnFindings {
					t.Fatal("error on findings is off")
				}
			},
		},
		{
			name: "--no-error overrides PROVEN_ERROR",
			args: []string{"ci", "--no-error"},
			env:  map[string]string{"PROVEN_ERROR": "true"},
			check: func(t *testing.T, c config) {
				if c.errorOnFindings {
					t.Fatal("error on findings")
				}
			},
		},
		{
			name: "baseline defaults to empty on scan",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.baseline != "" {
					t.Fatalf("baseline %q", c.baseline)
				}
			},
		},
		{
			name: "SEMGREP_BASELINE_REF is the last baseline variable",
			args: []string{"scan"},
			env:  map[string]string{"SEMGREP_BASELINE_REF": "from-ref"},
			check: func(t *testing.T, c config) {
				if c.baseline != "from-ref" {
					t.Fatalf("baseline %q", c.baseline)
				}
			},
		},
		{
			name: "SEMGREP_BASELINE_COMMIT overrides SEMGREP_BASELINE_REF",
			args: []string{"scan"},
			env: map[string]string{
				"SEMGREP_BASELINE_COMMIT": "from-commit",
				"SEMGREP_BASELINE_REF":    "from-ref",
			},
			check: func(t *testing.T, c config) {
				if c.baseline != "from-commit" {
					t.Fatalf("baseline %q", c.baseline)
				}
			},
		},
		{
			name: "PROVEN_BASELINE_COMMIT overrides the Semgrep variables",
			args: []string{"scan"},
			env: map[string]string{
				"PROVEN_BASELINE_COMMIT":  "from-proven",
				"SEMGREP_BASELINE_COMMIT": "from-commit",
				"SEMGREP_BASELINE_REF":    "from-ref",
			},
			check: func(t *testing.T, c config) {
				if c.baseline != "from-proven" {
					t.Fatalf("baseline %q", c.baseline)
				}
			},
		},
		{
			name: "--baseline-commit overrides PROVEN_BASELINE_COMMIT",
			args: []string{"scan", "--baseline-commit=from-flag"},
			env:  map[string]string{"PROVEN_BASELINE_COMMIT": "from-proven"},
			check: func(t *testing.T, c config) {
				if c.baseline != "from-flag" {
					t.Fatalf("baseline %q", c.baseline)
				}
			},
		},
		{
			name: "config defaults to empty when proven.yaml is absent",
			args: []string{"scan", "/repo"},
			check: func(t *testing.T, c config) {
				if c.configPath != "" {
					t.Fatalf("config %q", c.configPath)
				}
			},
		},
		{
			name:  "proven.yaml in the target is the default config",
			args:  []string{"scan", "/repo"},
			files: []string{"/repo/proven.yaml"},
			check: func(t *testing.T, c config) {
				if c.configPath != "/repo/proven.yaml" {
					t.Fatalf("config %q", c.configPath)
				}
			},
		},
		{
			name:  "PROVEN_CONFIG overrides proven.yaml",
			args:  []string{"scan", "/repo"},
			env:   map[string]string{"PROVEN_CONFIG": "/env.yaml"},
			files: []string{"/repo/proven.yaml", "/env.yaml"},
			check: func(t *testing.T, c config) {
				if c.configPath != "/env.yaml" {
					t.Fatalf("config %q", c.configPath)
				}
			},
		},
		{
			name:  "--config overrides PROVEN_CONFIG",
			args:  []string{"scan", "--config=/flag.yaml", "/repo"},
			env:   map[string]string{"PROVEN_CONFIG": "/env.yaml"},
			files: []string{"/repo/proven.yaml", "/env.yaml", "/flag.yaml"},
			check: func(t *testing.T, c config) {
				if c.configPath != "/flag.yaml" {
					t.Fatalf("config %q", c.configPath)
				}
			},
		},
		{
			name: "include and exclude default to empty",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if len(c.include) != 0 || len(c.exclude) != 0 {
					t.Fatalf("include %v exclude %v", c.include, c.exclude)
				}
			},
		},
		{
			name: "PROVEN_INCLUDE is a comma-separated list",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_INCLUDE": "a, b"},
			check: func(t *testing.T, c config) {
				if strings.Join(c.include, "|") != "a|b" {
					t.Fatalf("include %v", c.include)
				}
			},
		},
		{
			name: "--include overrides PROVEN_INCLUDE",
			args: []string{"scan", "--include=flag"},
			env:  map[string]string{"PROVEN_INCLUDE": "a,b"},
			check: func(t *testing.T, c config) {
				if strings.Join(c.include, "|") != "flag" {
					t.Fatalf("include %v", c.include)
				}
			},
		},
		{
			name: "repeated --include keeps every glob",
			args: []string{"scan", "--include=a", "--include=b"},
			check: func(t *testing.T, c config) {
				if strings.Join(c.include, "|") != "a|b" {
					t.Fatalf("include %v", c.include)
				}
			},
		},
		{
			name: "PROVEN_EXCLUDE is a comma-separated list",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_EXCLUDE": "x,y"},
			check: func(t *testing.T, c config) {
				if strings.Join(c.exclude, "|") != "x|y" {
					t.Fatalf("exclude %v", c.exclude)
				}
			},
		},
		{
			name: "--exclude overrides PROVEN_EXCLUDE",
			args: []string{"scan", "--exclude=flag"},
			env:  map[string]string{"PROVEN_EXCLUDE": "x,y"},
			check: func(t *testing.T, c config) {
				if strings.Join(c.exclude, "|") != "flag" {
					t.Fatalf("exclude %v", c.exclude)
				}
			},
		},
		{
			name: "severity defaults to INFO",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.severity != review.SeverityInfo {
					t.Fatalf("severity %q", c.severity)
				}
			},
		},
		{
			name: "PROVEN_SEVERITY overrides the default",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_SEVERITY": "warning"},
			check: func(t *testing.T, c config) {
				if c.severity != review.SeverityWarning {
					t.Fatalf("severity %q", c.severity)
				}
			},
		},
		{
			name: "--severity overrides PROVEN_SEVERITY",
			args: []string{"scan", "--severity=ERROR"},
			env:  map[string]string{"PROVEN_SEVERITY": "INFO"},
			check: func(t *testing.T, c config) {
				if c.severity != review.SeverityError {
					t.Fatalf("severity %q", c.severity)
				}
			},
		},
		{
			name: "timeout defaults to no deadline",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.timeoutSec != 0 {
					t.Fatalf("timeout %d", c.timeoutSec)
				}
			},
		},
		{
			name: "PROVEN_TIMEOUT overrides the default",
			args: []string{"scan"},
			env:  map[string]string{"PROVEN_TIMEOUT": "30"},
			check: func(t *testing.T, c config) {
				if c.timeoutSec != 30 {
					t.Fatalf("timeout %d", c.timeoutSec)
				}
			},
		},
		{
			name: "--timeout overrides PROVEN_TIMEOUT",
			args: []string{"scan", "--timeout=5"},
			env:  map[string]string{"PROVEN_TIMEOUT": "30"},
			check: func(t *testing.T, c config) {
				if c.timeoutSec != 5 {
					t.Fatalf("timeout %d", c.timeoutSec)
				}
			},
		},
		{
			name: "sample defaults to off",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.sample {
					t.Fatal("sample is on")
				}
			},
		},
		{
			name: "SAMPLE_MODE=true turns sample on",
			args: []string{"scan"},
			env:  map[string]string{"SAMPLE_MODE": "true"},
			check: func(t *testing.T, c config) {
				if !c.sample {
					t.Fatal("sample is off")
				}
			},
		},
		{
			name: "--sample overrides SAMPLE_MODE=false",
			args: []string{"scan", "--sample"},
			env:  map[string]string{"SAMPLE_MODE": "false"},
			check: func(t *testing.T, c config) {
				if !c.sample {
					t.Fatal("sample is off")
				}
			},
		},
		{
			name: "--sample=false overrides SAMPLE_MODE=true",
			args: []string{"scan", "--sample=false"},
			env:  map[string]string{"SAMPLE_MODE": "true"},
			check: func(t *testing.T, c config) {
				if c.sample {
					t.Fatal("sample is on")
				}
			},
		},
		{
			name: "quiet and verbose default to off",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.quiet || c.verbose {
					t.Fatalf("quiet %v verbose %v", c.quiet, c.verbose)
				}
			},
		},
		{
			name: "--quiet",
			args: []string{"scan", "--quiet"},
			check: func(t *testing.T, c config) {
				if !c.quiet || c.verbose {
					t.Fatalf("quiet %v verbose %v", c.quiet, c.verbose)
				}
			},
		},
		{
			name: "--verbose",
			args: []string{"scan", "--verbose"},
			check: func(t *testing.T, c config) {
				if c.quiet || !c.verbose {
					t.Fatalf("quiet %v verbose %v", c.quiet, c.verbose)
				}
			},
		},
		{
			name: "scan target defaults to dot",
			args: []string{"scan"},
			check: func(t *testing.T, c config) {
				if c.command != "scan" || c.target != "." {
					t.Fatalf("command %q target %q", c.command, c.target)
				}
			},
		},
		{
			name: "scan takes one target",
			args: []string{"scan", "--sample", "src"},
			check: func(t *testing.T, c config) {
				if c.target != "src" || !c.sample {
					t.Fatalf("target %q sample %v", c.target, c.sample)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustResolve(t, tt.args, envOf(tt.env), filesOf(tt.files))
			tt.check(t, cfg)
		})
	}
}

func TestConfigAutoIsRefused(t *testing.T) {
	msg := mustRefuse(t, []string{"scan", "--config=auto"}, nil, nil, exitConfig)
	if !strings.Contains(msg, "auto") || !strings.Contains(msg, "configuration file") {
		t.Fatalf("message %q", msg)
	}
}

func TestConfigRegistryValueIsRefused(t *testing.T) {
	for _, value := range []string{"p/python", "r/java", "AUTO"} {
		msg := mustRefuse(t, []string{"scan", "--config=" + value}, nil, nil, exitConfig)
		if !strings.Contains(msg, "configuration file") {
			t.Fatalf("%s: %q", value, msg)
		}
	}
}

func TestProvenConfigAutoIsRefused(t *testing.T) {
	msg := mustRefuse(t, []string{"scan"}, envOf(map[string]string{"PROVEN_CONFIG": "auto"}), nil, exitConfig)
	if !strings.Contains(msg, "auto") {
		t.Fatalf("message %q", msg)
	}
}

func TestMissingConfigFileIsRefused(t *testing.T) {
	msg := mustRefuse(t, []string{"scan", "--config=/missing.yaml"}, nil, nil, exitConfig)
	if !strings.Contains(msg, "/missing.yaml") {
		t.Fatalf("message %q", msg)
	}
}

func TestUnknownSeverityIsRefused(t *testing.T) {
	msg := mustRefuse(t, []string{"scan"}, envOf(map[string]string{"PROVEN_SEVERITY": "CRITICAL"}), nil, exitConfig)
	if !strings.Contains(msg, "INFO, WARNING, ERROR") {
		t.Fatalf("message %q", msg)
	}
}

func TestUnknownFlagListsKnownFlags(t *testing.T) {
	msg := mustRefuse(t, []string{"scan", "--not-a-flag"}, nil, nil, exitFatal)
	if !strings.Contains(msg, "known flags:") {
		t.Fatalf("message %q", msg)
	}
	for _, name := range knownFlagNames() {
		if !strings.Contains(msg, name) {
			t.Fatalf("missing %s in %q", name, msg)
		}
	}
}

func TestSecretIsNotAFlag(t *testing.T) {
	msg := mustRefuse(t, []string{"scan", "--meterless-runtime-token=sek"}, nil, nil, exitFatal)
	if strings.Contains(msg, "sek") {
		t.Fatalf("value echoed: %q", msg)
	}
	if !strings.Contains(msg, "known flags:") {
		t.Fatalf("message %q", msg)
	}
}

// TestRefusedBoolFlagEmptyValue proves Set("") is a refusal. The value
// string is empty, which is also how a flag that was never set looks, so
// the value alone cannot be the signal.
func TestRefusedBoolFlagEmptyValue(t *testing.T) {
	h := &holders{}
	h.sample.name = "--sample"
	if err := h.sample.Set(""); err == nil {
		t.Fatal("Set accepted an empty value")
	}
	name, value, ok := refusedBoolFlag(h)
	if !ok || name != "--sample" || value != "" {
		t.Fatalf("refusedBoolFlag() = %q, %q, %v", name, value, ok)
	}
	if err := h.sample.Set("true"); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := refusedBoolFlag(h); ok {
		t.Fatal("a parsed value is still a refusal")
	}
}

func TestBadValuesAreRefused(t *testing.T) {
	// flag, when set, is a boolean flag. The mutated field is that flag's
	// value. A bad value exits 7 and names true, false, the same as the variable.
	cases := []struct {
		name string
		args []string
		env  map[string]string
		flag string
	}{
		{name: "format", args: []string{"scan"}, env: map[string]string{"PROVEN_OUTPUT_FORMAT": "xml"}},
		{name: "error", args: []string{"scan"}, env: map[string]string{"PROVEN_ERROR": "maybe"}},
		{name: "timeout", args: []string{"scan"}, env: map[string]string{"PROVEN_TIMEOUT": "-1"}},
		{name: "timeout word", args: []string{"scan", "--timeout=soon"}},
		{name: "sample", args: []string{"scan"}, env: map[string]string{"SAMPLE_MODE": "yes"}},
		{name: "severity flag", args: []string{"scan", "--severity=CRITICAL"}},
		{name: "--text", args: []string{"scan", "--text=maybe"}, flag: "--text"},
		{name: "--json", args: []string{"scan", "--json=maybe"}, flag: "--json"},
		{name: "--sarif", args: []string{"scan", "--sarif=maybe"}, flag: "--sarif"},
		{name: "--error", args: []string{"scan", "--error=maybe"}, flag: "--error"},
		{name: "--no-error", args: []string{"scan", "--no-error=maybe"}, flag: "--no-error"},
		{name: "--sample", args: []string{"scan", "--sample=maybe"}, flag: "--sample"},
		{name: "--quiet", args: []string{"scan", "--quiet=maybe"}, flag: "--quiet"},
		{name: "--verbose", args: []string{"scan", "--verbose=maybe"}, flag: "--verbose"},
		{name: "--help", args: []string{"scan", "--help=maybe"}, flag: "--help"},
		{name: "--version", args: []string{"scan", "--version=maybe"}, flag: "--version"},
		// An empty value is still a value. Set("") used to leave the
		// refusal indistinguishable from a flag that was never set.
		{name: "--text=", args: []string{"scan", "--text="}, flag: "--text"},
		{name: "--json=", args: []string{"scan", "--json="}, flag: "--json"},
		{name: "--sarif=", args: []string{"scan", "--sarif="}, flag: "--sarif"},
		{name: "--error=", args: []string{"scan", "--error="}, flag: "--error"},
		{name: "--no-error=", args: []string{"scan", "--no-error="}, flag: "--no-error"},
		{name: "--sample=", args: []string{"scan", "--sample="}, flag: "--sample"},
		{name: "--quiet=", args: []string{"scan", "--quiet="}, flag: "--quiet"},
		{name: "--verbose=", args: []string{"scan", "--verbose="}, flag: "--verbose"},
		{name: "--help=", args: []string{"scan", "--help="}, flag: "--help"},
		{name: "--version=", args: []string{"scan", "--version="}, flag: "--version"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			msg := mustRefuse(t, tt.args, envOf(tt.env), nil, exitConfig)
			if tt.flag == "" {
				return
			}
			if !strings.Contains(msg, tt.flag) || !strings.Contains(msg, "true, false") {
				t.Fatalf("message %q", msg)
			}
		})
	}
}

func TestUsageErrors(t *testing.T) {
	cases := [][]string{
		{"scan", "--error", "--no-error"},
		{"scan", "--quiet", "--verbose"},
		{"scan", "--text", "--json"},
		{"other"},
		{"scan", "a", "b"},
		{"ci", "target"},
		{},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			mustRefuse(t, args, nil, nil, exitFatal)
		})
	}
}

func TestHelpAndVersionSkipValidation(t *testing.T) {
	cfg := mustResolve(t, []string{"--help", "--severity=NOPE"}, nil, nil)
	if !cfg.help {
		t.Fatal("help is off")
	}
	cfg = mustResolve(t, []string{"--version"}, nil, nil)
	if !cfg.version {
		t.Fatal("version is off")
	}
}

func TestKnownFlagsMatchTheFlagSet(t *testing.T) {
	fs, _ := defineFlags()
	var got []string
	fs.VisitAll(func(f *flag.Flag) {
		if len(f.Name) == 1 {
			got = append(got, "-"+f.Name)
			return
		}
		got = append(got, "--"+f.Name)
	})
	sort.Strings(got)
	want := append([]string{}, knownFlagNames()...)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("flag set %v\nknown %v", got, want)
	}
}
