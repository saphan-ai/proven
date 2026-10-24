// Package table holds the capability table both validators read.
package table

// Row is one editor the build tool can integrate with.
type Row struct {
	Editor         string
	Status         string
	InstallRoute   string
	SessionSwitch  bool
	PluginFolder   string
	FailureChannel string
}

// Table is a small table held as data.
type Table []Row

// Shipped returns the table as the change ships it. Every delivery field of the
// one preview row is empty, so the shipped data is clean.
func Shipped() Table {
	return Table{
		{Editor: "ed-alder", Status: "supported", InstallRoute: "pkg:ed-alder", SessionSwitch: true, PluginFolder: "plugins/alder", FailureChannel: "exit code 1"},
		{Editor: "ed-birch", Status: "unsupported", FailureChannel: "stderr line"},
		{Editor: "ed-cedar", Status: "preview", FailureChannel: "exit code 2"},
	}
}

// Clone returns a copy a caller may mutate.
func (t Table) Clone() Table {
	c := make(Table, len(t))
	copy(c, t)
	return c
}

// Result is one validator's answer: pass, or fail with one refusal per refused row.
type Result struct {
	Refusals []string
}

// Pass reports whether no row was refused.
func (r Result) Pass() bool { return len(r.Refusals) == 0 }
