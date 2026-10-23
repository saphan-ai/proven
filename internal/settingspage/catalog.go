// Package settingspage lists the environment settings the configuration
// page is generated from. Each row is declared beside the constant it names.
//
// It is the one list ADR-0013 asks for. cmd/settings-doc renders the page's
// tables from it, and the test in this package fails when the page and the
// code disagree.
package settingspage

import (
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

//go:generate go run ../../cmd/settings-doc

// guardrailAfter is the sentence that follows the guardrail table. It is
// not a setting. It stays inside the generated block so the one marker pair
// still covers every environment table.
const guardrailAfter = "Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name. An empty key is not off: every model call is refused and the review names the guardrail."

// containerAfter follows the container table. It is not a setting.
const containerAfter = "Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these."

// readsAfter follows the read-bound table. It is not a setting.
const readsAfter = "A read over its bound is refused by name before the review goes on; the refusal names the file and the bound, never its content. Sample mode reads none of these."

// Catalog is the environment sections of docs/CONFIGURATION.md, in page order.
// PROVEN_EPSS_VIEW is declared in epss. PROVEN_KEV_VIEW is declared in osv,
// beside PROVEN_OSV_TABLE. PROVEN_EVIDENCE_VIEW is declared in livecheck.
// PROVEN_ADAPTER_TIMEOUT is declared in limit.
// The evidence table joins the rows of four packages: the Zetaris adapter,
// OSV, EPSS and livecheck.
func Catalog() []envdoc.Section {
	evidence := append(append(append(query.Environment(), osv.Environment()...), epss.Environment()...), livecheck.Environment()...)
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
		{
			Heading: "One adapter call",
			Rows:    limit.Environment(),
		},
		{
			Heading: "Reads of the repository under review",
			Rows:    change.Environment(),
			After:   readsAfter,
		},
	}
}
