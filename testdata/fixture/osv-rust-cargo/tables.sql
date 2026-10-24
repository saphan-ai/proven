CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-9191', 'example.com', 'regex', '2026-10-24', 'Unknown');

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-9191', '0.33000', '0.91000', '2026-10-24');

CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-rs-sample-0001', 'CVE-2026-9191', 'crates.io', 'regex', '1.7.0', '1.10.0', '', 'HIGH', '2026-10-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-rs-sample-0002', '', 'crates.io', 'regex', '1.11.0', '1.12.0', '', 'LOW', '2026-10-02');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-rs-other', '', 'crates.io', 'unused_crate', '0', '9.0.0', '', 'LOW', '2026-10-03');
