CREATE TABLE px_usl.px_epss_score (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_usl.px_epss_score VALUES ('CVE-2021-44228', '0.99999', '1.0', '2026-10-20');
INSERT INTO px_usl.px_epss_score VALUES ('CVE-2026-0001', '0.50000', '0.90000', '2026-10-20');
INSERT INTO px_usl.px_epss_score VALUES ('CVE-2026-0002', '0', '0', '2026-10-20');
