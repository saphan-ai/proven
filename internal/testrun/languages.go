package testrun

import (
	_ "embed"
	"encoding/json"
	"path"
	"regexp"
	"strings"
	"unicode"
)

// LanguageAdapter supplies only fixed commands and test-name rules. Container
// remains the Adapter/FixRunner: every language shares its isolation, deadlines,
// staging, differential decisions, fix reruns and cleanup reporting.
type LanguageAdapter interface {
	Name() string
	Matches(testPath string) bool
	ValidName(name string) bool
	Command(testPath, name string, packageRun bool) []string
}

// LanguageAdapters is a deterministic registry, like manifest.Readers. These
// adapters do not fetch runtimes or evaluate model-provided commands.
func LanguageAdapters() []LanguageAdapter {
	return []LanguageAdapter{goLanguage{}, scriptLanguage{"pytest", "python3", pytestScript}, scriptLanguage{"vitest", "node", vitestScript}, scriptLanguage{"node", "node", nodeScript}}
}

//go:embed runners/pytest.py
var pytestScript string

//go:embed runners/node.mjs
var nodeScript string

//go:embed runners/vitest.mjs
var vitestScript string

type goLanguage struct{}

func (goLanguage) Name() string            { return "go" }
func (goLanguage) Matches(p string) bool   { return isGoTestFile(p) }
func (goLanguage) ValidName(n string) bool { return namePattern.MatchString(n) }
func (goLanguage) Command(p, n string, pkg bool) []string {
	dir := path.Dir(p)
	if dir != "." {
		dir = "./" + dir
	}
	if pkg {
		return []string{"go", "test", jsonFlag, dir}
	}
	return []string{"go", "test", jsonFlag, "-run", "^" + n + "$", dir}
}

type scriptLanguage struct{ id, executable, script string }

func (a scriptLanguage) Name() string { return a.id }
func (a scriptLanguage) Matches(p string) bool {
	base := path.Base(p)
	switch a.id {
	case "pytest":
		return strings.HasSuffix(base, ".py") && (strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py"))
	case "vitest":
		for _, ext := range []string{".ts", ".tsx", ".mts", ".cts"} {
			if strings.HasSuffix(base, ".test"+ext) || strings.HasSuffix(base, ".spec"+ext) || strings.Contains(p, "__tests__/") && strings.HasSuffix(base, ext) {
				return true
			}
		}
		return strings.HasSuffix(base, ".spec.js") || strings.HasSuffix(base, ".spec.mjs") || strings.HasSuffix(base, ".spec.cjs")
	case "node":
		return strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, ".test.mjs") || strings.HasSuffix(base, ".test.cjs")
	}
	return false
}

var pythonName = regexp.MustCompile(`^(?:[A-Za-z_][A-Za-z0-9_]*::)?test[A-Za-z0-9_]*$`)

func (a scriptLanguage) ValidName(n string) bool {
	if a.id == "pytest" {
		return pythonName.MatchString(n)
	}
	return n != "" && len(n) <= 200 && !strings.ContainsFunc(n, unicode.IsControl)
}
func (a scriptLanguage) Command(p, n string, pkg bool) []string {
	mode := "test"
	if pkg {
		mode, p, n = "package", path.Dir(p), ""
	}
	if a.id == "pytest" {
		return []string{a.executable, "-I", "-c", a.script, mode, p, n}
	}
	return []string{a.executable, "--input-type=module", "--eval", a.script, mode, p, n}
}

func languageForTest(p string) LanguageAdapter {
	for _, a := range LanguageAdapters() {
		if a.Matches(p) {
			return a
		}
	}
	return nil
}
func validTestName(p, n string) bool {
	if a := languageForTest(p); a != nil {
		return a.ValidName(n)
	}
	return namePattern.MatchString(n)
}
func isTestFile(p string) bool { return languageForTest(p) != nil }

func scriptForArgv(argv []string) (scriptLanguage, bool) {
	for _, a := range LanguageAdapters()[1:] {
		s := a.(scriptLanguage)
		if len(argv) == 7 && argv[0] == s.executable && argv[3] == s.script {
			if s.id == "pytest" && argv[1] == "-I" && argv[2] == "-c" || s.id != "pytest" && argv[1] == "--input-type=module" && argv[2] == "--eval" {
				return s, true
			}
		}
	}
	return scriptLanguage{}, false
}

type languageResult struct {
	Version   int    `json:"proven_test_report"`
	Adapter   string `json:"adapter"`
	File      string `json:"file"`
	Name      string `json:"name"`
	Passed    int    `json:"passed"`
	Failed    int    `json:"failed"`
	Skipped   int    `json:"skipped"`
	Errors    int    `json:"errors"`
	Completed bool   `json:"completed"`
	Output    string `json:"output"`
}

func judgeLanguage(argv []string, out string, code int) (runKind, bool) {
	a, ok := scriptForArgv(argv)
	if !ok {
		return kindNotRun, false
	}
	var r languageResult
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &r) != nil || r.Version != 1 || r.Adapter != a.id || r.File != argv[5] || r.Name != argv[6] || !r.Completed || r.Errors != 0 || r.Passed < 0 || r.Failed < 0 || r.Skipped < 0 {
		return kindNotRun, true
	}
	if code == 0 && r.Passed > 0 && r.Failed == 0 {
		return kindPass, true
	}
	if code == 1 && r.Failed > 0 {
		return kindFail, true
	}
	return kindNotRun, true
}
