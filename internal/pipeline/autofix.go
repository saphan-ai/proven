package pipeline

// This file carries out ADR-0050 in the pipeline: after every finding of the review is final, the
// one whole-change fix attempt is made, and the instruments that judge it are built here. The data
// judge runs the dependency check again on a patched copy of the change, through the same store seam
// the review used; the tool judge runs the finding's own tool in the Prover's container on a patched
// copy. Nothing here writes to the reviewed trees, and no model is asked about the result.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/agent/autofix"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/tool"
)

// autofix makes the one whole-change fix attempt for the findings of the review and writes what it
// did to the trace (ADR-0050 points 1 to 3). It returns nil when no attempt was made. A generator
// that is not connected or did not answer is named as a failed service by the caller, and the review
// goes on without a verdict: an attempt that was never made is not a rejected one.
func (p *Pipeline) autofix(ctx context.Context, st *state, req Request, root, snapshot string, gen generator.Generator, q query.Runner, run testrun.Adapter, cands []candidate.Candidate) (*candidate.Autofix, error) {
	if req.AutofixSpecial {
		// ADR-0050 point 6: the generator that writes the patch is a setting. The second reviewer's
		// generator is the only other one a review holds.
		if p.special == nil {
			return nil, &service.Down{Service: "autofix generator", Reason: "no second reviewer is connected"}
		}
		gen = p.special
	}
	judges := autofix.Judges{
		Data: dataJudge{q: q, root: root, table: p.osvTable, snapshot: snapshot, maxBytes: p.writtenLimit()},
		Tool: toolJudge{box: containerOf(run), timeout: p.toolTimeout, changeDir: filepath.Join(root, "change"), maxBytes: p.writtenLimit()},
	}
	if runner, ok := run.(testrun.FixRunner); ok {
		judges.Test = runner
	}
	if runner, ok := run.(testrun.BreakageRunner); ok {
		judges.Breakage = runner
	}
	a, err := autofix.Attempt(ctx, cands, gen, judges, autofix.Options{
		BaseDir:         filepath.Join(root, "base"),
		ChangeDir:       filepath.Join(root, "change"),
		Timeout:         req.AutofixTimeout,
		MaxWrittenBytes: p.writtenLimit(),
	})
	if err != nil || a == nil {
		return nil, err
	}
	return a, st.autofixEvents(a)
}

// redactAutofix returns a copy of a with rd applied to every string a model, a test run or a tool wrote.
// The attempt is, with the findings, the last place that text reaches the result, so it is redacted
// with them (ADR-0005, ADR-0013 point 4). A nil attempt stays nil.
func redactAutofix(rd *redact.Redactor, a *candidate.Autofix) *candidate.Autofix {
	if a == nil {
		return nil
	}
	cp := *a
	cp.What, cp.Why, cp.Diff, cp.Refused = rd.Text(a.What), rd.Text(a.Why), rd.Text(a.Diff), rd.Text(a.Refused)
	cp.Edits = nil
	for _, e := range a.Edits {
		cp.Edits = append(cp.Edits, candidate.FixEdit{Path: rd.Text(e.Path), Old: rd.Text(e.Old), New: rd.Text(e.New)})
	}
	cp.Findings = nil
	for _, f := range a.Findings {
		nf := candidate.AutofixFinding{ID: rd.Text(f.ID), Closed: f.Closed}
		for _, p := range f.Proofs {
			nf.Proofs = append(nf.Proofs, redactProof(rd, p))
		}
		cp.Findings = append(cp.Findings, nf)
	}
	cp.Breakage = nil
	if a.Breakage != nil {
		b := redactProof(rd, *a.Breakage)
		cp.Breakage = &b
	}
	return &cp
}

// redactProof returns a copy of p with rd applied to every string a test run or a tool wrote.
func redactProof(rd *redact.Redactor, p candidate.AutofixProof) candidate.AutofixProof {
	np := candidate.AutofixProof{
		Kind: p.Kind, Closed: p.Closed,
		Reason: rd.Text(p.Reason), Command: rd.Text(p.Command), Output: rd.Text(p.Output),
	}
	for _, r := range p.Runs {
		np.Runs = append(np.Runs, candidate.FixRun{
			Label: r.Label, Command: rd.Text(r.Command), Result: r.Result, ExitCode: r.ExitCode, Output: rd.Text(r.Output),
		})
	}
	return np
}

// containerOf returns the Prover's container behind a test runner, or nil when the runner is not one:
// a sample review starts no container, so a tool cannot be run again in it.
func containerOf(run testrun.Adapter) *testrun.Container {
	if c, ok := run.(*testrun.Container); ok {
		return c
	}
	if w, ok := run.(checkoutRun); ok {
		return containerOf(w.inner)
	}
	return nil
}

// autofixEvents writes the trace lines of the attempt: the verdict, then one line for each finding
// and one for each of its proofs with the command that ran, so a reader of the trace can run it again.
func (s *state) autofixEvents(a *candidate.Autofix) error {
	text := "autofix verdict: " + a.Verdict
	if a.Refused != "" {
		text += ": " + a.Refused
	}
	if err := s.event("review", "autofix", "autofix", text); err != nil {
		return err
	}
	for _, f := range a.Findings {
		state := "open"
		if f.Closed {
			state = "closed"
		}
		if err := s.event(f.ID, "autofix", "autofix", "finding "+state+" after the patch"); err != nil {
			return err
		}
		for _, pr := range f.Proofs {
			line := fmt.Sprintf("%s proof %s", pr.Kind, map[bool]string{true: "closed", false: "open"}[pr.Closed])
			if pr.Command != "" {
				line += ": " + pr.Command
			}
			if pr.Reason != "" {
				line += ": " + pr.Reason
			}
			if err := s.event(f.ID, "autofix", "autofix", line); err != nil {
				return err
			}
		}
	}
	if b := a.Breakage; b != nil {
		line := fmt.Sprintf("breakage proof %s", map[bool]string{true: "closed", false: "open"}[b.Closed])
		for _, r := range b.Runs {
			line += fmt.Sprintf("; %s: %s", r.Command, r.Result)
		}
		if b.Reason != "" {
			line += ": " + b.Reason
		}
		return s.event("review", "autofix", "autofix", line)
	}
	return nil
}

// dataJudge runs a data-proved finding's own query again on the patched change (ADR-0050 point 2):
// the dependency check, which derives every query from the change's manifests, is run over a
// patched copy of the two trees, and the finding is still there when that check names it again or
// binds the same query request to the change. The store is the one the review used.
type dataJudge struct {
	q        query.Runner
	root     string
	table    string
	snapshot string
	maxBytes int
}

// Rerun implements autofix.DataRerun.
func (d dataJudge) Rerun(ctx context.Context, edits []testrun.Edit, c candidate.Candidate, qe *candidate.QueryEvidence) (autofix.Reading, error) {
	// The dependency check's own evidence records the query it ran and carries no template request;
	// a query a model named carries the request. Either way the line says what was asked.
	r := autofix.Reading{Command: requestText(qe.Request)}
	if qe.Request.Template == "" {
		r.Command = qe.SQL
	}
	if d.root == "" {
		r.Reason = "the review holds no trees to patch"
		return r, nil
	}
	patched, err := testrun.StagePatchedRoot(d.root, edits, d.maxBytes)
	if err != nil {
		return autofix.Reading{}, err
	}
	defer os.RemoveAll(patched)
	out, err := osv.Check(ctx, d.q, patched, d.table, d.snapshot)
	if err != nil {
		if ctx.Err() != nil {
			return autofix.Reading{}, ctx.Err()
		}
		r.Reason = "the dependency check could not run again: " + err.Error()
		return r, nil
	}
	if out.Down != nil {
		r.Reason = "the dependency check could not run again: " + out.Down.Error()
		return r, nil
	}
	r.Ran = true
	for _, n := range out.Candidates {
		if n.ID != c.ID {
			continue
		}
		r.Present = true
		for _, e := range n.Evidence {
			if e.Query != nil {
				r.Output = rowsText(e.Query.Columns, e.Query.Rows)
				return r, nil
			}
		}
	}
	if out.Facts.Binds(qe.Request) {
		r.Present = true
		r.Output = "the same request still binds to the patched change"
	}
	return r, nil
}

// requestText is a template request as one line: the template and its values in name order.
func requestText(req querytpl.Request) string {
	keys := make([]string, 0, len(req.Params))
	for k := range req.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := []string{req.Template}
	for _, k := range keys {
		parts = append(parts, k+"="+req.Params[k])
	}
	return strings.Join(parts, " ")
}

// rowsText is a query's rows as lines of text, the first line the column names.
func rowsText(columns []string, rows [][]string) string {
	lines := []string{strings.Join(columns, " | ")}
	for _, r := range rows {
		lines = append(lines, strings.Join(r, " | "))
	}
	return strings.Join(lines, "\n")
}

// toolJudge runs a tool-proved finding's own tool again, in the Prover's container, on a patched
// copy of the change (ADR-0050 points 2 and 5). The command is the adapter's, built by Proven.
type toolJudge struct {
	box       *testrun.Container
	timeout   time.Duration
	changeDir string
	maxBytes  int
}

// Rerun implements autofix.ToolRerun.
func (j toolJudge) Rerun(ctx context.Context, edits []testrun.Edit, c candidate.Candidate, t *candidate.ToolEvidence) (autofix.Reading, error) {
	adapter, ok := tool.ByID(t.Tool)
	if !ok {
		return autofix.Reading{Reason: "no command is known for the tool " + t.Tool}, nil
	}
	if j.box == nil || j.timeout <= 0 {
		return autofix.Reading{Reason: t.Tool + " cannot run again: no container, or no time limit for a tool"}, nil
	}
	patched, err := testrun.StagePatched(j.changeDir, edits, j.maxBytes)
	if err != nil {
		return autofix.Reading{}, err
	}
	defer os.RemoveAll(patched)
	got, err := tool.Rerun(ctx, adapter, tool.ContainerHost{Box: j.box, Timeout: j.timeout}, patched, []string{c.Path})
	if err != nil {
		return autofix.Reading{}, err
	}
	return autofix.Reading{Ran: got.Answered, Present: got.Present, Command: got.Command, Output: got.Output, Reason: got.Reason}, nil
}
