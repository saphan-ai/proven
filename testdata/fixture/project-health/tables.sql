CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);

CREATE TABLE px_evidence.px_scorecard (ecosystem TEXT, package TEXT, maintained TEXT, signed TEXT, release_age_days TEXT);
INSERT INTO px_evidence.px_scorecard VALUES ('Go', 'evil.example/young', '0', 'true', '400');
