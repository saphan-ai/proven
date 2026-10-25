package trace

import "github.com/saphan-ai/proven/internal/redact"

// redactEvent returns a copy of e with r applied to every field that can carry
// text from a service or from a repository under review. The caller's event and
// the slices it holds are not changed. The run id and the time are written by
// the writer itself and carry no such text.
func redactEvent(e Event, r *redact.Redactor) Event {
	if r == nil {
		return e
	}
	e.EventID = r.Text(e.EventID)
	e.FindingID = r.Text(e.FindingID)
	e.Kind = r.Text(e.Kind)
	e.Agent = r.Text(e.Agent)
	e.Text = r.Text(e.Text)
	e.ChangeKey = r.Text(e.ChangeKey)
	e.RuleID = r.Text(e.RuleID)
	e.Path = r.Text(e.Path)
	e.Title = r.Text(e.Title)
	e.Run = r.Text(e.Run)
	if e.Evidence != nil {
		ev := make([]Evidence, len(e.Evidence))
		for i, x := range e.Evidence {
			ev[i] = redactEvidence(x, r)
		}
		e.Evidence = ev
	}
	if e.Refutation != nil {
		ref := *e.Refutation
		ref.Reason = r.Text(ref.Reason)
		e.Refutation = &ref
	}
	if e.Guardrail != nil {
		g := *e.Guardrail
		g.Side = r.Text(g.Side)
		g.Verdict = r.Text(g.Verdict)
		g.Categories = r.Texts(g.Categories)
		// The model id comes from a setting, so it is redacted like any other
		// configured value; backend and down are fixed names, redacted for symmetry.
		g.Backend = r.Text(g.Backend)
		g.Model = r.Text(g.Model)
		g.Down = r.Text(g.Down)
		e.Guardrail = &g
	}
	if e.Model != nil {
		m := *e.Model
		m.Route = r.Text(m.Route)
		m.Model = r.Text(m.Model)
		m.ReceiptID = r.Text(m.ReceiptID)
		m.Requested = r.Text(m.Requested)
		m.ResponseID = r.Text(m.ResponseID)
		e.Model = &m
	}
	return e
}

// redactEvidence returns a copy of one piece of evidence with r applied to its text.
func redactEvidence(x Evidence, r *redact.Redactor) Evidence {
	x.Kind = r.Text(x.Kind)
	if x.Test != nil {
		t := *x.Test
		t.Path = r.Text(t.Path)
		t.Name = r.Text(t.Name)
		t.Command = r.Text(t.Command)
		t.Output = r.Text(t.Output)
		x.Test = &t
	}
	if x.Query != nil {
		q := *x.Query
		q.SQL = r.Text(q.SQL)
		q.Form = r.Text(q.Form)
		q.Engine = r.Text(q.Engine)
		q.Snapshot = r.Text(q.Snapshot)
		q.Columns = r.Texts(q.Columns)
		if q.Rows != nil {
			rows := make([][]string, len(q.Rows))
			for i, row := range q.Rows {
				rows[i] = r.Texts(row)
			}
			q.Rows = rows
		}
		x.Query = &q
	}
	if x.Tool != nil {
		// A tool's output is text from a checkout, so a secret in it is redacted
		// the same way as a test's output (ADR-0005).
		t := *x.Tool
		t.Tool = r.Text(t.Tool)
		t.Version = r.Text(t.Version)
		t.Command = r.Text(t.Command)
		t.Output = r.Text(t.Output)
		t.BaseOutput = r.Text(t.BaseOutput)
		x.Tool = &t
	}
	return x
}
