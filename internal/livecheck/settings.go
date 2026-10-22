// Package livecheck declares the one setting the livecheck command reads
// that a review does not. The query adapter has no view.
package livecheck

import "github.com/saphan-ai/proven/internal/envdoc"

const (
	// EnvEvidenceView is the evidence view livecheck queries.
	EnvEvidenceView = "PROVEN_EVIDENCE_VIEW"
	// DefaultEvidenceView is used when EnvEvidenceView is empty.
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
