package markov

// Chunk is the spec's per-step record, ChunkInfo (AGENTS.md section 2.2).
type Chunk struct {
	// ID is the 1-based step index.
	ID int `json:"id"`
	// Tokens is the estimated size of the step's cleaned output.
	Tokens int `json:"tokens"`
	// Carryover is the state passed INTO this step.
	Carryover string `json:"carryover"`
	// Content is the step's display-safe output.
	Content string `json:"content"`
}

// Efficiency is one run's cost against the append-history baseline (AGENTS.md section 6.3).
type Efficiency struct {
	TotalTokensUsed   int     `json:"totalTokensUsed"`
	StandardCost      int     `json:"standardCost"`
	MarkovianCost     int     `json:"markovianCost"`
	TokensSaved       int     `json:"tokensSaved"`
	EfficiencyPercent float64 `json:"efficiencyPercent"`
}

// RunEfficiency computes a run's efficiency against the configuration it is given, which must be the run's own
// snapshot, never the current one (AGENTS.md sections 2.3 and 14). The standard cost of step i is every earlier
// step's tokens plus its own; the Markovian cost is one carried state per step plus the tokens used.
func RunEfficiency(chunks []Chunk, cfg Config) Efficiency {
	var e Efficiency
	prefix := 0
	for _, c := range chunks {
		e.TotalTokensUsed += c.Tokens
		e.StandardCost += prefix + c.Tokens
		prefix += c.Tokens
	}
	e.MarkovianCost = len(chunks)*cfg.CarryoverTokens + e.TotalTokensUsed
	if s := e.StandardCost - e.MarkovianCost; s > 0 {
		e.TokensSaved = s
	}
	if e.StandardCost > 0 {
		e.EfficiencyPercent = float64(e.TokensSaved) / float64(e.StandardCost) * 100
	}
	return e
}

// CostModel holds the constants of docs/efficiency-model.md: framing F, goal G, step input S, per-step output O
// and carryover C, all in tokens.
type CostModel struct {
	F, G, S, O, C int
}

// DocDefaults are the efficiency model's documented constants; ModelTotals(20, DocDefaults) is its 86% example.
var DocDefaults = CostModel{F: 400, G: 200, S: 300, O: 1200, C: 800}

// Totals is the modelled input-token total of N steps, naive and Markovian, and the share saved (0 to 1).
type Totals struct {
	NaiveTotal     int     `json:"naiveTotal"`
	MarkovianTotal int     `json:"markovianTotal"`
	Savings        int     `json:"savings"`
	Efficiency     float64 `json:"efficiency"`
}

// ModelTotals is docs/efficiency-model.md's total cost across n steps:
// naive = n(F+G+S) + O·n(n-1)/2, which grows with the square of n, and Markovian = n(F+G+C+S), which grows with n.
// n(n-1) is always even, so the naive total is a whole number.
func ModelTotals(n int, k CostModel) Totals {
	t := Totals{
		NaiveTotal:     n*(k.F+k.G+k.S) + k.O*n*(n-1)/2,
		MarkovianTotal: n * (k.F + k.G + k.C + k.S),
	}
	t.Savings = t.NaiveTotal - t.MarkovianTotal
	if t.NaiveTotal != 0 {
		t.Efficiency = float64(t.Savings) / float64(t.NaiveTotal)
	}
	return t
}
