CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-4242', 'example.com', 'widget', '2026-04-01', 'Unknown');

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-4242', '0.42000', '0.91000', '2026-10-21');

CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-w19-sample-0001', 'CVE-2026-4242', 'Go', 'example.com/widget', '1.3.0', '1.5.0', '', 'HIGH', '2026-04-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-w19-sample-0002', '', 'Go', 'example.com/widget', '2.0.0', '2.1.0', '', 'LOW', '2026-05-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-w19-other', '', 'Go', 'example.com/other', '0', '1.0.0', '', 'LOW', '2026-06-01');
