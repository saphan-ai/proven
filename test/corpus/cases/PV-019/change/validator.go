// Package change is the validator after the change: it adds the preview status
// and a check that the failure channel is not empty.
package change

import (
	"fmt"

	"pv019/table"
)

// Validate refuses a row that breaks the shape rules of its status.
func Validate(t table.Table) table.Result {
	var r table.Result
	refuse := func(row table.Row, why string) {
		r.Refusals = append(r.Refusals, fmt.Sprintf("%s: %s", row.Editor, why))
	}
	for _, row := range t {
		switch row.Status {
		case "supported":
		case "unsupported":
			// no delivery field is checked
		case "preview":
			if row.InstallRoute != "" {
				refuse(row, "preview row has an install route")
			}
		default:
			refuse(row, fmt.Sprintf("unknown status %q", row.Status))
		}
		if row.FailureChannel == "" {
			refuse(row, "failure channel is empty")
		}
	}
	return r
}
