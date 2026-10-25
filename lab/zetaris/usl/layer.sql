-- The Proven semantic layer over the evidence (docs/ZETARIS_USL.md).
--
-- The repository carries no Zetaris documentation of a "Unified Semantic Layer" object, so the layer is plain
-- views in one SchemaStore container, px_usl. The form is the one the registration runbook already uses
-- (CREATE SCHEMASTORE CONTAINER, CREATE SCHEMASTORE VIEW ... WITH CONTAINER ... AS SELECT).
-- A statement ends with a semicolon at the end of its line. A comment is a whole line. A statement has no
-- comment inside it. apply.sh sends the statements in order, one call each, as the administrator.
-- A "tolerate" line lets the next statement be refused when the refusal says that word, for example a drop
-- of a view that is not there yet. Nothing else is tolerated.
--
-- ENTITIES. Each entity view keeps the columns, and the names of the columns, of the table Proven's template
-- already reads, so no template changes. Source and key are named on each.
--
-- RELATIONS. The relations are join keys. They are written in px_package_evidence at the end, and each is
-- named there. A real USL object would add what plain views cannot: a declared relationship that a client can
-- list and follow, a description per column, and the data-quality rule bound to the entity instead of kept in
-- gate.sql. None of that was measured on this build.

-- tolerate: exist
CREATE SCHEMASTORE CONTAINER px_usl;

-- ENTITY package. One row is one affected range of one advisory of one package.
-- Source: px_osvfile.px_osv_pkg (docs/OSV_TABLE.md). Key: (ecosystem, package). Read by osv-by-package.
-- tolerate: exist
DROP SCHEMASTORE VIEW px_package FROM px_usl;
CREATE SCHEMASTORE VIEW px_package WITH CONTAINER px_usl AS
SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published
FROM px_osvfile.px_osv_pkg;

-- ENTITY OSV advisory. One row is one advisory id and one alias of it.
-- Source: px_evidence.px_osv_id (ADR-0023). Key: id. Read by osv-by-advisory.
-- tolerate: exist
DROP SCHEMASTORE VIEW px_advisory FROM px_usl;
CREATE SCHEMASTORE VIEW px_advisory WITH CONTAINER px_usl AS
SELECT id, alias, published, severity
FROM px_evidence.px_osv_id;

-- ENTITY KEV CVE. One row is one CVE in the CISA known-exploited list.
-- Source: px_evidence.px_kev_vulns (ADR-0023). Key: cve_id. Read by kev-by-cve.
-- tolerate: exist
DROP SCHEMASTORE VIEW px_kev_cve FROM px_usl;
CREATE SCHEMASTORE VIEW px_kev_cve WITH CONTAINER px_usl AS
SELECT cve_id, vendor_project, product, date_added, known_ransomware
FROM px_evidence.px_kev_vulns;

-- ENTITY EPSS score. One row is the EPSS score of one CVE on one day.
-- Source: px_evidence.px_epss_cve_rows (ADR-0023). Key: cve_id. Read by epss-by-cve.
-- The data-quality rule of gate.sql is read over this view.
-- tolerate: exist
DROP SCHEMASTORE VIEW px_epss_score FROM px_usl;
CREATE SCHEMASTORE VIEW px_epss_score WITH CONTAINER px_usl AS
SELECT cve_id, epss, percentile, score_date
FROM px_evidence.px_epss_cve_rows;

-- RELATIONS. One row per package range, per alias of its advisory, per EPSS score date. That is the grain: the
-- relation view does not collapse aliases or score dates, it shows each one, so a package range with two aliases
-- is two rows. A package range with no advisory row, an alias that is not a CVE and a CVE with no score are each
-- one row with empty columns. The grain assumes px_kev_cve holds one row per CVE; a repeated KEV row repeats the
-- rows of that CVE. docs/ZETARIS_USL.md, "The grain of the relation view".
-- package has advisories:  px_package.id = px_advisory.id
-- advisory names CVEs:     px_advisory.alias = cve id
-- CVE may be in KEV:       cve id = px_kev_cve.cve_id
-- CVE has an EPSS score:   cve id = px_epss_score.cve_id
-- Every join is a LEFT JOIN. A join is an inner join by default (ADR-0023), and a package with no advisory
-- row, an advisory with no KEV row or a CVE with no score must stay a row, because a missing row is not
-- "not exploited" and not "not scored". The advisory side holds only the advisories an administrator
-- registered (ADR-0023, point 3), so a package whose advisory has no row there is kept with empty advisory
-- columns.
-- tolerate: exist
DROP SCHEMASTORE VIEW px_package_evidence FROM px_usl;
CREATE SCHEMASTORE VIEW px_package_evidence WITH CONTAINER px_usl AS
SELECT p.ecosystem AS ecosystem, p.package AS package, p.id AS advisory_id, p.introduced AS introduced, p.fixed AS fixed, a.alias AS cve_id, k.date_added AS kev_date_added, k.known_ransomware AS kev_known_ransomware, e.epss AS epss, e.percentile AS epss_percentile, e.score_date AS epss_score_date
FROM px_usl.px_package p
LEFT JOIN px_usl.px_advisory a ON p.id = a.id
LEFT JOIN px_usl.px_kev_cve k ON a.alias = k.cve_id
LEFT JOIN px_usl.px_epss_score e ON a.alias = e.cve_id;
