package report

// AnalysisStatement describes execution, independently of any reviewer label.
func (r Result) AnalysisStatement() string {
	switch r.AnalysisMode {
	case "recorded":
		return "Recorded demo: replayed answers and outcomes; no live model, safety service or project tests were run."
	case "deterministic":
		return "Deterministic analysis: no model was used."
	}
	return ""
}
