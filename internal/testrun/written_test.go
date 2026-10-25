package testrun

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestWrittenTestIsCopiedAndTheCheckoutStaysUntouched proves a written test
// runs from copies of both checkouts, leaves the caller's trees unchanged,
// and removes the copies afterwards.
func TestWrittenTestIsCopiedAndTheCheckoutStaysUntouched(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "keep.txt"), []byte("base-keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "keep.txt"), []byte("change-keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	const content = "package demo\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {}\n"
	c := mustContainer(t, DefaultContainerConfig())
	var mounts []string
	sawFile := 0
	c.run = func(ctx context.Context, argv []string) (string, error) {
		if len(argv) > 1 && argv[1] == "rm" {
			return "", nil
		}
		host := mountHost(argv)
		mounts = append(mounts, host)
		if filepath.Clean(host) == filepath.Clean(base) || filepath.Clean(host) == filepath.Clean(change) {
			t.Fatalf("mounted the user's tree %s", host)
		}
		got, err := os.ReadFile(filepath.Join(host, "add_test.go"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != content {
			t.Fatalf("copy content %q", got)
		}
		keep, err := os.ReadFile(filepath.Join(host, "keep.txt"))
		if err != nil {
			t.Fatal(err)
		}
		sawFile++
		switch string(keep) {
		case "change-keep":
			return "FAIL TestAdd", exitCode(t, 1)
		case "base-keep":
			return passedRun(argv), nil
		default:
			t.Fatalf("keep %q in %s", keep, host)
			return "", nil
		}
	}
	out, err := c.Run(context.Background(), Request{
		Path: "add_test.go", Name: "TestAdd",
		Content: content, BaseDir: base, ChangeDir: change,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found || !out.FailsOnChange || !out.PassesOnBase || sawFile != 2 {
		t.Fatalf("saw %d %#v", sawFile, out)
	}
	for _, dir := range []string{base, change} {
		if _, statErr := os.Stat(filepath.Join(dir, "add_test.go")); !os.IsNotExist(statErr) {
			t.Fatalf("wrote %s", dir)
		}
		keep, readErr := os.ReadFile(filepath.Join(dir, "keep.txt"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if dir == base && string(keep) != "base-keep" {
			t.Fatalf("base keep %q", keep)
		}
		if dir == change && string(keep) != "change-keep" {
			t.Fatalf("change keep %q", keep)
		}
	}
	for _, host := range mounts {
		if _, statErr := os.Stat(host); !os.IsNotExist(statErr) {
			t.Fatalf("copy remains %s", host)
		}
	}
}

// TestContainerRefusesAWrittenPathBeforeItRuns proves each unsafe written test
// is refused with its reason and starts no container.
func TestContainerRefusesAWrittenPathBeforeItRuns(t *testing.T) {
	c := mustContainer(t, DefaultContainerConfig())
	c.run = func(context.Context, []string) (string, error) {
		t.Fatal("started")
		return "", nil
	}
	base := t.TempDir()
	change := t.TempDir()
	cases := []struct {
		name     string
		path     string
		content  string
		testName string
		reason   string
		prep     func()
	}{
		{"absolute", "/tmp/add_test.go", "package demo\n", "TestAdd", ReasonAbsolutePath, nil},
		{"leaves", "../add_test.go", "package demo\n", "TestAdd", ReasonLeavesRepo, nil},
		{"not a test", "add.go", "package demo\n", "TestAdd", ReasonNotATestFile, nil},
		{"empty name", "add_test.go", "package demo\n", "", ReasonBadName, nil},
		{"name with shell text", "add_test.go", "package demo\n", "TestAdd; id", ReasonBadName, nil},
		{"path with a pattern", "a.../add_test.go", "package demo\n", "TestAdd", ReasonBadPackage, nil},
		{"too big", "add_test.go", strings.Repeat("x", 8), "TestAdd", ReasonTooLarge, func() {
			c.cfg.MaxWrittenBytes = 4
		}},
		{"overwrite", "add.go", "package demo\n", "TestAdd", ReasonOverwrites, func() {
			if err := os.WriteFile(filepath.Join(change, "add.go"), []byte("package demo\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.cfg.MaxWrittenBytes = DefaultMaxWrittenBytes
			if tc.prep != nil {
				tc.prep()
			}
			out, err := c.Run(context.Background(), Request{
				Path: tc.path, Content: tc.content,
				BaseDir: base, ChangeDir: change, Name: tc.testName,
			})
			if err != nil {
				t.Fatal(err)
			}
			if out.Found || out.FailsOnChange || out.PassesOnBase {
				t.Fatalf("%#v", out)
			}
			if !strings.Contains(out.Output, "refused: "+tc.reason) {
				t.Fatalf("output %q", out.Output)
			}
		})
	}
}

// TestWrittenGoTestRunsInTheContainer proves, in a real container, that a
// written test proves a defect the change brings in, and that neither
// checkout gets the file. It runs only with PROVEN_TEST_DOCKER=1.
func TestWrittenGoTestRunsInTheContainer(t *testing.T) {
	if os.Getenv("PROVEN_TEST_DOCKER") != "1" {
		t.Skip("PROVEN_TEST_DOCKER=1")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker")
	}
	base := t.TempDir()
	change := t.TempDir()
	mod := "module demo\n\ngo 1.26\n"
	for _, dir := range []string{base, change} {
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(base, "add.go"), []byte("package demo\n\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "add.go"), []byte("package demo\n\nfunc Add(a, b int) int { return a + b + 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "package demo\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif got := Add(1, 1); got != 2 {\n\t\tt.Fatalf(\"Add(1, 1) = %d\", got)\n\t}\n}\n"
	c, err := NewContainer(DefaultContainerConfig(), exec.LookPath)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Run(context.Background(), Request{
		Path: "add_test.go", Name: "TestAdd",
		Content: content, BaseDir: base, ChangeDir: change,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.FailsOnChange || !out.PassesOnBase {
		t.Fatalf("%#v\n%s", out, out.Output)
	}
	if _, statErr := os.Stat(filepath.Join(base, "add_test.go")); !os.IsNotExist(statErr) {
		t.Fatal("wrote the base checkout")
	}
	if _, statErr := os.Stat(filepath.Join(change, "add_test.go")); !os.IsNotExist(statErr) {
		t.Fatal("wrote the change checkout")
	}
}
