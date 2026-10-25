// Command guardcount counts guardrail checks in one or more trace files.
//
// It reads the JSONL trace a review writes (logs/trace.jsonl) and prints one
// row per UTC day, side, verdict and backend: how many checks there were, and
// their median latency. The verdict is "safe", "unsafe" or "down:<reason>",
// where the reason is the trace's structured down value. It reads only the
// guardrail lines; every other line, and the "no guardrail" line of a review
// run with the guardrail off, is skipped and adds no row. It sends nothing
// anywhere and changes no file.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/saphan-ai/proven/internal/trace"
)

const (
	// Exit codes, as the help text lists them.
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// helpText is printed by -h and --help, and after a refused argument.
const helpText = `guardcount counts guardrail checks in trace files.

Usage: guardcount <trace.jsonl> [<trace.jsonl> ...]

One row per UTC day, side, verdict and backend, sorted in that order:
  day  side  verdict  backend  count  median_ms
verdict is safe, unsafe or down:<reason>. A field an older trace line
does not carry prints as "-"; an older down line with no structured
check counts as down:unstructured.

Exit codes: 0 counted (even when there is nothing to count),
1 a file could not be read or a line could not be decoded, 2 usage.
`

// lineMax is the longest trace line read. A trace line carries evidence, so
// it can be long; a line past this bound is refused, not cut.
const lineMax = 16 << 20

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run counts the checks in the files named by args and prints the table.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, helpText)
		return exitUsage
	}
	for _, a := range args {
		if a == "-h" || a == "--help" {
			fmt.Fprint(stdout, helpText)
			return exitOK
		}
		// No flags are defined, so a dash-led argument is a mistake, not a file.
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(stderr, "guardcount: unknown flag %s\n\n%s", a, helpText)
			return exitUsage
		}
	}
	groups := map[key][]int64{}
	for _, path := range args {
		if err := countFile(path, groups); err != nil {
			fmt.Fprintf(stderr, "guardcount: %v\n", err)
			return exitError
		}
	}
	if err := write(stdout, groups); err != nil {
		fmt.Fprintf(stderr, "guardcount: %v\n", err)
		return exitError
	}
	return exitOK
}

// key is one row of the table.
type key struct {
	day, side, verdict, backend string
}

// countFile adds the guardrail checks of one trace file to groups.
func countFile(path string, groups map[key][]int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), lineMax)
	n := 0
	for sc.Scan() {
		n++
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var e trace.Event
		if err := json.Unmarshal(line, &e); err != nil {
			return fmt.Errorf("%s:%d: not a trace line: %v", path, n, err)
		}
		k, latency, ok, err := checkOf(e)
		if err != nil {
			return fmt.Errorf("%s:%d: %v", path, n, err)
		}
		if ok {
			groups[k] = append(groups[k], latency)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	return nil
}

// checkOf returns the row and latency of one guardrail check, and false for a
// line that is not a check.
func checkOf(e trace.Event) (key, int64, bool, error) {
	if e.Kind != "guardrail" {
		return key{}, 0, false, nil
	}
	g := e.Guardrail
	// A down line written before the trace had a structured check carries
	// only its text. It is still a check that could not be made, so it is
	// counted, under a value no current line uses. "no guardrail" is not a
	// check and has no "guardrail: " prefix, so it is skipped here.
	if g == nil {
		if !strings.HasPrefix(e.Text, "guardrail: ") {
			return key{}, 0, false, nil
		}
		g = &trace.Check{Down: "unstructured"}
	}
	t, err := time.Parse(time.RFC3339, e.Time)
	if err != nil {
		return key{}, 0, false, errors.New("guardrail line has no RFC 3339 time")
	}
	verdict := g.Verdict
	if g.Down != "" {
		verdict = "down:" + g.Down
	}
	k := key{
		// The day is UTC, so traces written on hosts in different zones count
		// into the same day.
		day:     t.UTC().Format(time.DateOnly),
		side:    orDash(g.Side),
		verdict: orDash(verdict),
		backend: orDash(g.Backend),
	}
	return k, g.LatencyMS, true, nil
}

// orDash prints a field an older line does not carry as "-", so every row
// has the same number of columns.
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// write prints the header and one row per key, sorted, so the same traces
// always give the same bytes.
func write(w io.Writer, groups map[key][]int64) error {
	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.day != b.day {
			return a.day < b.day
		}
		if a.side != b.side {
			return a.side < b.side
		}
		if a.verdict != b.verdict {
			return a.verdict < b.verdict
		}
		return a.backend < b.backend
	})
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "day\tside\tverdict\tbackend\tcount\tmedian_ms")
	for _, k := range keys {
		lat := groups[k]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%d\n", k.day, k.side, k.verdict, k.backend, len(lat), median(lat))
	}
	return tw.Flush()
}

// median is the middle latency; for an even count, the mean of the two middle
// values, rounded down to a whole millisecond.
func median(values []int64) int64 {
	s := append([]int64(nil), values...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	mid := len(s) / 2
	if len(s)%2 == 1 {
		return s[mid]
	}
	return (s[mid-1] + s[mid]) / 2
}
