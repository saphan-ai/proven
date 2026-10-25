package formatter

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestLinesMatchGolden proves three properties of Lines:
// one finding is one line "path:line: title";
// a control byte, a DEL, a C1 rune, a newline and an invalid UTF-8 byte
// in the path or the title are escaped before they reach the writer;
// an opinion and a message are not lines.
// The golden file is the expected bytes. One changed byte fails this test.
func TestLinesMatchGolden(t *testing.T) {
	result := review.Result{
		Findings: []review.Finding{
			{
				ID:       "f1",
				Title:    "discount ignores the cap" + "\x1b",
				Message:  "not printed",
				Severity: review.SeverityError,
				Location: review.Location{Path: "billing/discount.go", StartLine: 12, EndLine: 14},
			},
			{
				ID:       "f2",
				Title:    "reached" + "\n" + "\u009b" + "\xff",
				Message:  "not printed either",
				Severity: review.SeverityWarning,
				Location: review.Location{Path: "a" + "\x7f" + ".go", StartLine: 4},
			},
		},
		Opinions: []review.Opinion{{
			ID:       "o1",
			Title:    "agreed only",
			Location: review.Location{Path: "billing/discount.go", StartLine: 1},
		}},
	}
	var buf bytes.Buffer
	if err := (Lines{}).Write(&buf, result); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := buf.Bytes()
	want, err := os.ReadFile("testdata/lines.golden")
	if err != nil {
		t.Fatalf("golden: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("lines differ\n got: %q\nwant: %q", got, want)
	}
	for i := 0; i < len(got); i++ {
		c := got[i]
		if (c < 0x20 && c != '\n') || c == 0x7f || (c >= 0x80 && c <= 0x9f) || c == 0xff {
			t.Fatalf("byte %02x at %d left the change unescaped", c, i)
		}
	}
}

// TestEmptyResultNamesItself proves a review with no findings is the line
// "findings: none", so an empty body is not how a reader sees a clean review.
func TestEmptyResultNamesItself(t *testing.T) {
	var buf bytes.Buffer
	if err := (Lines{}).Write(&buf, review.Result{}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	const want = "findings: none\n"
	if buf.String() != want {
		t.Fatalf("got %q, want %q", buf.String(), want)
	}
}

// TestWriteReturnsTheWriterError proves Write returns the destination's
// error unchanged, so the caller can match it with errors.Is.
func TestWriteReturnsTheWriterError(t *testing.T) {
	err := (Lines{}).Write(failWriter{}, review.Result{})
	if !errors.Is(err, errDisk) {
		t.Fatalf("got %v, want %v", err, errDisk)
	}
}

// errDisk is the error failWriter returns. The test matches this value.
var errDisk = errors.New("disk full")

// failWriter refuses every write with errDisk.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errDisk }
