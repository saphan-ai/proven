package tool

import (
	"encoding/json"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Govulncheck is the Go reachability tool (ADR-0043).
// It runs govulncheck -json on the base and on the change, against a local
// vulnerability database. The container has no network, so the database is a
// file URL under the checkout mount and never the public vulnerability host.
// DB empty means there is no command: the runner names the tool, and every
// Go dependency finding stays "not measured".
type Govulncheck struct {
	DB string
}

// ID is the tool's stable name.
func (Govulncheck) ID() string { return "govulncheck" }

// Version is empty. A live run does not start a second container to ask,
// and the runner records unknown. A sample scenario carries its own version.
func (Govulncheck) Version() string { return "" }

// Applies reports whether the change is one govulncheck can read.
// A Go file, a go.mod or a go.sum can add a call or a dependency.
// An absolute path is not a path inside the checkout.
func (Govulncheck) Applies(paths []string) bool {
	for _, p := range paths {
		if goChange(p) {
			return true
		}
	}
	return false
}

// Argv is govulncheck -json with the local database, on the whole module.
// The pattern is ./... and not the changed files: a new dependency can be
// called from a file the diff did not touch (ADR-0043). There is no shell.
// A database that is not a file URL inside the checkout is refused, so the
// command cannot name a network host or a path the container cannot see.
func (g Govulncheck) Argv(paths []string) ([]string, string) {
	if strings.TrimSpace(g.DB) == "" {
		return nil, "no local vulnerability database"
	}
	if why := refuseDB(g.DB); why != "" {
		return nil, why
	}
	return []string{"govulncheck", "-json", "-db", g.DB, "./..."}, ""
}

// Read reports whether the output names a reached vulnerable function.
// The output is kept whole so a later step can label each dependency.
func (Govulncheck) Read(output string) Reading {
	reached := false
	for _, sym := range Symbols(output) {
		if sym.Label == candidate.ReachReached {
			reached = true
			break
		}
	}
	return Reading{Present: reached, Output: output}
}

// Symbol is one vulnerable function govulncheck reported for a module.
// Label is ReachReached or ReachImported. CallPath is set only for a reached function.
// The negative control mutates Label: an unreached function labeled reached turns its test red.
type Symbol struct {
	Module   string
	Function string
	Label    string
	CallPath string
}

// Symbols reads a govulncheck -json stream.
// The field names are the govulncheck protocol (golang.org/x/vuln v1.1.4,
// the version the module proxy named on 2026-10-24). The module zip was not
// opened: the checksum database was not available, so this comment is not a
// reading of those source files. One JSON object per line. A line that is
// not an object is skipped, because the container joins stderr into the output.
func Symbols(output string) []Symbol {
	var out []Symbol
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var msg struct {
			Finding *struct {
				Trace []vulnFrame `json:"trace"`
			} `json:"finding"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil || msg.Finding == nil {
			continue
		}
		frames := msg.Finding.Trace
		if len(frames) == 0 || frames[0].Module == "" {
			continue
		}
		label, call := labelOf(frames)
		out = append(out, Symbol{
			Module: frames[0].Module, Function: frames[0].Function, Label: label, CallPath: call,
		})
	}
	return out
}

// ScanStarted reports whether output is a govulncheck -json scan that started.
// A scan that started prints a config object naming this scanner. An error
// line, including a missing database, is not a measurement: silence after a
// failed start is not "imported but not reached" (ADR-0043).
func ScanStarted(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}
		var msg struct {
			Config *struct {
				Scanner string `json:"scanner_name"`
			} `json:"config"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil || msg.Config == nil {
			continue
		}
		if msg.Config.Scanner == "govulncheck" {
			return true
		}
	}
	return false
}

// vulnFrame is one frame of a govulncheck trace.
// The first frame is the vulnerable symbol. Later frames are its callers.
type vulnFrame struct {
	Module   string `json:"module"`
	Package  string `json:"package"`
	Function string `json:"function"`
	Receiver string `json:"receiver"`
	Position *struct {
		Filename string `json:"filename"`
		Line     int    `json:"line"`
	} `json:"position"`
}

// labelOf classifies one trace (ADR-0043).
// A function with a caller is reached. Every other trace was measured and
// the function was not called, which includes a module that is required but
// whose package is not imported: the order names no fourth label.
// The unreached result is Symbol.Label. The negative control changes that
// result to reached and the test goes red.
func labelOf(frames []vulnFrame) (string, string) {
	if len(frames) > 1 && frames[0].Function != "" {
		return candidate.ReachReached, callPath(frames)
	}
	return candidate.ReachImported, ""
}

// callPath writes the trace from the caller down to the vulnerable function.
// -json has no sentence of its own. That order is the order the text report
// uses, so a reader sees the same path either way.
func callPath(frames []vulnFrame) string {
	var parts []string
	for i := len(frames) - 1; i >= 0; i-- {
		if name := frameName(frames[i]); name != "" {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, " calls ")
}

// frameName is one frame as a reader would say it.
func frameName(f vulnFrame) string {
	sym := f.Function
	if f.Receiver != "" && sym != "" {
		sym = f.Receiver + "." + sym
	}
	if f.Package != "" && sym != "" {
		sym = f.Package + "." + sym
	}
	if sym == "" {
		sym = f.Package
	}
	if sym == "" {
		sym = f.Module
	}
	if sym == "" {
		return ""
	}
	if f.Position != nil && f.Position.Filename != "" && f.Position.Line > 0 {
		return f.Position.Filename + ":" + strconv.Itoa(f.Position.Line) + ": " + sym
	}
	return sym
}

// LabelFindings sets the reachability of each dependency finding (ADR-0043).
// A non-Go ecosystem is not measured. A Go finding is not measured when the
// tool did not answer, or when it answered and never mentioned the module:
// silence is not "imported but not reached". A module with any reached
// function is reached, and the call paths are kept. Nothing is dropped.
func LabelFindings(cands []candidate.Candidate, answered bool, symbols []Symbol) []candidate.Candidate {
	if len(cands) == 0 {
		return cands
	}
	byMod := map[string][]Symbol{}
	for _, sym := range symbols {
		byMod[sym.Module] = append(byMod[sym.Module], sym)
	}
	out := make([]candidate.Candidate, len(cands))
	for i, c := range cands {
		if c.Ecosystem == "" {
			out[i] = c
			continue
		}
		label, call := candidate.ReachNotMeasured, ""
		if c.Ecosystem == "Go" && answered {
			label, call = matchModule(byMod[moduleOf(c)])
		}
		c.ReachLabel = label
		c.CallPath = call
		c.Message = appendReach(c.Message, label, call)
		out[i] = c
	}
	return out
}

// moduleOf is the package the dependency finding names.
// The check stores it after the "osv:" prefix of the finding id.
func moduleOf(c candidate.Candidate) string {
	return strings.TrimPrefix(c.ID, "osv:")
}

// matchModule picks the strongest label for one module.
// Reached wins over imported. Several call paths are joined in sorted order
// so the sentence does not depend on map order.
func matchModule(syms []Symbol) (string, string) {
	if len(syms) == 0 {
		return candidate.ReachNotMeasured, ""
	}
	var paths []string
	reached := false
	for _, sym := range syms {
		if sym.Label != candidate.ReachReached {
			continue
		}
		reached = true
		if sym.CallPath != "" {
			paths = append(paths, sym.CallPath)
		}
	}
	if !reached {
		return candidate.ReachImported, ""
	}
	sort.Strings(paths)
	return candidate.ReachReached, strings.Join(uniqueStrings(paths), "; ")
}

// uniqueStrings drops repeated paths. The input is sorted.
func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := []string{in[0]}
	for _, s := range in[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}

// appendReach adds the label, and the call path when the function is reached,
// to the finding the reader already has. The comparison sentences stay.
func appendReach(msg, label, call string) string {
	sentence := "Reachability: " + label + "."
	if label == candidate.ReachReached && call != "" {
		sentence = "Reachability: " + label + ". Call path: " + call + "."
	}
	if msg == "" {
		return sentence
	}
	return msg + " " + sentence
}

// goChange reports a relative Go file or a Go dependency file.
func goChange(p string) bool {
	if p == "" || path.IsAbs(p) {
		return false
	}
	base := path.Base(p)
	return strings.HasSuffix(p, ".go") || base == "go.mod" || base == "go.sum"
}

// refuseDB says why raw cannot be the database argument, or "" when it can.
// The container opens the URL itself. Only a file URL under the checkout
// mount is visible, and an http URL would need the network the container
// does not have (ADR-0025). A refused URL is not replaced with the public host.
func refuseDB(raw string) string {
	const why = "not a file URL inside the checkout"
	if strings.ContainsAny(raw, " \t\r\n") {
		return why
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "file" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" {
		return why
	}
	if strings.Contains(u.Path, "..") {
		return why
	}
	p := path.Clean(u.Path)
	mount := testrun.CheckoutDir
	if p != mount && !strings.HasPrefix(p, mount+"/") {
		return why
	}
	return ""
}
