# PV-010 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in change
- the finding that must be PROVEN: a note closing on `-999`, `+41` or other-script digits passes in change, though the
  reference reads it as stale and the page says every closing number is compared; and change's own tests pass on
  change-edited, whose description states widths the matcher refuses.
- evidence required: the check lines for the signed and other-script notes, showing change = pass and reference =
  stale; the twenty-digit line, showing base = pass and change = fail (the previous round's repair is real); the
  description line, showing change-edited stating widths up to 256, the reference measuring up to 48, and change's tests
  passing on change-edited.
- what must NOT be claimed: that the three-way split of the cannot-convert outcome is wrong, or that `-999` produces a
  pair — it produces none, as reviewer A observed (its measurements were accurate; its reading of them was not); or
  that a published note carried such a number.
- how the record settled it: the next round adopted all three findings of the rejecting review as real. It gave one
  definition of a closing number, with signed and other-script numbers refused by name; removed the width part of the
  description; made the third check follow a value into the called function; and removed every published sentence that
  described the mechanism. Its executor first re-took each measurement on the reviewed version and reported that all
  held. The label rests on those measurements and the repair, not on which reviewer rejected.
