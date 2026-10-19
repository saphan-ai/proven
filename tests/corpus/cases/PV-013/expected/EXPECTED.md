# PV-013 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: setting `credentials_env` to a token-shaped value in `change/` returns an error whose
  text contains the token.
- evidence required: the check line showing the token inside the error text; and `go test ./...` green in `change/`,
  which shows the shipped no-leak test does not walk the refusal path.
- what must NOT be claimed: that the refusals do not fire — they do (reviewer A's measurements were accurate).
- how the record settled it: a later round stopped the refusal from printing the value. The label rests on the
  executed check, not on which reviewer rejected.
