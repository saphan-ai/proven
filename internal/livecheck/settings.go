// Package livecheck declares the one setting the livecheck command reads
// that a review does not. The query adapter has no view.
//
// It sits outside the review. The livecheck command asks each live service a
// review needs, including Zetaris, the Investigator's query source
// (ADR-0015), before a demo; this package names the view it queries, so the
// configuration page lists it beside the review's own settings (ADR-0013).
package livecheck

import "github.com/saphan-ai/proven/internal/envdoc"

const (
	// EnvEvidenceView is the evidence view livecheck queries.
	EnvEvidenceView = "PROVEN_EVIDENCE_VIEW"
	// DefaultEvidenceView is used when EnvEvidenceView is empty. It is the
	// known-exploited vulnerabilities view loaded into Zetaris (ADR-0023).
	DefaultEvidenceView = "px_evidence.px_kev_vulns"
)

// Environment is the livecheck setting the configuration page lists.
// It sits beside EnvEvidenceView so the name and the default are the code.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{{
		Name:        EnvEvidenceView,
		Default:     envdoc.Tick(DefaultEvidenceView),
		Secret:      false,
		Description: "the view `livecheck` queries to prove Zetaris answers (livecheck only)",
	}}
}
