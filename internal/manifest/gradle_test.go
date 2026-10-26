package manifest

import (
	"strings"
	"testing"
)

// TestGradleCatalogResolvesRefs proves a version.ref is the [versions]
// string, a literal coordinate string is that coordinate, a missing ref
// and a plugin are notes, and a bundle is not a dependency.
func TestGradleCatalogResolvesRefs(t *testing.T) {
	base := []byte(strings.Join([]string{
		"[versions]",
		"widget = \"1.2.0\"",
		"quiet = \"2.0.0\"",
		"",
		"[libraries]",
		"widget = { module = \"com.example:widget\", version.ref = \"widget\" }",
		"quiet = { group = \"com.example\", name = \"quiet\", version.ref = \"quiet\" }",
		"",
	}, "\n"))
	change := []byte(strings.Join([]string{
		"[libraries]",
		"widget = { module = \"com.example:widget\", version.ref = \"widget\" }",
		"quiet = { group = \"com.example\", name = \"quiet\", version = \"1.9.0\" }",
		"added = \"com.example:added:1.0.0\"",
		"missing = { module = \"com.example:missing\", version.ref = \"nope\" }",
		"rich = { module = \"com.example:rich\", version = { strictly = \"[1,2)\" } }",
		"bom = { group = \"com.example\", name = \"bom\" }",
		"plain = 1",
		"",
		"[versions]",
		"widget = \"1.4.0\"",
		"",
		"[bundles]",
		"web = [\"widget\", \"quiet\"]",
		"",
		"[plugins]",
		"android = { id = \"com.android.application\", version.ref = \"agp\" }",
		"",
	}, "\n"))
	got, err := diffGradleCatalog(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "com.example:added", Version: "1.0.0", Line: 4, Added: true},
		{Package: "com.example:widget", Version: "1.4.0", Line: 2, Added: false},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("got[%d] %#v want %#v", i, got[i], w)
		}
	}
	notes := strings.Join(gradleCatalogNotes(change), "\n")
	for _, want := range []string{
		"libraries.missing version.ref nope is not in [versions]",
		"libraries.rich version is not a string",
		"libraries.bom has no version",
		"libraries.plain is not a string",
		"plugins.android is a Gradle plugin, not a Maven coordinate",
	} {
		if !strings.Contains(notes, want) {
			t.Fatalf("notes %q missing %q", notes, want)
		}
	}
	if strings.Contains(notes, "widget") && strings.Contains(notes, "libraries.widget") {
		t.Fatal(notes)
	}
}

// TestGradleCatalogRejectsABrokenLibrary proves a libraries line that is
// not an assignment fails the check instead of reading as no dependency.
func TestGradleCatalogRejectsABrokenLibrary(t *testing.T) {
	if _, err := diffGradleCatalog(nil, []byte("[libraries]\nnot a dependency\n")); err == nil {
		t.Fatal("broken catalog was accepted")
	}
	if _, err := diffGradleCatalog(nil, []byte("[metadata]\nformat = \"1\"\n")); err == nil {
		t.Fatal("unknown table was accepted")
	}
}

// TestGradleCatalogMatch proves the reader matches libs.versions.toml at any
// depth and not a backup or another Gradle file.
func TestGradleCatalogMatch(t *testing.T) {
	r := GradleCatalog()
	if r.Ecosystem != "Maven" || r.Lock || r.Compare == nil || r.Explain == nil {
		t.Fatal(r.Ecosystem)
	}
	if !r.Match("libs.versions.toml") || !r.Match("gradle/libs.versions.toml") {
		t.Fatal("missed libs.versions.toml")
	}
	if r.Match("libs.versions.toml.bak") || r.Match("build.gradle.kts") || r.Match("gradle.lockfile") {
		t.Fatal("matched a file that is not libs.versions.toml")
	}
}
