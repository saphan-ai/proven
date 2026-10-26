package custody

import (
	"bufio"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/generator"
)

// AutoModel is the model field of a request that names no model: the router
// picks one. There is nothing to compare an answer's routed model with, so
// such an answer is counted apart from the ones whose model is checked.
const AutoModel = "meterless/auto"

// maxLine bounds one line of a receipts file or a trace.
const maxLine = 4 << 20

// Answer is one model answer a review recorded in its trace.
type Answer struct {
	Time       string
	Agent      string
	Requested  string
	Routed     string
	ReceiptID  string
	ResponseID string
	// Source names where the answer was read: the trace file and its line.
	Source string
}

// Failure is one answer that is not in custody, with the reason.
type Failure struct {
	Answer Answer
	Reason string
}

// Day is the numbers of one UTC day.
type Day struct {
	Date string
	// Receipts is how many lines of the receipts file were written that day;
	// SignaturesVerified is how many of them verify under the pinned key.
	Receipts           int
	SignaturesVerified int
	// Answers is how many live model answers the traces recorded that day.
	Answers int
	// Matched is the answers whose response id equals the request id of a
	// receipt that verifies.
	Matched int
	// ModelChecked is the matched answers that asked for a specific model,
	// and ModelMatched the ones whose routed model is that model. AutoRouted
	// is the matched answers that asked for AutoModel; they are in neither.
	ModelChecked int
	ModelMatched int
	AutoRouted   int
	// InCustody is the answers that pass every check.
	InCustody int
	Failures  []Failure
}

// Report is the result of one audit.
type Report struct {
	Days []Day
	// Refused names each receipts-file line that was refused, with the cause.
	Refused []string
	// Skipped is how many trace model events were not live answers (a sample
	// run's calls, which name no request and no receipt).
	Skipped int
}

// Failed is the number of answers that are not in custody, over all days.
func (r Report) Failed() int {
	n := 0
	for _, d := range r.Days {
		n += len(d.Failures)
	}
	return n
}

// refusal is a receipts-file line that did not verify. Its request id is read
// without trust, only so an answer that names it can say why it has no receipt.
type refusal struct {
	err error
}

// Audit reads a receipts file and one or more traces and decides, for every
// live answer in the traces, whether it is in custody. An answer is in custody
// when (1) a receipt whose request id equals the answer's response id exists
// and its signature verifies under the pinned key, and (2) the model that
// answered is the model requested (or the request named no model: AutoModel).
func Audit(receipts io.Reader, key ed25519.PublicKey, traces map[string]io.Reader) (Report, error) {
	var rep Report
	days := map[string]*Day{}
	day := func(d string) *Day {
		if days[d] == nil {
			days[d] = &Day{Date: d}
		}
		return days[d]
	}

	good := map[string]Receipt{}
	bad := map[string]refusal{}
	sc := newScanner(receipts)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		id, ms := peek([]byte(line))
		d := day(dayOf(ms))
		d.Receipts++
		r, err := Verify([]byte(line), key)
		if err != nil {
			rep.Refused = append(rep.Refused, fmt.Sprintf("receipts line %d: %v", n, err))
			if id != "" {
				if _, seen := bad[id]; !seen {
					bad[id] = refusal{err: err}
				}
			}
			continue
		}
		d.SignaturesVerified++
		if _, seen := good[r.RequestID]; !seen {
			good[r.RequestID] = r
		}
	}
	if err := sc.Err(); err != nil {
		return Report{}, fmt.Errorf("reading receipts: %w", err)
	}

	// counted maps a response id to the source of the answer it vouched for.
	counted := map[string]string{}
	names := make([]string, 0, len(traces))
	for name := range traces {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		answers, skipped, err := readAnswers(name, traces[name])
		if err != nil {
			return Report{}, err
		}
		rep.Skipped += skipped
		for _, a := range answers {
			d := day(dayOf(parseTime(a.Time)))
			d.Answers++
			judge(d, a, good, bad, counted)
		}
	}

	for _, d := range days {
		rep.Days = append(rep.Days, *d)
	}
	sort.Slice(rep.Days, func(i, j int) bool { return rep.Days[i].Date < rep.Days[j].Date })
	return rep, nil
}

// judge applies the three checks to one answer and records the outcome on the
// day. The checks run in the order a reader would ask them; the first that
// fails names the reason, so one answer is one failure.
func judge(d *Day, a Answer, good map[string]Receipt, bad map[string]refusal, counted map[string]string) {
	fail := func(reason string) { d.Failures = append(d.Failures, Failure{Answer: a, Reason: reason}) }
	if a.ReceiptID == generator.NoReceipt {
		fail("no receipt: the answer came from an API that signs nothing")
		return
	}
	if a.ResponseID == "" {
		fail("no response id recorded, so no receipt can be matched")
		return
	}
	r, ok := good[a.ResponseID]
	if !ok {
		if b, refused := bad[a.ResponseID]; refused {
			fail("receipt refused: " + b.err.Error())
			return
		}
		fail("no receipt matches response id " + a.ResponseID)
		return
	}
	// A receipt that verifies but says the request failed vouches for nothing.
	if r.Status != "ok" {
		fail(fmt.Sprintf("receipt status is %q, not \"ok\"", r.Status))
		return
	}
	// One receipt vouches for one answer: a second answer on the same
	// response id is not a second answer in custody.
	if first, dup := counted[a.ResponseID]; dup {
		fail("response id already matched at " + first)
		return
	}
	counted[a.ResponseID] = a.Source
	d.Matched++
	routed := a.Routed
	if routed == "" || routed == "unknown" {
		fail("routed model unknown")
		return
	}
	// The signed receipt names the winning route. When it names a model, the
	// trace must agree with it: the receipt is the one signed statement.
	if r.Model != "" && r.Model != routed {
		fail(fmt.Sprintf("model mismatch: the stream names %s, the signed receipt names %s", routed, r.Model))
		return
	}
	switch {
	case a.Requested == AutoModel:
		d.AutoRouted++
	case a.Requested == "":
		fail("requested model not recorded, so the routed model cannot be compared")
		return
	case a.Requested != routed:
		d.ModelChecked++
		fail(fmt.Sprintf("model mismatch: requested %s, answered by %s", a.Requested, routed))
		return
	default:
		d.ModelChecked++
		d.ModelMatched++
	}
	d.InCustody++
}

// traceLine is the part of a trace event the audit reads.
type traceLine struct {
	Time  string `json:"time"`
	Kind  string `json:"kind"`
	Agent string `json:"agent"`
	Model *struct {
		Model      string `json:"model"`
		ReceiptID  string `json:"receipt_id"`
		Requested  string `json:"requested"`
		ResponseID string `json:"response_id"`
	} `json:"model"`
}

// readAnswers returns the live answers of one trace. A model event is a live
// answer when it names the model it requested or a receipt id; a sample run's
// events name neither and are counted as skipped. A line that is not JSON
// makes the whole trace unreadable: an audit that skipped it would count less
// than the review did.
func readAnswers(name string, r io.Reader) ([]Answer, int, error) {
	var out []Answer
	skipped := 0
	sc := newScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e traceLine
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, 0, fmt.Errorf("%s line %d: not JSON", name, n)
		}
		if e.Kind != "model" || e.Model == nil {
			continue
		}
		if e.Model.Requested == "" && e.Model.ReceiptID == "" {
			skipped++
			continue
		}
		out = append(out, Answer{
			Time: e.Time, Agent: e.Agent, Requested: e.Model.Requested, Routed: e.Model.Model,
			ReceiptID: e.Model.ReceiptID, ResponseID: e.Model.ResponseID,
			Source: fmt.Sprintf("%s:%d", name, n),
		})
	}
	if err := sc.Err(); err != nil {
		return nil, 0, fmt.Errorf("reading %s: %w", name, err)
	}
	return out, skipped, nil
}

// peek reads, without any trust, the request id and the time of a receipts
// line. It is used only to file a refused line under its day and to name it
// when an answer points at it.
func peek(line []byte) (string, int64) {
	var env envelope
	if json.Unmarshal(line, &env) != nil {
		return "", 0
	}
	var b body
	if json.Unmarshal(env.Body, &b) != nil || b.Receipt == nil {
		return "", b.TSMS
	}
	ms := b.TSMS
	if ms == 0 {
		ms = b.Receipt.CreatedAt
	}
	return b.Receipt.RequestID, ms
}

// dayOf is the UTC date of a time in milliseconds, or "unknown" for zero.
func dayOf(ms int64) string {
	if ms <= 0 {
		return "unknown"
	}
	return time.UnixMilli(ms).UTC().Format("2006-01-02")
}

// parseTime reads a trace time (RFC 3339) as milliseconds; 0 when it does not parse.
func parseTime(s string) int64 {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func newScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), maxLine)
	return sc
}

// AuditFiles is Audit over files on disk: one receipts file, one pinned key
// (base64) and the trace files. A file that cannot be opened is an error.
func AuditFiles(receiptsPath, keyB64 string, tracePaths []string) (Report, error) {
	key, err := PublicKey(strings.TrimSpace(keyB64))
	if err != nil {
		return Report{}, err
	}
	rf, err := os.Open(receiptsPath)
	if err != nil {
		return Report{}, err
	}
	defer rf.Close()
	traces := map[string]io.Reader{}
	for _, p := range tracePaths {
		f, err := os.Open(p)
		if err != nil {
			return Report{}, err
		}
		defer f.Close()
		traces[p] = f
	}
	if len(traces) == 0 {
		return Report{}, errors.New("no trace file given")
	}
	return Audit(rf, key, traces)
}

// Write prints the report: per day, the five numbers and then every answer
// that fails custody with its reason, then the totals.
func Write(w io.Writer, rep Report) {
	var all Day
	for _, d := range rep.Days {
		fmt.Fprintf(w, "day %s\n", d.Date)
		writeDay(w, d)
		all.Receipts += d.Receipts
		all.SignaturesVerified += d.SignaturesVerified
		all.Answers += d.Answers
		all.Matched += d.Matched
		all.ModelChecked += d.ModelChecked
		all.ModelMatched += d.ModelMatched
		all.AutoRouted += d.AutoRouted
		all.InCustody += d.InCustody
	}
	fmt.Fprintln(w, "all days")
	writeNumbers(w, all)
	for _, line := range rep.Refused {
		fmt.Fprintf(w, "refused %s\n", line)
	}
	fmt.Fprintf(w, "trace events skipped, not live answers: %d\n", rep.Skipped)
	if n := rep.Failed(); n > 0 {
		fmt.Fprintf(w, "custody FAILED for %d answers\n", n)
		return
	}
	if all.Answers == 0 {
		// With no live answer there is nothing custody could fail on, so
		// "holds" would read the same whether it holds or not.
		fmt.Fprintln(w, "custody UNMEASURED: the traces hold no live answer")
		return
	}
	fmt.Fprintln(w, "custody holds for every answer")
}

func writeDay(w io.Writer, d Day) {
	writeNumbers(w, d)
	for _, f := range d.Failures {
		fmt.Fprintf(w, "  FAIL %s %s agent %s response %s: %s\n",
			f.Answer.Source, f.Answer.Time, f.Answer.Agent, f.Answer.ResponseID, f.Reason)
	}
}

func writeNumbers(w io.Writer, d Day) {
	fmt.Fprintf(w, "  receipts: %d\n", d.Receipts)
	fmt.Fprintf(w, "  signatures verified: %d\n", d.SignaturesVerified)
	fmt.Fprintf(w, "  answers: %d\n", d.Answers)
	fmt.Fprintf(w, "  answers matched to a receipt by request id: %d\n", d.Matched)
	fmt.Fprintf(w, "  answers whose routed model equals the requested model: %d of %d checked (auto-routed, no model asked for: %d)\n",
		d.ModelMatched, d.ModelChecked, d.AutoRouted)
	fmt.Fprintf(w, "  answers in custody: %d\n", d.InCustody)
}
