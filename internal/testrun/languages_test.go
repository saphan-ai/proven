package testrun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLanguageCommandsAndEvidenceContract(t *testing.T) {
	cases := []struct{ path, name, adapter string }{
		{"tests/test_prices.py", "test_discount", "pytest"},
		{"tests/test_prices.py", "TestPrices::test_discount", "pytest"},
		{"tests/test_prices.py", "PriceTests::testDiscount", "pytest"},
		{"src/prices.test.ts", "discount is applied once", "vitest"},
		{"src/prices.spec.tsx", "Prices renders total", "vitest"},
		{"src/prices.test.mjs", "discount.$(touch file)", "node"},
		{"src/prices.spec.js", "discount", "vitest"},
	}
	for _, tc := range cases {
		t.Run(tc.path+tc.name, func(t *testing.T) {
			argv, reason := TestArgv(tc.path, tc.name)
			if reason != "" {
				t.Fatal(reason)
			}
			a, ok := scriptForArgv(argv)
			if !ok || a.id != tc.adapter {
				t.Fatalf("adapter=%s argv=%v", a.id, argv[:3])
			}
			for _, test := range []struct {
				passed, failed, skipped, errors, code int
				completed                             bool
				want                                  runKind
			}{
				{1, 0, 0, 0, 0, true, kindPass}, {0, 1, 0, 0, 1, true, kindFail},
				{0, 0, 0, 0, 0, true, kindNotRun}, {0, 0, 1, 0, 0, true, kindNotRun},
				{1, 0, 0, 1, 0, true, kindNotRun}, {1, 0, 0, 0, 1, true, kindNotRun},
				{1, 0, 0, 0, 0, false, kindNotRun}, {0, 1, 0, 1, 1, true, kindNotRun},
				{1, 0, 0, 0, 125, true, kindNotRun},
			} {
				r := languageResult{Version: 1, Adapter: a.id, File: tc.path, Name: tc.name, Passed: test.passed, Failed: test.failed, Skipped: test.skipped, Errors: test.errors, Completed: test.completed}
				b, _ := json.Marshal(r)
				got, recognized := judgeLanguage(argv, string(b), test.code)
				if !recognized || got != test.want {
					t.Fatalf("result=%s code=%d got=%v want=%v", b, test.code, got, test.want)
				}
				if test.code == 0 && RecordedPass(displayCommand(argv), string(b)) != (test.want == kindPass) {
					t.Fatalf("recorded judgement disagrees: %s", b)
				}
			}
			wrong := `{"proven_test_report":1,"adapter":"` + a.id + `","file":"other.test.js","name":"other","passed":1,"completed":true}`
			if got, _ := judgeLanguage(argv, wrong, 0); got != kindNotRun {
				t.Fatal("accepted another test's result")
			}
			if got, _ := judgeLanguage(argv, "child log\n"+wrong, 0); got != kindNotRun {
				t.Fatal("accepted stdout as protocol")
			}
			pkg, reason := PackageArgv(tc.path, tc.name)
			if reason != "" || pkg[4] != "package" || pkg[6] != "" {
				t.Fatalf("package argv: %q %s", pkg[4:], reason)
			}
		})
	}
}

func TestNativeLanguagePathAndFixGuards(t *testing.T) {
	for _, p := range []string{"tests/test_a.py", "a.test.ts", "a.test.mjs", "a.spec.js"} {
		n := "one"
		if strings.HasSuffix(p, ".py") {
			n = "test_one"
		}
		if reason := RefuseWritten(p, "content", n, "", "", 0); reason != "" {
			t.Fatal(p, reason)
		}
		if reason := refuseEditPath(p); reason != ReasonEditsATest {
			t.Fatal(p, reason)
		}
		if _, reason := TestArgv("../"+p, n); reason == "" {
			t.Fatal("accepted traversal")
		}
		if _, reason := TestArgv("/"+p, n); reason == "" {
			t.Fatal("accepted absolute path")
		}
		if _, reason := TestArgv(p, "one\nother"); reason == "" {
			t.Fatal("accepted multiline name")
		}
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "tests")); err != nil {
		t.Fatal(err)
	}
	// Written staging still enforces the same symlink confinement as Go.
	staged, err := stageWritten(root, "tests/test_a.py", "content")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(staged)
	if _, err := os.Stat(filepath.Join(outside, "test_a.py")); !os.IsNotExist(err) {
		t.Fatal("staging wrote through the checkout link")
	}
	if b, err := os.ReadFile(filepath.Join(staged, "tests/test_a.py")); err != nil || string(b) != "content" {
		t.Fatalf("staged file: %q %v", b, err)
	}
	argvs := BreakageArgvs([]Edit{{Path: "src/a.py"}, {Path: "src/a.ts"}, {Path: "package.json"}})
	if len(argvs) != 2 {
		t.Fatalf("wanted pytest + Vitest, got %d", len(argvs))
	}
	for _, argv := range argvs {
		if _, ok := scriptForArgv(argv); !ok || argv[4] != "package" || argv[5] != "." {
			t.Fatal("not a bounded package recheck")
		}
	}
}
