package testrun

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

// goTest is the command words TestArgv builds for the test used across these tests.
var goTest = []string{"go", "test", "-json", "-run", "^TestDiscounted$", "./billing"}

// TestArgvIsTheTemplateWithNoShell proves a valid test name and path give
// exactly the fixed per-language command, as separate words with no shell.
func TestArgvIsTheTemplateWithNoShell(t *testing.T) {
	cases := []struct {
		path, name string
		want       []string
	}{
		{"billing/discount_test.go", "TestDiscounted", goTest},
		{"add_test.go", "TestAdd", []string{"go", "test", "-json", "-run", "^TestAdd$", "."}},
		{"a/b/c_test.go", "Test_x1", []string{"go", "test", "-json", "-run", "^Test_x1$", "./a/b"}},
	}
	for _, tc := range cases {
		got, reason := TestArgv(tc.path, tc.name)
		if reason != "" || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s %s: got %q (%q), want %q", tc.path, tc.name, got, reason, tc.want)
		}
		for _, word := range got {
			if word == "sh" || word == "-c" {
				t.Errorf("a shell in %q", got)
			}
		}
	}
}

// TestDisplayCommandIsTheReadableTemplate proves the trace shows the template
// with the pattern quoted, and that it is only text.
func TestDisplayCommandIsTheReadableTemplate(t *testing.T) {
	if got, want := displayCommand(goTest), "go test -json -run '^TestDiscounted$' ./billing"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

// TestRefuseTestKeepsNamesAndPathsInsideThePattern proves a name or a path
// that could mean more than one test in one directory gets no command.
func TestRefuseTestKeepsNamesAndPathsInsideThePattern(t *testing.T) {
	cases := []struct {
		label, path, name, reason string
	}{
		{"empty name", "a_test.go", "", ReasonBadName},
		{"shell text", "a_test.go", "TestA; id", ReasonBadName},
		{"regexp", "a_test.go", "Test.*", ReasonBadName},
		{"alternation", "a_test.go", "TestA|TestB", ReasonBadName},
		{"flag", "a_test.go", "-test.run=x", ReasonBadName},
		{"not a Test function", "a_test.go", "BenchmarkA", ReasonBadName},
		{"newline", "a_test.go", "TestA\nid", ReasonBadName},
		{"lowercase after Test", "a_test.go", "Testa", ReasonBadName},
		{"absolute", "/a/b_test.go", "TestA", ReasonAbsolutePath},
		{"leaves", "../a_test.go", "TestA", ReasonLeavesRepo},
		{"not a test file", "a.go", "TestA", ReasonNotATestFile},
		{"no path", "", "TestA", ReasonNotATestFile},
		{"wildcard directory", "a.../b_test.go", "TestA", ReasonBadPackage},
		{"flag directory", "-x/b_test.go", "TestA", ReasonBadPackage},
		{"space in directory", "a b/c_test.go", "TestA", ReasonBadPackage},
		{"quote in directory", "a'b/c_test.go", "TestA", ReasonBadPackage},
		{"dollar in directory", "$HOME/c_test.go", "TestA", ReasonBadPackage},
		{"empty segment", "a//c_test.go", "TestA", ReasonBadPackage},
	}
	for _, tc := range cases {
		if got := RefuseTest(tc.path, tc.name); got != tc.reason {
			t.Errorf("%s: reason %q, want %q", tc.label, got, tc.reason)
		}
		if argv, reason := TestArgv(tc.path, tc.name); argv != nil || reason != tc.reason {
			t.Errorf("%s: TestArgv gave %q (%q)", tc.label, argv, reason)
		}
	}
}

// TestRunnerStartsNoContainerForAHostileName proves a name with shell text is
// refused with a named reason before any container is started, for a named
// test and for a written one.
func TestRunnerStartsNoContainerForAHostileName(t *testing.T) {
	const hostile = "TestA'; printf ready; '"
	for _, content := range []string{"", "package billing\n"} {
		c := mustContainer(t, DefaultContainerConfig())
		started := 0
		c.run = func(context.Context, []string) (string, error) { started++; return "", nil }
		out, err := c.Run(context.Background(), Request{
			Path: "billing/discount_test.go", Name: hostile, Content: content,
			BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
		})
		if err != nil {
			t.Fatal(err)
		}
		if out.Found || started != 0 || !strings.Contains(out.Output, "refused: "+ReasonBadName) {
			t.Fatalf("content %q: found %v, containers started %d, output %q", content, out.Found, started, out.Output)
		}
	}
}

// TestRunnerRunsExactlyTheTemplate proves the line the runtime receives ends
// with the image and the template words, and that the outcome shows the
// readable template.
func TestRunnerRunsExactlyTheTemplate(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	var runs [][]string
	c.run = func(_ context.Context, argv []string) (string, error) {
		runs = append(runs, argv)
		return "", nil
	}
	out, err := c.Run(context.Background(), Request{
		Path: "billing/discount_test.go", Name: "TestDiscounted",
		BaseDir: t.TempDir(), ChangeDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("runtime runs %d", len(runs))
	}
	for _, argv := range runs {
		tail := argv[len(argv)-len(goTest)-1:]
		if tail[0] != DefaultContainerConfig().Image || !reflect.DeepEqual(tail[1:], goTest) {
			t.Errorf("the run line ends %q", tail)
		}
	}
	if out.Command != "go test -json -run '^TestDiscounted$' ./billing" {
		t.Errorf("outcome command %q", out.Command)
	}
}
