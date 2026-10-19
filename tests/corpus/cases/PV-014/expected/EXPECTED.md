# PV-014 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: with server A answering 307 (and 308) to server B on another port, B receives a body
  containing the synthetic token.
- evidence required: the check lines for 307 and 308 (token at B) and the 302 control (no body at B); `go test ./...`
  green in `change/`, which shows the shipped tests do not cover the HTTP path.
- what must NOT be claimed: that the import guard or the destination check is wrong — both work (reviewer A's
  measurements were accurate within its assigned subject).
- how the record settled it: the orchestrating agent verified the finding by code and Go semantics and recorded a
  reject; the repair added a client that refuses redirects to other hosts. The label rests on the executed redirect check.
