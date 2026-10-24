CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-9090', 'example.com', 'monolog', '2026-10-24', 'Unknown');

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-9090', '0.25000', '0.91000', '2026-10-24');

CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-php-sample-0001', 'CVE-2026-9090', 'Packagist', 'monolog/monolog', '2.4.0', '2.10.0', '', 'HIGH', '2026-10-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-php-sample-0002', '', 'Packagist', 'monolog/monolog', '3.0.0', '3.1.0', '', 'LOW', '2026-10-02');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-php-other', '', 'Packagist', 'other/unused', '0', '9.0.0', '', 'LOW', '2026-10-03');
