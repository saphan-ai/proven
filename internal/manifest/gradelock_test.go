package manifest

import (
	"strings"
	"testing"
)

// TestGradleLockReadsPins proves a dependency line is a pin, a lowered
// version is not a raise, and empty= is not a package.
func TestGradleLockReadsPins(t *testing.T) {
	base := []byte(strings.Join([]string{
		"# This is a Gradle generated file for dependency locking.",
		"com.example:quiet:2.0.0=compileClasspath,runtimeClasspath",
		"com.example:widget:1.2.0=compileClasspath,runtimeClasspath",
		"empty=annotationProcessor",
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"# This is a Gradle generated file for dependency locking.",
		"com.example:quiet:1.9.0=compileClasspath,runtimeClasspath",
		"com.example:widget:1.4.0=compileClasspath,runtimeClasspath",
		"com.example:added:1.0.0=compileClasspath",
		"empty=",
		"",
	}, "\n"))
	got, err := diffGradleLock(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "com.example:added", Version: "1.0.0", Line: 4, Added: true},
		{Package: "com.example:widget", Version: "1.4.0", Line: 3, Added: false},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
}

// TestGradleLockRejectsAFileThatIsNotALock proves a lock file that is not
// one refuses, and the error names the side.
func TestGradleLockRejectsAFileThatIsNotALock(t *testing.T) {
	if _, err := diffGradleLock([]byte("this is not a lock\n"), []byte("empty=\n")); err == nil || !strings.Contains(err.Error(), "base:") {
		t.Fatal(err)
	}
	if _, err := diffGradleLock([]byte("# only a comment\n"), []byte("empty=\n")); err == nil || !strings.Contains(err.Error(), "base:") {
		t.Fatal(err)
	}
}

// TestGradleLockMatch proves the reader matches gradle.lockfile at any depth
// and not a manifest or a backup.
func TestGradleLockMatch(t *testing.T) {
	r := GradleLock()
	if r.Ecosystem != "Maven" || !r.Lock || !r.NoteErr || r.Compare == nil || r.Explain == nil {
		t.Fatal(r.Ecosystem)
	}
	if !r.Match("gradle.lockfile") || !r.Match("app/gradle.lockfile") {
		t.Fatal("missed gradle.lockfile")
	}
	if r.Match("gradle.lockfile.bak") || r.Match("build.gradle.kts") || r.Match("libs.versions.toml") {
		t.Fatal("matched a file that is not gradle.lockfile")
	}
}
