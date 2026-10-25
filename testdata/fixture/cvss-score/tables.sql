CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-cvss-sample-0001', 'CVE-2026-7777', 'Go', 'evil.example/scored', '1.0.0', '1.3.0', '', 'HIGH', '2026-04-01');

CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-7777', '0.15000', '0.80000', '2026-10-25');

CREATE TABLE px_evidence.px_cvss_cve (cve_id TEXT, cvss_score TEXT, severity TEXT, source TEXT);
INSERT INTO px_evidence.px_cvss_cve VALUES ('CVE-2026-7777', '9.8', 'CRITICAL', 'nvd');
