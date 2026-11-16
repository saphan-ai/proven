package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeliveryMetadataIsExplicitBoundedData(t *testing.T) {
	if got, err := loadDelivery(nil); got != nil || err != nil {
		t.Fatal(got, err)
	}
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`{"title":"Reject negatives","description":"Do it","commits":[{"sha":"abc","message":"Add a guard"}]}`, true},
		{`{"command":"touch secret"}`, false}, {`{} {}`, false}, {strings.Repeat("x", (64<<10)+1), false},
	} {
		file := filepath.Join(t.TempDir(), "context.json")
		if err := os.WriteFile(file, []byte(tc.body), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := loadDelivery(func(string) string { return file })
		if (err == nil) != tc.valid {
			t.Fatalf("valid=%v result=%+v err=%v", tc.valid, got, err)
		}
	}
	file := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink("missing", file); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDelivery(func(string) string { return file }); err == nil {
		t.Fatal("accepted linked metadata")
	}
}
