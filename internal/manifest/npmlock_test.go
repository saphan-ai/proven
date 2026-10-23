package manifest

import (
	"strings"
	"testing"
)

func TestNPMLockDiffAddedRaisedUnchangedRemoved(t *testing.T) {
	base := []byte(`{
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "app", "version": "0.1.0"},
    "node_modules/gone": {"version": "1.0.0"},
    "node_modules/lower": {"version": "2.0.0"},
    "node_modules/old": {"version": "1.0.0"},
    "node_modules/same": {"version": "1.2.3"}
  }
}
`)
	change := []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "": {"name": "app", "version": "0.2.0"},
    "node_modules/lower": {"version": "1.0.0"},
    "node_modules/old": {"version": "1.2.0"},
    "node_modules/same": {"version": "1.2.3"},
    "node_modules/added": {"version": "1.0.0"},
    "node_modules/@scope/pkg": {"version": "1.4.0"},
    "node_modules/foo/node_modules/nested": {"version": "9.0.0", "name": "nested"},
    "node_modules/linked": {"link": true}
  }
}
`)
	got, err := diffNPMLock(base, change)
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{Package: "@scope/pkg", Version: "1.4.0", Line: 9, Added: true},
		{Package: "added", Version: "1.0.0", Line: 8, Added: true},
		{Package: "nested", Version: "9.0.0", Line: 10, Added: true},
		{Package: "old", Version: "1.2.0", Line: 6, Added: false},
	}
	if len(got) != len(want) {
		t.Fatalf("%#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got[i], want[i])
		}
	}
}

func TestNPMLockReportsEveryNewVersion(t *testing.T) {
	change := []byte(`{
  "lockfileVersion": 3,
  "packages": {
    "node_modules/left-pad": {"version": "1.0.0"},
    "node_modules/foo/node_modules/left-pad": {"version": "1.2.3"}
  }
}
`)
	got, err := diffNPMLock(nil, change)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Package != "left-pad" || got[1].Package != "left-pad" {
		t.Fatalf("%#v", got)
	}
	if got[0].Version != "1.0.0" || got[1].Version != "1.2.3" {
		t.Fatalf("%#v", got)
	}
}

func TestNPMLockMalformed(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"broken", "{", "invalid JSON"},
		{"array", "[]", "not an object"},
		{"version", `{"lockfileVersion":1,"packages":{}}`, "lockfileVersion 1 is not 2 or 3"},
		{"packages", `{"lockfileVersion":3}`, "packages is missing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := diffNPMLock(nil, []byte(tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("%v", err)
			}
		})
	}
	got, err := diffNPMLock(nil, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty: %v %#v", err, got)
	}
}

func TestNPMLockMatch(t *testing.T) {
	r := NPMLock()
	if r.Ecosystem != "npm" || !r.Lock || !r.NoteErr || !r.Match("package-lock.json") || !r.Match("svc/package-lock.json") {
		t.Fatal(r.Ecosystem)
	}
	if r.Match("package.json") || r.Match("package-lock.json.bak") || r.Match("yarn.lock") || r.Match("npm-shrinkwrap.json") {
		t.Fatal("matched a file that is not package-lock.json")
	}
}
