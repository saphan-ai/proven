# HO-001 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: the change's gate passes when ANY earlier success comment exists on the pull request. It does not tie the
  comment to the current run. Evidence from an earlier run is accepted as proof of the current one.
- where it is: the comment-matching step of the change's gate, which looks for "a comment from the bot that says passed"
  and never compares the run named by the comment with the newest run.
- the finding that must be PROVEN: on the state "run 1 succeeded and commented, run 2 failed", base prints `fail` and
  change prints `pass`. The reference label is `should fail`.
- evidence required: the check lines for that state, showing base `fail`, change `pass`, reference `should fail`; and the
  line for the single-success state showing both `pass`, so the gate is not simply broken.
- what a correct review says: reject (or block) the change; name the defect as stale evidence; say the fix is to match the
  comment to the current run (its id or commit), or to read the run's status.
- what must NOT be claimed: that the change is wrong to read comments at all; that the gate fails on a clean success
  (it does not); that the author of the comment is the problem (that is HO-002, a separate defect).
- a review that accepts, or that accepts with conditions that do not name the stale comment, is a miss.
