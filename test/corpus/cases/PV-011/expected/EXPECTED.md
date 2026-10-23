# PV-011 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in change
- the finding that must be PROVEN: change accepts rows with a text value over 800 bytes in a map, a list of lists, a
  nested record or a byte sequence, and a row of forty 700-byte values (28,000 bytes); and change's own tests pass on
  change-raised, so nothing independent holds the 800 ceiling.
- evidence required: the check lines for the map, list-of-lists, nested-record and byte-sequence inputs, showing change
  = accept and reference largest value over 800; the forty-value line, showing change = accept and a reference total of
  28,000; the 801-byte single-text and list lines, showing change = refuse (the budget is real where it looks); the
  compliant row, showing all three agree; and change's tests passing on change-raised.
- what must NOT be claimed: that a row served today is over budget or that a served sentence is false today — none is
  (reviewer A's grading point was accurate); or that reviewer A missed the facts — it measured most of the same escapes.
  The disagreement was on the grade, not on the facts.
- how the record settled it: the next round carried both reviews, listed five measured escapes, called the whole-row
  escape the one that voids the round's headline, closed all five (a walk over every field, a bound on each row's total,
  the ceiling pinned outside the constant) and merged. That matches the rejecting grade. Its own review pair (P25) then
  found more of the same; that repair did not merge. The label rests on the measured escapes and the next round's
  repair, not on which reviewer rejected.
