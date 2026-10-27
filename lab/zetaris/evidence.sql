-- The evidence definitions Proven queries (ADR-0023): Lightning REST tables over the public sources and the views
-- that turn their JSON into rows. A definition holds a URL and a request; no row is stored. One statement per
-- block, separated by a blank line; register.sh sends each block on its own to POST <REST>/sql-editor/sqls/run as
-- the Zetaris administrator. Names carry the prefix px_ (the platform's exploration objects).
-- Source: hackathon/fleet/infra/zetaris/zetaris-local.md section C and ADR-0023. NOT RUN against a platform:
-- the seat that wrote it had none. The first `make up ZETARIS=local` is the check.

CREATE LIGHTNING DATABASE PX_EVIDENCE_DB DESCRIBE BY "exploration"

CREATE SCHEMASTORE CONTAINER px_evidence

CREATE LIGHTNING REST TABLE px_kev FROM PX_EVIDENCE_DB REQUEST(endpoint "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json", method "get", response_type "json", http_encoding "URLENCODED") HEADER () BODY ()

CREATE SCHEMASTORE VIEW px_kev_vulns WITH CONTAINER px_evidence AS SELECT v.cveID AS cve_id, v.vendorProject AS vendor_project, v.product AS product, v.dateAdded AS date_added, v.knownRansomwareCampaignUse AS known_ransomware FROM px_evidence.px_kev LATERAL VIEW explode(vulnerabilities) AS v
