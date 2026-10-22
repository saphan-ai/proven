CREATE TABLE advisories (package TEXT, version TEXT, cve TEXT, kev TEXT);
INSERT INTO advisories VALUES ('evil.example/leftpad', '1.2.3', 'CVE-2026-0001', 'true');

CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001', 'evil.example', 'leftpad', '2026-01-15', 'Known');
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2021-44228', 'Apache', 'Log4j2', '2021-12-10', 'Known');
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0002', 'other.example', 'widget', '2026-02-01', 'Unknown');

CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2021-44228', '0.99999', '1.0', '2026-10-20');
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-0001', '0.50000', '0.90000', '2026-10-20');

CREATE TABLE px_evidence.px_epss_page_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_page_rows VALUES ('CVE-2026-99001', '0.01000', '0.40000', '2026-10-20');
INSERT INTO px_evidence.px_epss_page_rows VALUES ('CVE-2026-99002', '0.02000', '0.45000', '2026-10-20');

CREATE TABLE px_evidence.px_osv_id (id TEXT, alias TEXT, published TEXT, severity TEXT);
INSERT INTO px_evidence.px_osv_id VALUES ('GHSA-jfh8-c2jp-5v3q', 'CVE-2021-44228', '2021-12-10', 'CRITICAL');
INSERT INTO px_evidence.px_osv_id VALUES ('GHSA-px-sample-0001', 'CVE-2026-0001', '2026-01-15', 'HIGH');
