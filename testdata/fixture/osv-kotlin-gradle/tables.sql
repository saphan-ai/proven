CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-9291', 'example.com', 'widget', '2026-10-25', 'Unknown');

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-9291', '0.27000', '0.91000', '2026-10-25');

CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-kt-sample-0001', 'CVE-2026-9291', 'Maven', 'com.example:widget', '1.3.0', '1.5.0', '', 'HIGH', '2026-10-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-kt-sample-0002', '', 'Maven', 'com.example:widget', '2.0.0', '2.1.0', '', 'LOW', '2026-10-02');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-kt-other', '', 'Maven', 'com.example:unused', '0', '9.0.0', '', 'LOW', '2026-10-03');
