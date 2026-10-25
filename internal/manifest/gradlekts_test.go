package manifest

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestGradleKtsReadsLiteralCoordinates proves a literal coordinate is a
// dependency, a "$" version and a plugin and a libs reference are notes,
// and a comment is not a coordinate.
func TestGradleKtsReadsLiteralCoordinates(t *testing.T) {
	base := []byte(strings.Join([]string{
		"dependencies {",
		"    implementation(\"com.example:widget:1.2.0\")",
		"    implementation(\"com.example:quiet:2.0.0\")",
		"}",
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"plugins {",
		"    id(\"com.android.application\") version \"8.2.0\"",
		"}",
		"dependencies {",
		"    implementation(\"com.example:widget:1.4.0\")",
		"    implementation(\"com.example:quiet:1.9.0\")",
		"    implementation(\"com.example:variable:$widgetVersion\")",
		"    implementation(libs.widget)",
		"    // implementation(\"com.example:commented:9.0.0\")",
		"    implementation(group = \"com.example\", name = \"named\", version = \"3.0.0\")",
		"}",
		"",
	}, "\n"))
	got, err := diffGradleKts(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "com.example:named", Version: "3.0.0", Line: 10, Added: true},
		{Package: "com.example:widget", Version: "1.4.0", Line: 5, Added: false},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
	notes := strings.Join(gradleKtsNotes(change), "\n")
	for _, want := range []string{
		"plugin com.android.application is not resolved to a Maven coordinate",
		"com.example:variable version is a variable",
		"libs.widget is a catalog reference, not a literal coordinate",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("notes %q missing %q", notes, want)
		}
	}
	if strings.Contains(notes, "commented") || strings.Contains(notes, "named") {
		t.Fatal(notes)
	}
}

// TestGradleKtsMatch proves the reader matches build.gradle.kts at any depth
// and not a Groovy build file, a settings file, or a backup.
func TestGradleKtsMatch(t *testing.T) {
	r := GradleKts()
	if r.Ecosystem != "Maven" || r.Lock || r.Compare == nil {
		t.Fatal(r.Ecosystem)
	}
	if !r.Match("build.gradle.kts") || !r.Match("app/build.gradle.kts") {
		t.Fatal("missed build.gradle.kts")
	}
	if r.Match("build.gradle") || r.Match("settings.gradle.kts") || r.Match("build.gradle.kts.bak") {
		t.Fatal("matched a file that is not build.gradle.kts")
	}
}

// TestGradleLockWinsOverTheKotlinManifest proves the lock file's version is
// the one kept when it and build.gradle.kts name the same package.
func TestGradleLockWinsOverTheKotlinManifest(t *testing.T) {
	files := []change.File{
		{Path: "build.gradle.kts", Change: []byte("dependencies {\n    implementation(\"com.example:widget:1.4.0\")\n    implementation(\"com.example:only:1.0.0\")\n}\n")},
		{Path: "gradle.lockfile", Change: []byte("com.example:widget:1.4.0=compileClasspath\nempty=\n")},
	}
	got, notes, err := Collect(files, []Reader{GradleKts(), GradleLock()})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	if got[0].File != "build.gradle.kts" || got[0].Package != "com.example:only" || got[0].Lock {
		t.Fatalf("%#v", got[0])
	}
	if got[1].File != "gradle.lockfile" || got[1].Package != "com.example:widget" || got[1].Version != "1.4.0" || !got[1].Lock {
		t.Fatalf("%#v", got[1])
	}
}
