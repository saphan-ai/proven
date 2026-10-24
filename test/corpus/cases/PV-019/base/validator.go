// Package base is the validator before the change: two statuses, unknown refused.
package base

import (
	"fmt"

	"pv019/table"
)

// Validate refuses a row whose status is not supported or unsupported.
func Validate(t table.Table) table.Result {
	var r table.Result
	for _, row := range t {
		switch row.Status {
		case "supported", "unsupported":
		default:
			r.Refusals = append(r.Refusals, fmt.Sprintf("%s: unknown status %q", row.Editor, row.Status))
		}
	}
	return r
}
