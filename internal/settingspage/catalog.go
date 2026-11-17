// Package settingspage lists the environment settings the configuration
// page is generated from. Each row is declared beside the constant it names.
//
// It is the one list  asks for. cmd/settings-doc renders the page's
// tables from it, and the test in this package fails when the page and the
// code disagree.
package settingspage

import (
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/feeds"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/pipeline"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/store"
	"github.com/saphan-ai/proven/internal/telemetry"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/tool"
	"github.com/saphan-ai/proven/internal/trace"
)

//go:generate go run ../../cmd/settings-doc

// policyAfter follows the policy table. It is not a setting.
const policyAfter = "A repository's `proven.yaml` may make a policy setting stricter, never weaker, and an exception without an owner or an expiry is refused on load (ADR-0036)."

// guardrailAfter is the sentence that follows the guardrail table. It is
// not a setting. It stays inside the generated block so the one marker pair
// still covers every environment table.
const guardrailAfter = "When enabled, the safety service checks every prompt and answer. An unsafe response or unavailable check stops that model step. A self-hosted service may use no key; hosted NVIDIA mode requires one."

// containerAfter follows the container table. It is not a setting.
const containerAfter = "Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these."

// toolAfter follows the tool table. It is not a setting.
const toolAfter = "Each tool runs in the Prover's container with no network and no secret (ADR-0025, ADR-0040). A tool that cannot run is named. Sample mode reads none of these."

// readsAfter follows the read-bound table. It is not a setting.
const readsAfter = "A read over its bound is refused by name before the review goes on; the refusal names the file and the bound, never its content. Sample mode reads none of these."

// telemetryAfter follows the telemetry table. It is not a setting.
const telemetryAfter = "Telemetry is off unless an endpoint is set. It only observes: once a review has started it never changes what the review decides, and a collector that cannot be reached never fails a review (the export is bounded at 5 seconds, and the failure goes to stderr). The one refusal is at the start: a malformed `OTEL_*` setting stops `Open` by name, like any other setting. No secret is put in an attribute, an event or a span status: the same redactor that cleans the trace file cleans the span events and the text of every failed review and failed agent, so a status carries the redactor's marker where a configured secret stood. The headers, timeout and certificates of the exporter are the SDK's own variables (`OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_TIMEOUT`, `OTEL_EXPORTER_OTLP_CERTIFICATE`), read from the process environment. Sample mode reads the `OTEL_*` variables and no other."

// generatorAfter says, per mode, which keys the generator needs and where to get them .
const generatorAfter = "Without model configuration, deterministic analysis runs alone. The openai adapter uses your endpoint and optional API key; command uses an operator-configured local CLI. Explicit nightshift and nvidia adapters retain their provider-specific settings. Sample mode uses recorded answers."

// evidenceAfter follows the evidence table. It is not a setting.
const evidenceAfter = "A clean installation selects SQLite. PostgreSQL, CouchDB, local fixture files and Zetaris are optional stores. Query templates and source definitions are independent of the backend. Explicit legacy Zetaris configuration retains its selection during migration. Sample mode uses local fixtures; optional evidence records and telemetry keep their own settings."

// packAfter follows the evidence-pack table. It is not a setting.
const packAfter = "The pack holds source excerpts and query rows, so keep it where the review itself may be read. The private key is never written into it (ADR-0035). See `docs/EVIDENCE_PACK.md`."

// hmemAfter follows the H-MEM table. It is not a setting.
const hmemAfter = "H-MEM is bounded by the configured state-size and query limits. It is intended for a small repository store. A state file above its limit is refused. Sample mode uses the defaults."

// boundedAfter follows the bounded-steps table. It is not a setting.
const boundedAfter = "Bounded steps are on by default and can be disabled with PROVEN_MARKOVIAN=off. Their effect on review quality is unmeasured. Each trace records whether the bounded planner ran and the size of its steps and carried state. Sample mode reads the same switches."

// storesAfter follows the stores table. It is not a setting.
const storesAfter = "Both local stores are on by default and have independent off switches. A disabled store is not opened, read or written. Their effect on review quality is unmeasured. Sample mode reads the same switches."

// networkAfter follows the network table. It is not a setting.
const networkAfter = "In local mode, every configured network destination must pass the policy. Hosted NVIDIA modes and the external CLI model transport are refused. Runner images are not pulled. See AIR_GAPPED.md."

// Catalog is the environment sections of docs/CONFIGURATION.md, in page order.
// PROVEN_EPSS_VIEW is declared in epss. PROVEN_KEV_VIEW is declared in osv,
// beside PROVEN_OSV_TABLE.
// PROVEN_ADAPTER_TIMEOUT is declared in limit.
// PROVEN_EVIDENCE_STORE and PROVEN_EVIDENCE_FILE are declared in store.
// The evidence table joins the store choice, the Zetaris adapter, OSV, EPSS
// settings.
func Catalog() []envdoc.Section {
	evidenceRows := append(store.Environment(), query.Environment()...)
	evidenceRows = append(append(evidenceRows, osv.Environment()...), epss.Environment()...)
	return []envdoc.Section{
		{
			Heading: "The network mode: air-gapped",
			Rows:    network.Environment(),
			After:   networkAfter,
		},
		{
			Heading: "Model transports",
			Rows:    generator.Environment(),
			After:   generatorAfter,
		},
		{
			Heading: "The second reviewer: API or CLI",
			Rows:    generator.SecondEnvironment(),
		},
		{Heading: "PR delivery assessment", Rows: review.DeliveryEnvironment()},
		{Heading: "Code comment review", Rows: review.CommentsEnvironment()},
		{Heading: "Source refresh and delta cache", Rows: feeds.Environment()},
		{
			Heading: "Optional model safety service",
			Rows:    guardrail.Environment(),
			After:   guardrailAfter,
		},
		{
			Heading: "Evidence stores and source settings",
			Rows:    evidenceRows,
			After:   evidenceAfter,
		},
		{
			Heading: "The Prover's container",
			Rows:    testrun.Environment(),
			After:   containerAfter,
		},
		{
			Heading: "Deterministic tools",
			Rows:    tool.Environment(),
			After:   toolAfter,
		},
		{
			Heading: "Bounded steps: Meterless's Markovian engine",
			Rows:    markov.Environment(),
			After:   boundedAfter,
		},
		{
			Heading: "The two stores: Meterless's H-MEM and World Model",
			Rows:    pipeline.Environment(),
			After:   storesAfter,
		},
		{
			Heading: "One adapter call",
			Rows:    limit.Environment(),
		},
		{
			Heading: "Telemetry: OpenTelemetry (ADR-0045)",
			Rows:    telemetry.Environment(),
			After:   telemetryAfter,
		},
		{
			Heading: "Reads of the repository under review",
			Rows:    change.Environment(),
			After:   readsAfter,
		},
		{
			Heading: "The organisation policy",
			Rows:    review.EnvironmentPolicy(),
			After:   policyAfter,
		},
		{
			Heading: "The evidence pack",
			Rows:    evidence.Environment(),
			After:   packAfter,
		},
		{
			Heading: "H-MEM: the memory store",
			Rows:    hmem.Environment(),
			After:   hmemAfter,
		},
		{
			Heading: "The trace",
			Rows:    trace.Environment(),
			After:   "The default is the user cache directory, with proven/logs under it, created so only that user can read it. An existing configured directory is narrowed to 0700. A path that resolves inside the repository under review is refused when Proven loads, and the refusal names PROVEN_LOG_DIR and the target. The link count of each file written in that directory is read on every OS. Sample mode reads this setting.",
		},
	}
}
