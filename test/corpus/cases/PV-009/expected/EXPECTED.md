# PV-009 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in change (also present in base: the change did not add it, it left it in
  place one step below its repair)
- the finding that must be PROVEN: a note whose closing number is 9223372036854775808 or longer passes in change while
  the reference reads it as stale; the guard treats "cannot convert" as "equals the pinned count".
- evidence required: the check lines for the twenty-digit and the 9223372036854775808 notes, showing change = pass and
  reference = stale; the 9223372036854775807 line, showing change = fail (the boundary); and the line for a stale pair
  after a current one, showing base = pass and change = fail (the repair is real).
- what must NOT be claimed: that the change's repair of the reader is wrong — it catches every shape the order named
  (reviewer A's measurements were accurate; they were incomplete); or that a published note carried such a number.
- how the record settled it: the next round took the finding as blocking, first reproduced it on the reviewed version
  (a twenty-digit closing number passed with all tests green), then split the outcome into three: differs is refused,
  equals is accepted, cannot convert is refused. The label rests on that reproduction and repair, not on which reviewer
  raised it.
