# PV-005 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: for a request that names a stored version, change fetches "latest" and then "by
  version", while the reference allows exactly one "by version"; and the shipped test named "naming a version reads that
  version and nothing else" is green on that code.
- evidence required: the check line for the named, stored version (change fetch list = latest, by version; reference =
  by version only; the document returned is the right one); the shipped tests green; and the line for the malformed
  query, where change refuses with no fetch (the change does what the order asked).
- what must NOT be claimed: that the strict query parse is wrong — it works, and its test turns red when it is removed
  (reviewer A's measurements were accurate; they did not reach the reader); or that a wrong document is shown — none
  is.
- how the record settled it: a later round confirmed the finding at its source and repaired the reader to make one
  fetch per case; its strengthened test asserts the whole fetch list and fails on the old reader, naming the extra
  "latest". Both reviewers of that round confirmed the repair. The label rests on that repair and the contract, not on
  which reviewer rejected.
