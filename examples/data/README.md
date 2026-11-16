# Prepared evidence data

`snapshot.json` is a **synthetic teaching dataset**, not a vulnerability feed. The advisory, CVE and scores are invented. It demonstrates the import format for SQLite/PostgreSQL and the difference between an empty registered source and a missing source.

The [data walkthrough](../../docs/DATA_WALKTHROUGH.md) creates a tiny Git repository, imports this dataset and checks a version raise without a model or container runtime. Keep the imported database outside the reviewed repository. Never use this dataset for an operational security decision.
