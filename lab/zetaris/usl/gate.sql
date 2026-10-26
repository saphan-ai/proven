-- rule: usl-epss-score-valid
-- The data-quality rule that gates the layer. The query returns the violating rows, so no row is a pass.
-- A row of px_epss_score violates the rule when
--   its cve_id is empty or is not CVE-, four digits, a hyphen and four to twelve digits (Proven's cve_id format), or
--   its epss is not a number from 0 to 1, or
--   its percentile is not a number from 0 to 1.
-- A value that is not a number casts to NULL and violates. NaN violates. docs/ZETARIS_USL.md says why this rule.
-- The file is one SELECT with no semicolon. apply.sh counts the rows it returns.
SELECT cve_id, epss, percentile, score_date
FROM px_usl.px_epss_score
WHERE cve_id IS NULL
   OR NOT (cve_id RLIKE '^CVE-[0-9]{4}-[0-9]{4,12}$')
   OR CAST(epss AS DOUBLE) IS NULL
   OR isnan(CAST(epss AS DOUBLE))
   OR CAST(epss AS DOUBLE) < 0
   OR CAST(epss AS DOUBLE) > 1
   OR CAST(percentile AS DOUBLE) IS NULL
   OR isnan(CAST(percentile AS DOUBLE))
   OR CAST(percentile AS DOUBLE) < 0
   OR CAST(percentile AS DOUBLE) > 1
