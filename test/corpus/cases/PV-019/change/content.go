package change

import (
	"fmt"

	"pv019/table"
)

// PinnedChannel is the new editor's failure channel, as measured.
const PinnedChannel = "exit code 2"

// NewEditor is the editor the change adds.
const NewEditor = "ed-cedar"

// ContentTest is the content test shipped with the change: it pins the new
// editor's failure channel to its measured value.
func ContentTest(t table.Table) error {
	for _, row := range t {
		if row.Editor == NewEditor {
			if row.FailureChannel != PinnedChannel {
				return fmt.Errorf("%s failure channel = %q, want %q", NewEditor, row.FailureChannel, PinnedChannel)
			}
			return nil
		}
	}
	return fmt.Errorf("%s not in table", NewEditor)
}
