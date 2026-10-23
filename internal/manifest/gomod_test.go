package manifest

import (
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

func TestGoDiffAddsAndRaises(t *testing.T) {
	base := []byte("module example.com/shop\n\ngo 1.26\n\nrequire example.com/old v1.0.0\n")
	change := []byte(`module example.com/shop

go 1.26

require (
	example.com/old v1.2.0
	example.com/new v0.1.0 // indirect
	example.com/lower v0.9.0
)
`)
	base = append(base, []byte("require example.com/lower v1.0.0\n")...)
	got, err := diffGo(base, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("%#v", got)
	}
	if got[0].Package != "example.com/new" || !got[0].Added || got[0].Version != "v0.1.0" || got[0].Line != 7 {
		t.Fatalf("added %#v", got[0])
	}
	if got[1].Package != "example.com/old" || got[1].Added || got[1].Version != "v1.2.0" || got[1].Line != 6 {
		t.Fatalf("raised %#v", got[1])
	}
}

func TestGoDiffIgnoresAReplace(t *testing.T) {
	base := []byte("module m\n\ngo 1.26\n")
	change := []byte("module m\n\ngo 1.26\n\nrequire example.com/a v1.0.0\nreplace example.com/a => example.com/a v1.2.0\n")
	got, err := diffGo(base, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Version != "v1.0.0" {
		t.Fatalf("%#v", got)
	}
}

func TestCollectRunsEveryReader(t *testing.T) {
	extra := Reader{
		Ecosystem: "Demo",
		Match:     func(path string) bool { return path == "deps.txt" },
		Diff: func(base, change []byte) ([]Change, error) {
			if len(change) == 0 {
				return nil, nil
			}
			return []Change{{Package: "demo", Version: "1", Line: 1, Added: true}}, nil
		},
	}
	got, notes, err := Collect([]change.File{
		{Path: "go.mod", Change: []byte("module m\n\ngo 1.26\n\nrequire example.com/a v1.0.0\n")},
		{Path: "deps.txt", Change: []byte("demo\n")},
	}, []Reader{Go(), extra})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 {
		t.Fatalf("notes %#v", notes)
	}
	if len(got) != 2 || got[0].Ecosystem != "Demo" || got[1].Ecosystem != "Go" {
		t.Fatalf("%#v", got)
	}
}
