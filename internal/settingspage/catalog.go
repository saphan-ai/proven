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
	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/limit"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/store"
	"github.com/saphan-ai/proven/internal/telemetry"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/tool"
)

//go:generate go run ../../cmd/settings-doc

// policyAfter follows the policy table. It is not a setting.
const policyAfter = "A repository's `proven.yaml` may make a policy setting stricter, never weaker, and an exception without an owner or an expiry is refused on load (ADR-0036)."

// guardrailAfter is the sentence that follows the guardrail table. It is
// not a setting. It stays inside the generated block so the one marker pair
// still covers every environment table.
const guardrailAfter = "Every prompt and every answer is checked; an `unsafe` verdict, or a guardrail that does not answer, stops that step by name. An empty key is not off: every model call is refused and the review names the guardrail."

// containerAfter follows the container table. It is not a setting.
const containerAfter = "Each test runs with no network and no secret (ADR-0025). With no container runtime the Prover is off, and the review says so. Sample mode reads none of these."

// toolAfter follows the tool table. It is not a setting.
const toolAfter = "Each tool runs in the Prover's container with no network and no secret (ADR-0025, ADR-0040). A tool that cannot run is named. Sample mode reads none of these."

// readsAfter follows the read-bound table. It is not a setting.
const readsAfter = "A read over its bound is refused by name before the review goes on; the refusal names the file and the bound, never its content. Sample mode reads none of these."

// telemetryAfter follows the telemetry table. It is not a setting.
const telemetryAfter = "Telemetry is off unless an endpoint is set. It only observes: once a review has started it never changes what the review decides, and a collector that cannot be reached never fails a review (the export is bounded at 5 seconds, and the failure goes to stderr). The one refusal is at the start: a malformed `OTEL_*` setting stops `Open` by name, like any other setting. No secret is put in an attribute, an event or a span status: the same redactor that cleans the trace file cleans the span events and the text of every failed review and failed agent, so a status carries the redactor's marker where a configured secret stood. The headers, timeout and certificates of the exporter are the SDK's own variables (`OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_TIMEOUT`, `OTEL_EXPORTER_OTLP_CERTIFICATE`), read from the process environment. Sample mode reads the `OTEL_*` variables and no other."

// generatorAfter says, per mode, which keys the generator needs and where to get them (ADR-0031 point 6).
const generatorAfter = "Which keys each mode needs (ADR-0031). `nightshift` (the default, ADR-0019): `METERLESS_RUNTIME_TOKEN`, issued by the operator of the Nightshift you run; nothing else is required. `nvidia`: `PROVEN_GENERATOR_MODE=nvidia` and `PROVEN_GENERATOR_KEY`, a free NVIDIA API key created at https://build.nvidia.com (sign in, open any model, choose *Get API Key*); the hosted API's address (`PROVEN_NVIDIA_BASE_URL`, `https://` only) and the models have defaults, so the key is the only value needed. Each mode reads its own address: `nightshift` reads `PROVEN_GENERATOR_BASE_URL`, `nvidia` reads `PROVEN_NVIDIA_BASE_URL` and ignores the other, so the key is never sent to a Nightshift address. The guardrail has its own key in either mode (below). In `nvidia` mode the hosted API gives no signed receipt, so the trace says *no receipt* and the `nvidia/*` requirement is checked against the `model` field of its answer. A free key is rate limited; a review that meets the limit names it as a failed service. Sample mode reads none of these."

// evidenceAfter follows the evidence table. It is not a setting.
const evidenceAfter = "The store is zetaris unless PROVEN_EVIDENCE_STORE says local or couchdb. The local store is one file and opens no network (ADR-0034, ADR-0047), and a review that chooses it does not read the Zetaris variables. The couchdb store reads the PROVEN_COUCHDB variables and not the Zetaris ones. A live review passes the template request, and each store runs the form of its own language: sql on Zetaris and the local store, mango on CouchDB (ADR-0047, docs/ADAPTERS.md). Sample mode uses the local store on the fixture file and reads none of these, except that it reads PROVEN_COUCHDB_RECORDS and PROVEN_EVIDENCE_LINK_BASE_URL, and, when PROVEN_COUCHDB_RECORDS is set, also the other PROVEN_COUCHDB_* variables (URL, USER, PASSWORD, PREFIX, ROW_LIMIT, TIMEOUT, RECORDS_USER, RECORDS_PASSWORD) and GITHUB_REF, and nothing else besides OTEL_*, so the lab's review can write its records (docs/ADAPTERS.md)."

// packAfter follows the evidence-pack table. It is not a setting.
const packAfter = "The pack holds source excerpts and query rows, so keep it where the review itself may be read. The private key is never written into it (ADR-0035). See `docs/EVIDENCE_PACK.md`."

// hmemAfter follows the H-MEM table. It is not a setting.
const hmemAfter = "H-MEM reads its whole state file at open and scans every record on each query, so it is built for a small store; see the scale section of `docs/ARCHITECTURE.md`. A state file over the cap is refused by name with its size and the cap. Sample mode reads none of these."

// Catalog is the environment sections of docs/CONFIGURATION.md, in page order.
// PROVEN_EPSS_VIEW is declared in epss. PROVEN_KEV_VIEW is declared in osv,
// beside PROVEN_OSV_TABLE. PROVEN_EVIDENCE_VIEW is declared in livecheck.
// PROVEN_ADAPTER_TIMEOUT is declared in limit.
// PROVEN_EVIDENCE_STORE and PROVEN_EVIDENCE_FILE are declared in store.
// The evidence table joins the store choice, the Zetaris adapter, OSV, EPSS
// and livecheck.
func Catalog() []envdoc.Section {
	evidenceRows := append(store.Environment(), query.Environment()...)
	evidenceRows = append(append(append(evidenceRows, osv.Environment()...), epss.Environment()...), livecheck.Environment()...)
	return []envdoc.Section{
		{
			Heading: "The model generator: Meterless Nightshift or NVIDIA's hosted API",
			Rows:    generator.Environment(),
			After:   generatorAfter,
		},
		{
			Heading: "The second reviewer: any OpenAI-compatible model",
			Rows:    generator.SecondEnvironment(),
		},
		{
			Heading: "The guardrail: NVIDIA Nemotron content safety",
			Rows:    guardrail.Environment(),
			After:   guardrailAfter,
		},
		{
			Heading: "Evidence: the store, and Zetaris over REST",
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
	}
}
