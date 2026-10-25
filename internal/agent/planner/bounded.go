package planner

import (
	"context"
	"fmt"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/markov"
)

// stepSpec is the JSON the Planner asks for at one bounded step (ADR-0048 point 1): the risks this part shows, and
// the state to carry to the next part.
var stepSpec = answer.Spec{
	Task: "Name the risks in one part of a larger change. You see this part and the state carried from the parts before it, never the whole change. " +
		"A claim the carried state lists as awaiting evidence that this part supplies is now a risk: name it. " +
		"Then return the state to carry on, short: what is established, what is suspected, and which claims await evidence from a part not yet read, each saying what evidence it needs.",
	Shape: "one JSON object. It has risks (a JSON array of objects, each with id (string), rule_id (string), title (string), message (string), severity (string), path (string), start_line (number) and end_line (number)), " +
		"and established, suspected and awaiting_evidence (each a JSON array of strings).",
	Example: `{"risks":[{"id":"F1","rule_id":"example-rule","title":"Example risk","message":"What is wrong.","severity":"ERROR","path":"file.go","start_line":1,"end_line":1}],` +
		`"established":["file.go validates its input"],"suspected":[],"awaiting_evidence":["handler.go may pass unchecked input to file.go; needs the caller"]}`,
}

// partRecord is one part as a step's prompt carries it: a whole file as the whole review carries it, or one hunk.
type partRecord struct {
	fileRecord
	Hunk *string `json:"hunk,omitempty"`
}

// stepRecord is the data of one step's prompt. The carried state is the spec's labelled block, as text.
type stepRecord struct {
	Step         int        `json:"step"`
	Steps        int        `json:"steps"`
	CarriedState string     `json:"carried_state"`
	Part         partRecord `json:"part"`
}

// StepPrompt is the Planner's prompt for one bounded step: this part and the carried state, and nothing else of the
// change. The sample generator keys on it, as on Prompt.
func StepPrompt(step, steps int, carried markov.Block, p markov.Part) string {
	rec := stepRecord{Step: step, Steps: steps, CarriedState: carried.Render()}
	if p.File != nil {
		rec.Part.Path, rec.Part.PathBase64 = part([]byte(p.File.Path))
		rec.Part.Base, rec.Part.BaseBase64 = part(p.File.Base)
		rec.Part.Change, rec.Part.ChangeBase64 = part(p.File.Change)
	} else {
		rec.Part.Path, rec.Part.PathBase64 = part([]byte(p.Path))
		h := p.Hunk
		rec.Part.Hunk = &h
	}
	return answer.Instruct(stepSpec, answer.Encode(rec))
}

// stepAnswer is one step's answer as the model returns it.
type stepAnswer struct {
	Risks       []row    `json:"risks"`
	Established []string `json:"established"`
	Suspected   []string `json:"suspected"`
	Awaiting    []string `json:"awaiting_evidence"`
}

// Step is what one bounded step did, for the trace (ADR-0048, the order's point 6).
type Step struct {
	Index, Steps int
	// Part names the file, or the file and the hunk header.
	Part string
	// PromptTokens is the step's prompt; CarriedTokens is the state it received, and Carried the number of items in it.
	PromptTokens  int
	CarriedTokens int
	Carried       int
	// Dropped is how many items the bound cut from the state this step handed on.
	Dropped int
	// Risks is how many risks the step named.
	Risks int
}

// Size is the change's size as the threshold reads it: the estimated tokens of the Planner's whole prompt.
func Size(files []change.File) int {
	return markov.EstimateTokens(Prompt(files))
}

// PlanBounded asks the generator for the risks one part at a time (ADR-0048 point 1). Each step's prompt holds one
// part and the state carried from the step before, bounded to s.Carry estimated tokens; no step sees the whole
// change. A risk id a later step repeats is renamed with its step, because the steps do not see each other's ids.
func PlanBounded(ctx context.Context, files []change.File, gen generator.Generator, s markov.Settings) ([]candidate.Risk, []Step, error) {
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("planner: the change has no files")
	}
	parts := markov.Split(files, s.Unit)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("planner: the change has no changed region")
	}
	var (
		state markov.ReviewState
		risks []candidate.Risk
		steps []Step
	)
	seen := map[string]bool{}
	// The first step's state is empty; it is fitted like every other, so the bound is applied in one place.
	carried, dropped := state.Block(1, len(parts)).Fit(s.Carry)
	for i, p := range parts {
		n := i + 1
		prompt := StepPrompt(n, len(parts), carried, p)
		rec := Step{
			Index: n, Steps: len(parts), Part: p.Name(), PromptTokens: markov.EstimateTokens(prompt),
			CarriedTokens: carried.Tokens(), Carried: len(carried.Decisions) + len(carried.Context) + len(carried.Pinned),
		}
		var got stepAnswer
		if err := answer.Read(ctx, gen, "planner", prompt, stepSpec, &got); err != nil {
			return nil, steps, err
		}
		for j, r := range got.Risks {
			// As in Plan, a risk with no id is refused, and the error names the position, never model text.
			if r.ID == "" {
				return nil, steps, fmt.Errorf("planner: step %d, risk %d has no id", n, j+1)
			}
			id := r.ID
			if seen[id] {
				id = fmt.Sprintf("%s-s%d", r.ID, n)
			}
			if seen[id] {
				return nil, steps, fmt.Errorf("planner: step %d, risk %d repeats an id", n, j+1)
			}
			seen[id] = true
			risks = append(risks, candidate.Risk{
				ID: id, RuleID: r.RuleID, Title: r.Title, Message: r.Message,
				Severity: r.Severity, Path: r.Path, StartLine: r.StartLine, EndLine: r.EndLine,
			})
		}
		rec.Risks = len(got.Risks)
		// Markovian: the model's answer replaces the state; nothing of this step but that state reaches the next.
		state = markov.ReviewState{Established: got.Established, Suspected: got.Suspected, Awaiting: got.Awaiting}
		if n < len(parts) {
			var cerr error
			carried, dropped, cerr = state.Carry(n, n+1, len(parts), s.Carry)
			rec.Dropped = dropped
			if cerr != nil {
				// ADR-0048: an open claim that does not fit is never dropped; the review stops and says so.
				return nil, append(steps, rec), cerr
			}
		}
		steps = append(steps, rec)
	}
	// ADR-0048: the last part has been read, so a claim still awaiting evidence can never be settled. Reporting the
	// risks alone as success would look complete when it is not; the review stops by name instead, as it does when
	// the carried state is full.
	if len(state.Awaiting) > 0 {
		return nil, steps, &markov.OpenClaimsError{Steps: len(parts), Open: len(state.Awaiting)}
	}
	return risks, steps, nil
}
