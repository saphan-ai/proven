package evidence

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

// The pipeline redacts a finding's text and its refutation rows after the Judge has decided, so
// the Result that reaches Build is already redacted. Two different configured secrets, one in the
// message and one in a refutation row that says no, are not about one another when the Judge reads
// them, and both become a marker that names its variable ("[redacted: NAME]"). The marker words
// then match. The recorded outcome (F3 reported) is the Judge's, made on the text before
// redaction; the rule over the redacted text drops F3. An honest pack is written and names F3: it
// is not refused at build.
func TestRedactionBeforeBuildMovesOutcomeAndIsNamed(t *testing.T) {
	in, pub := forgeryInput(t)
	secretA, secretB := strings.Repeat("A", 32), strings.Repeat("B", 32)
	rd := redact.New(map[string]string{"TOKEN_A": secretA, "TOKEN_B": secretB})

	f := &in.Result.Findings[2]
	f.Message = "pkg-a 1.2.0 is in an advisory; the lock file pins " + secretA
	f.Refutation = &review.Refutation{QueryRows: [][]string{{secretB, "no"}}}
	judged, err := review.Rederive(*f)
	if err != nil {
		t.Fatal(err)
	}
	if !judged.Reported || judged.Unresolved || judged.Level != f.ProofLevel {
		t.Fatalf("the Judge's reading of the text before redaction should keep F3 reported, got %+v", judged)
	}
	// What the pipeline hands over: the same finding with every configured value already replaced.
	f.Message = rd.Text(f.Message)
	f.Refutation.QueryRows = [][]string{{rd.Text(secretB), "no"}}
	if after, err := review.Rederive(*f); err != nil || after.Reported {
		t.Fatalf("the sample does not reproduce N1: the rule over the redacted text should drop F3 (got %+v, %v)", after, err)
	}

	in.Redactor = rd
	pack, err := Build(in)
	if err != nil {
		t.Fatalf("an honest pack whose finding text was redacted before the build is refused at build: %v", err)
	}
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatalf("an honest pack is refused by verify: %v", err)
	}
	var named string
	for _, s := range rep.NotRederived {
		if strings.Contains(s, "finding F3:") {
			named = s
		}
	}
	if !strings.Contains(named, "no longer holds the text") {
		t.Fatalf("F3 should be named: its text is not the text the rule matched, got %q (all: %q)", named, rep.NotRederived)
	}
}
