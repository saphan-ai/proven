# HO-002 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: the change's gate compares the run id and the commit named in the marker, but never checks which account
  posted the comment. The source of the evidence is unchecked, so a person can post the exact marker after a run that
  posted nothing, and the gate passes.
- where it is: the marker-matching step of the change's gate, which reads the comment body and ignores the comment's
  author.
- the finding that must be PROVEN: on the state "current run posted nothing; a person's account posted the exact marker
  for the current run and commit", change prints `pass`. The reference label is `should fail`.
- evidence required: the check line for that state, showing change `pass` and reference `should fail`; and the line for
  the bot's own marker showing `pass`, so the gate is not simply broken. The line for the state where the bot
  posted its marker for the current run and a person posted a marker for an older run shows change `pass` and reference
  `should pass`: the older marker neither helps nor hurts, and the change does not accept it as evidence on its own.
- what a correct review says: reject (or block) the change; name the defect as an unchecked source; say the fix is to
  accept the marker only from the bot's account.
- what must NOT be claimed: that the run and commit matching is wrong (it is right and is an improvement); that a stale
  marker passes (it does not; that is HO-001's defect, not this one).
- a review that accepts, or that accepts with conditions that do not name the author of the comment, is a miss.
