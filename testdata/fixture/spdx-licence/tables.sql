CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);

CREATE TABLE px_evidence.px_spdx_lic (ecosystem TEXT, package TEXT, licence TEXT);
INSERT INTO px_evidence.px_spdx_lic VALUES ('Go', 'evil.example/copyleft', 'GPL-3.0-only');
