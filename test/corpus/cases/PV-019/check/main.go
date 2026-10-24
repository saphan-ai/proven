// Command check runs the case's inputs against base and change and prints one
// line per input; it also runs the shipped content test on each input that
// changes the failure channel.
package main

import (
	"fmt"
	"strings"

	"pv019/base"
	"pv019/change"
	"pv019/table"
)

// closedChannels is the reference's closed set of failure-channel values.
var closedChannels = map[string]bool{"exit code 1": true, "exit code 2": true, "stderr line": true, "json error": true}

// reference applies the documented meaning of each status. The validators never see it.
func reference(t table.Table) table.Result {
	var r table.Result
	for _, row := range t {
		bad := func(why string) { r.Refusals = append(r.Refusals, row.Editor+": "+why) }
		switch row.Status {
		case "supported":
		case "unsupported":
			if row.InstallRoute != "" {
				bad("unsupported row has an install route")
			}
		case "preview":
			if row.InstallRoute != "" || row.SessionSwitch || row.PluginFolder != "" {
				bad("preview row has a delivery field")
			}
		default:
			bad("unknown status")
		}
		if !closedChannels[row.FailureChannel] {
			bad(fmt.Sprintf("failure channel %q is not in the closed set", row.FailureChannel))
		}
	}
	return r
}

func show(r table.Result) string {
	if r.Pass() {
		return "pass"
	}
	return "fail [" + strings.Join(r.Refusals, "; ") + "]"
}

// edit returns the shipped table with the named editor's row changed.
func edit(editor string, f func(*table.Row)) table.Table {
	t := table.Shipped().Clone()
	for i := range t {
		if t[i].Editor == editor {
			f(&t[i])
		}
	}
	return t
}

func main() {
	type input struct {
		name           string
		t              table.Table
		changesChannel bool
	}
	inputs := []input{
		{"shipped-table", table.Shipped(), false},
		{"preview-with-install-route", edit("ed-cedar", func(r *table.Row) { r.InstallRoute = "pkg:ed-cedar" }), false},
		{"preview-with-session-switch", edit("ed-cedar", func(r *table.Row) { r.SessionSwitch = true }), false},
		{"preview-with-plugin-folder", edit("ed-cedar", func(r *table.Row) { r.PluginFolder = "plugins/cedar" }), false},
		{"unsupported-with-install-route", edit("ed-birch", func(r *table.Row) { r.InstallRoute = "pkg:ed-birch" }), false},
		{"channel-wrong-nonempty", edit("ed-cedar", func(r *table.Row) { r.FailureChannel = "ignored silently" }), true},
		{"channel-empty", edit("ed-cedar", func(r *table.Row) { r.FailureChannel = "" }), true},
	}
	for _, in := range inputs {
		fmt.Printf("%s | base: %s | change: %s | reference: %s\n", in.name, show(base.Validate(in.t)), show(change.Validate(in.t)), show(reference(in.t)))
		if in.changesChannel {
			if err := change.ContentTest(in.t); err != nil {
				fmt.Printf("%s | shipped content test: fail [%v]\n", in.name, err)
			} else {
				fmt.Printf("%s | shipped content test: pass\n", in.name)
			}
		}
	}
}
