# PV-001 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: an element with `data-screen="<not required>" data-screen="<required>"` makes
  `checkFrames` pass in `change/` while the reference parser reads the first value, so the required frame is not there.
- evidence required: the check line for that input showing change = pass and reference = first value; and the same line
  for a document with a single real frame, showing all three agree (the guard is not simply broken).
- what must NOT be claimed: that the change fails the order's named attack positions — it refuses them all (reviewer A's
  measurements were accurate; they were incomplete).
- how the record settled it: a later round conceded the finding and repaired the guard to read the first duplicate.
  The label rests on the HTML standard and the later repair, not on which reviewer rejected.
