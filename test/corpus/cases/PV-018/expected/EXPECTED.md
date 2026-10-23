# PV-018 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in the tests shipped with `change/` (the claimed property is not pinned)
- the finding that must be PROVEN: the shipped budget tests pass on both change and the mutant that uses a list of page
  names, so nothing in the change pins "derived from content, never listed".
- evidence required: the suite lines showing change = pass and mutant = pass; and the planted-row line showing change =
  fail at the narrow budget, mutant = pass, reference = fail. That line shows the two rules differ, and that change
  behaves as claimed today.
- what must NOT be claimed: that the validator is wrong on today's data, or on a planted row — it refuses the planted row
  correctly (reviewer A's probe was accurate). Nor that reviewer A's probe proves the tests guard the rule: it did not
  test the tests.
- how the record settled it: the orchestrating agent rejected the change with this as a blocking finding, in its words
  the rule "behaves correctly and nothing in the branch pins it" — reviewer B right on the pin, reviewer A right on the
  behaviour. A repair round was rejected again on the same class, and the work landed after further rounds. The label
  rests on the mutation result, not on which reviewer rejected.
