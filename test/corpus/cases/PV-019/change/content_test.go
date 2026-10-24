package change

import (
	"testing"

	"pv019/table"
)

func TestNewEditorFailureChannelIsPinned(t *testing.T) {
	if err := ContentTest(table.Shipped()); err != nil {
		t.Fatal(err)
	}
}
