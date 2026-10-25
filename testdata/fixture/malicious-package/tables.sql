CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);

CREATE TABLE px_evidence.px_mal_pkg (id TEXT, ecosystem TEXT, package TEXT, summary TEXT);
INSERT INTO px_evidence.px_mal_pkg VALUES ('MAL-2024-1001', 'Go', 'evil.example/drop', 'published to take credentials');
