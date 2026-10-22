// Package settingspage lists the environment settings the configuration
// page is generated from. Each row is declared beside the constant it names.
package settingspage

import (
	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

//go:generate go run ../../cmd/settings-doc

// guardrailAfter is the sentence that follows the guardrail table. It is
// not a setting. It stays inside the generated block so the one marker pair
// still covers every environment table.
const guardrailAfter = "Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name."

// containerAfter follows the container table. It is not a setting.
const containerAfter = "Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these."

// Catalog is the environment sections of docs/CONFIGURATION.md, in page order.
// PROVEN_EVIDENCE_VIEW is declared in livecheck and listed on the evidence
// table, which is where the page already shows it.
func Catalog() []envdoc.Section {
	evidence := append(query.Environment(), livecheck.Environment()...)
	return []envdoc.Section{
		{
			Heading: "The model generator: Meterless Nightshift",
			Rows:    generator.Environment(),
		},
		{
			Heading: "The guardrail: NVIDIA Nemotron content safety",
			Rows:    guardrail.Environment(),
			After:   guardrailAfter,
		},
		{
			Heading: "Evidence: Zetaris over REST",
			Rows:    evidence,
		},
		{
			Heading: "The Prover's container",
			Rows:    testrun.Environment(),
			After:   containerAfter,
		},
	}
}
