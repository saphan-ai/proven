# PV-006 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: with `$wait` appended to the catalog entry, the product shows the raw `$wait` to the
  visitor while the change's text control passes; and a shown duration of 0 seconds against a retry-after value of 1
  passes the change's duration control.
- evidence required: the check line for the `$wait` input, showing change = pass and reference = fail; the line for
  zero against one, showing change = pass and reference = fail; and the lines for the well-formed entry and the earlier
  brace attacks, showing the change catches what it was built to catch (the controls are not simply broken).
- what must NOT be claimed: that the change misses the earlier brace attacks — it catches them all; or that reviewer
  A's holes were false — they were real and were repaired too (its measurements were accurate; they stayed inside the
  brace pattern). Nor that a second language was exposed: only one was served.
- how the record settled it: the orchestrating agent ruled that enumerating placeholder shapes is wrong in kind. The
  next round compared the whole rendered text with an exact expected text, which catches `$wait`, and required an
  exact duration comparison. It also fixed reviewer A's two holes. The label rests on the executed check and that
  repair, not on which reviewer rejected.
