package review

// This file chooses each review's plan before the agents run (ADR-0049). Scout Intent classifies the
// change; the class picks a plan inside the review classes of ADR-0014 (which agents, tools and sources
// run, and whether the review is doubled); and the floor keeps the plan from going below the class the
// configured rules choose. The mapping from kind to plan is a setting (ADR-0013), checked on load.

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/scout"
)

// classConfigured is the plan class that keeps the class the configured rules chose.
const classConfigured = "configured"

// planSetting is one kind's plan as proven.yaml states it. Class is a review class, or classConfigured.
// DoubledOn doubles the review whatever its class; without it the class's own doubling applies. There
// is no "doubled off": a plan never undoes a doubling the configuration asks for (ADR-0049 point 3).
type planSetting struct {
	Class     string
	Agents    []string
	Tools     []string
	Sources   []string
	DoubledOn bool
}

// defaultPlans are ADR-0049 point 2's plans: a dependency-only change runs the dependency evidence
// without the Prover (or the Fixer, which needs a test the Prover wrote), a security-sensitive change is
// doubled, and a documentation-only change runs no model, no tool and no source. Every other kind runs
// in full in the class its rules chose.
func defaultPlans() map[string]planSetting {
	full := planSetting{Class: classConfigured, Agents: pipeline.PlanAgents, Tools: pipeline.PlanTools, Sources: pipeline.PlanSources}
	dep := full
	dep.Agents = []string{pipeline.AgentPlanner, pipeline.AgentInvestigator, pipeline.AgentSkeptic}
	sec := full
	sec.Class = "security"
	sec.DoubledOn = true
	return map[string]planSetting{
		scout.KindDependency: dep,
		scout.KindRefactor:   full,
		scout.KindFeature:    full,
		scout.KindSecurity:   sec,
		scout.KindDocs:       {Class: "docs", Agents: []string{}, Tools: []string{}, Sources: []string{}},
		scout.KindConfig:     full,
	}
}

// classSecurity is the review class whose execution obligations no plan may cut (ADR-0049 point 3).
const classSecurity = "security"

// cutsWork reports whether a plan leaves out any agent, tool or source the full plan runs.
func cutsWork(ps planSetting) bool {
	for _, side := range []struct{ have, all []string }{
		{ps.Agents, pipeline.PlanAgents}, {ps.Tools, pipeline.PlanTools}, {ps.Sources, pipeline.PlanSources},
	} {
		for _, name := range side.all {
			if !containsName(side.have, name) {
				return true
			}
		}
	}
	return false
}

// planKind is the kind a plan_ key sets: plan_docs is change.docs.
func planKind(key string) string { return "change." + strings.TrimPrefix(key, "plan_") }

// allowedPlan is the value list a refused plan_ key names.
func allowedPlan() string {
	return "class=<" + classConfigured + " or " + strings.Join(classNames(), " or ") + ">, agents=<" + strings.Join(pipeline.PlanAgents, "+") +
		", all or none>, tools=<" + strings.Join(pipeline.PlanTools, "+") + ", all or none>, sources=<" +
		strings.Join(pipeline.PlanSources, "+") + ", all or none>, doubled=<class or on>"
}

// parsePlanSetting reads "class=docs, agents=none, tools=none, sources=none, doubled=class". All five
// fields are required: a key that is set replaces the whole setting, so the file says the whole plan.
func parsePlanSetting(key, val string) (planSetting, error) {
	refuse := func(reason string) (planSetting, error) {
		return planSetting{}, configRefused(key, val, reason, allowedPlan())
	}
	var ps planSetting
	seen := map[string]bool{}
	for _, part := range strings.Split(val, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if !ok || k == "" || v == "" {
			return refuse("each field is name=value")
		}
		if seen[k] {
			return refuse("field " + k + " is named twice")
		}
		seen[k] = true
		var err error
		switch k {
		case "class":
			if v != classConfigured && !knownClass(v) {
				return refuse("unknown class " + v)
			}
			ps.Class = v
		case "agents":
			ps.Agents, err = planList(v, pipeline.PlanAgents)
		case "tools":
			ps.Tools, err = planList(v, pipeline.PlanTools)
		case "sources":
			ps.Sources, err = planList(v, pipeline.PlanSources)
		case "doubled":
			// ADR-0049 point 3: "off" is not offered, because a plan may only add a second review.
			switch v {
			case "class":
			case "on":
				ps.DoubledOn = true
			default:
				return refuse("doubled is class or on")
			}
		default:
			return refuse("unknown field " + k)
		}
		if err != nil {
			return refuse(err.Error())
		}
	}
	for _, f := range []string{"class", "agents", "tools", "sources", "doubled"} {
		if !seen[f] {
			return refuse("field " + f + " is missing")
		}
	}
	return ps, nil
}

// planList reads a "+"-joined list of names from allowed, or "all", or "none".
func planList(v string, allowed []string) ([]string, error) {
	switch v {
	case "all":
		return append([]string(nil), allowed...), nil
	case "none":
		return []string{}, nil
	}
	out := []string{}
	for _, name := range strings.Split(v, "+") {
		name = strings.TrimSpace(name)
		if !containsName(allowed, name) {
			return nil, fmt.Errorf("unknown name %q", name)
		}
		if containsName(out, name) {
			return nil, fmt.Errorf("%q is named twice", name)
		}
		out = append(out, name)
	}
	return out, nil
}

// floorClass is the class no plan may go below (ADR-0049 point 3), and the rule that set it. Today it is
// the class the configured path rules choose (ADR-0014). The organisation policy of ADR-0036 is not in
// this build; when it lands, its minimum class is passed as policy and the higher of the two is the floor.
func floorClass(files []string, cfg fileSettings, policy string) (string, string) {
	class, rule := chooseClass(files, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
	if policy != "" && knownClass(policy) && classRank(cfg.ClassOrder, policy) < classRank(cfg.ClassOrder, class) {
		return policy, "policy"
	}
	return class, rule
}

// reviewPlan is what choosePlan decided: the class the review states, the rule that chose it, whether it
// is doubled, and the pipeline's plan with its trace line.
type reviewPlan struct {
	Class   string
	Rule    string
	Doubled bool
	Plan    *pipeline.Plan
}

// offPlan is the plan of a review with the switch off: everything runs, and the trace says by name that
// no plan was chosen, so a reader of the trace never takes a full review for a planned one.
func offPlan() *pipeline.Plan {
	return &pipeline.Plan{Full: true, Note: "scout-intent: no plan chosen (scout_intent is off); the full plan runs"}
}

// choosePlan classifies the change and chooses its plan. files are the paths the review classes read.
func choosePlan(cfg fileSettings, req Request, files []string) (reviewPlan, error) {
	floor, floorRule := floorClass(files, cfg, "")
	full := func(why string) reviewPlan {
		return reviewPlan{
			Class: floor, Rule: floorRule, Doubled: cfg.Doubled[floor],
			Plan: &pipeline.Plan{Full: true, Note: why + "; the full plan runs in class " + floor + " (" + floorRule + ")"},
		}
	}
	changed, err := changedStatus(req)
	if err != nil {
		return reviewPlan{}, err
	}
	reg, err := scout.ChangeRegistry(securityPatterns(cfg))
	if err != nil {
		return reviewPlan{}, err
	}
	eng, err := scout.New(reg, scout.Options{})
	if err != nil {
		return reviewPlan{}, err
	}
	d, err := eng.Decide(scout.Input{Prompt: scout.FormatChange(changed), User: scout.User{ID: "proven", Role: "review"}, Surface: "change"})
	if err != nil {
		return reviewPlan{}, err
	}
	head := classified(d)
	kind := d.Intent
	switch {
	case hasCandidate(d, scout.KindSecurity):
		// A security signal is never outvoted: raising a review is always safe and lowering one never is,
		// so a change with one sensitive file is planned as security-sensitive however many other files it has.
		kind = scout.KindSecurity
		if d.Intent != scout.KindSecurity {
			head += "; a security signal outranks the top class " + d.Intent
		}
	case d.Ambiguity.Ambiguous || d.Clarification != nil || d.Band != scout.BandHigh:
		// ADR-0049: an ambiguous change is never planned as a confident class.
		return full(head + "; ambiguous, so no plan is chosen"), nil
	}
	// ADR-0049: the documentation plan runs no model, so it is chosen only for files known to be prose.
	// Scout reads names, and a name proves nothing: docs/run.go, a script under docs/ or an executable
	// file with a prose-like suffix would be reviewed by no model at all. A change that is not verified
	// prose runs the configured class in full.
	if kind == scout.KindDocs {
		if why := notProse(cfg, req, changed); why != "" {
			return full(head + "; " + why + ", so it is not documentation-only"), nil
		}
	}
	ps, ok := cfg.Plans[kind]
	if !ok {
		return full(head + "; no plan is set for " + kind), nil
	}
	class := ps.Class
	if class == classConfigured {
		class = floor
	}
	// ADR-0049 point 3: the configured class is the minimum. A plan whose class ranks below it is not
	// used at all, because its cuts were chosen for a lower class than the rules give this change.
	if classRank(cfg.ClassOrder, class) > classRank(cfg.ClassOrder, floor) {
		return full(head + "; the configured class " + floor + " outranks the plan's class " + class), nil
	}
	// ADR-0049 point 3: the floor keeps its execution obligations, not only its rank. A security-floored
	// change runs every agent, tool and source whatever plan the kind proposes: a class label that stays
	// "security" while the Prover is cut is a weaker review under a name that claims the stronger one.
	if floor == classSecurity && cutsWork(ps) {
		return full(head + "; the configured class " + floor + " keeps every agent, tool and source, so the plan's cuts are not used"), nil
	}
	rule := floorRule
	if class != floor {
		rule = "scout-intent: " + kind
	}
	rp := reviewPlan{
		Class: class, Rule: rule, Doubled: cfg.Doubled[class] || ps.DoubledOn,
		Plan: &pipeline.Plan{Agents: set(ps.Agents), Tools: set(ps.Tools), Sources: set(ps.Sources)},
	}
	rp.Plan.Note = head + "; plan: class " + class + " (" + rule + "), agents " + listOr(ps.Agents) + ", tools " + listOr(ps.Tools) +
		", sources " + listOr(ps.Sources) + ", doubled " + onOff(rp.Doubled)
	return rp, nil
}

// allowedProseSuffixes is the value list a refused prose_suffixes key names.
const allowedProseSuffixes = "a comma-separated list of file suffixes, each starting with a dot, such as .md, .rst, .txt"

// builtSuffixes is the closed deny list of suffixes that name source, scripts, build files or templates:
// every language Proven's adapters read, the common languages beside them, shell and Windows scripts,
// Makefile-like and build-tool files, and template languages. None of them can be listed as prose.
var builtSuffixes = set([]string{
	// Go, Python, JavaScript and TypeScript, Java and JVM, Rust, C family, C#, Ruby, PHP.
	".go", ".py", ".pyi", ".pyw", ".js", ".mjs", ".cjs", ".jsx", ".ts", ".mts", ".cts", ".tsx",
	".java", ".kt", ".kts", ".scala", ".groovy", ".clj", ".rs", ".c", ".h", ".cc", ".cpp", ".cxx", ".hpp",
	".hh", ".cs", ".fs", ".vb", ".rb", ".php", ".swift", ".m", ".mm", ".dart", ".lua", ".pl", ".pm", ".r",
	".ex", ".exs", ".erl", ".hs", ".ml", ".zig", ".nim", ".vue", ".svelte", ".sql", ".proto",
	// Shell and other scripts.
	".sh", ".bash", ".zsh", ".fish", ".ksh", ".csh", ".ps1", ".psm1", ".bat", ".cmd", ".awk", ".sed", ".tcl",
	// Makefile-like and build files.
	".mk", ".mak", ".make", ".cmake", ".gradle", ".bazel", ".bzl", ".just", ".ninja", ".rake", ".gemspec",
	".mod", ".sum", ".lock", ".toml", ".cfg", ".ini", ".yml", ".yaml", ".json", ".xml",
	// Templates and markup that runs code.
	".tmpl", ".tpl", ".gotmpl", ".j2", ".jinja", ".jinja2", ".erb", ".ejs", ".hbs", ".mustache", ".liquid",
	".haml", ".pug", ".html", ".htm", ".xhtml", ".svg", ".jsp", ".asp", ".aspx",
})

// parseProseSuffixes reads the prose_suffixes setting: suffixes are lower-cased, each is a dot and then
// letters, digits, "+", "-" or "_", and each is named once. A suffix with a second dot is refused, because
// the suffix of a file is what follows its last dot.
func parseProseSuffixes(val string) ([]string, error) {
	parts, err := splitCSV("prose_suffixes", val)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		low := strings.ToLower(p)
		ok := len(low) > 1 && low[0] == '.'
		for _, r := range low[1:] {
			if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '+' || r == '-' || r == '_') {
				ok = false
			}
		}
		if !ok {
			return nil, configRefused("prose_suffixes", p, "not a file suffix", allowedProseSuffixes)
		}
		// ADR-0049: the list says what is inert, and a source, script, build or template suffix is never
		// inert. Accepting one would let a file the model must read skip every model call.
		if builtSuffixes[low] {
			return nil, configRefused("prose_suffixes", p, "a source, script, build or template suffix is never prose", allowedProseSuffixes)
		}
		if containsName(out, low) {
			return nil, configRefused("prose_suffixes", p, "named twice", allowedProseSuffixes)
		}
		out = append(out, low)
	}
	return out, nil
}

// notProse returns why the change is not verified prose, or "" when every changed file (and, for a rename,
// its old path) has a suffix on the prose list, carries no executable bit and starts with no shebang.
// What cannot be verified is not prose: the answer is built from the facts, never from the file name alone.
func notProse(cfg fileSettings, req Request, changed []scout.ChangedFile) string {
	if len(changed) == 0 {
		return "no changed file is named"
	}
	var paths []string
	for _, c := range changed {
		paths = append(paths, c.Path)
		if c.From != "" {
			paths = append(paths, c.From)
		}
	}
	var facts map[string]change.Facts
	if req.BaselineCommit != "" {
		var err error
		if facts, err = change.FactsBetween(req.Target, req.BaselineCommit, req.HeadCommit); err != nil {
			return "the files' modes could not be read (" + err.Error() + ")"
		}
	} else {
		facts = change.TreeFacts(req.Target, paths)
	}
	for _, p := range paths {
		ext := strings.ToLower(path.Ext(p))
		// The loader refuses a built suffix, but settings can also be built in code; the guard holds either way.
		if builtSuffixes[ext] || !containsName(cfg.ProseSuffixes, ext) {
			return p + " has no prose suffix"
		}
		f, ok := facts[p]
		switch {
		case !ok:
			return p + " has no verified mode"
		case f.Unverified != "":
			return p + " is unverified (" + f.Unverified + ")"
		case f.Executable:
			return p + " is executable"
		case f.Shebang:
			return p + " starts with a shebang line"
		}
	}
	return ""
}

// classified is the start of a plan's trace line: the class Scout chose, its confidence and band, the
// signals that led there, and the ambiguity Scout reported.
func classified(d scout.Decision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "scout-intent: class %s (confidence %.3f, band %s); signals: ", d.Intent, d.Confidence, d.Band)
	if len(d.Signals) == 0 {
		b.WriteString("none")
	}
	for i, s := range d.Signals {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(s.Input + " matched " + s.Match)
	}
	b.WriteString("; ambiguity: ")
	if !d.Ambiguity.Ambiguous && len(d.Ambiguity.Reasons) == 0 {
		b.WriteString("none")
	} else {
		b.WriteString(strings.Join(d.Ambiguity.Reasons, ", "))
	}
	if len(d.Unexplained) > 0 {
		b.WriteString("; unmatched: " + strings.Join(d.Unexplained, ", "))
	}
	return b.String()
}

// changedStatus lists the changed files with how each changed. Between two commits git says it; for the
// prepared base/ and change/ trees the two sides are compared, and a file the same on both is left out.
func changedStatus(req Request) ([]scout.ChangedFile, error) {
	var out []scout.ChangedFile
	if req.BaselineCommit != "" {
		sts, err := change.StatusBetween(req.Target, req.BaselineCommit, req.HeadCommit)
		if err != nil {
			return nil, err
		}
		for _, s := range sts {
			out = append(out, scout.ChangedFile{Status: s.Status, Path: s.Path, From: s.From})
		}
		return out, nil
	}
	files, err := change.Read(req.Target)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		switch {
		case f.Base == nil:
			out = append(out, scout.ChangedFile{Status: "A", Path: f.Path})
		case f.Change == nil:
			out = append(out, scout.ChangedFile{Status: "D", Path: f.Path})
		case !bytes.Equal(f.Base, f.Change):
			out = append(out, scout.ChangedFile{Status: "M", Path: f.Path})
		}
	}
	return out, nil
}

// securityPatterns are the patterns the configuration files under the security class: Scout reads them
// as security signals, so the classifier never calls a file ordinary that the rules call sensitive.
func securityPatterns(cfg fileSettings) []string {
	var out []string
	for _, r := range cfg.PathRules {
		if r.Class == "security" {
			out = append(out, r.Pattern)
		}
	}
	return out
}

// hasCandidate reports whether Scout gave kind any weight.
func hasCandidate(d scout.Decision, kind string) bool {
	for _, c := range d.Candidates {
		if c.Intent == kind && c.Score > 0 {
			return true
		}
	}
	return false
}

func set(names []string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

func listOr(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, "+")
}

func onOff(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func containsName(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
