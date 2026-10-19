# PV-013 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: setting the name field to the synthetic token in `change/` is refused, and the
  refusal message contains the token.
- evidence required: the check line for that input, showing change = refused and token in output = yes; and the
  change's own no-leak control passing on `change/`, which shows that the shipped control never walks the refusal path.
- what must NOT be claimed: that the refusals do not fire — they do; or that a token kept in the credentials file leaks
  through show — it does not (reviewer A's measurements were accurate; they did not cover the refusal's own message).
- how the record settled it: a later round conceded the finding and stopped the refusal from printing the rejected
  value. The label rests on the executed check and the later repair, not on which reviewer rejected.
