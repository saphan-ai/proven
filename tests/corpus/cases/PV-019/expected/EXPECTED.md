# PV-019 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: the change's validator passes a `preview` row whose per-session switch is on or whose
  plugin folder is set, and an `unsupported` row with an install route, while the reference refuses all three.
- evidence required: the check lines for those three inputs showing change = pass and reference = fail; the line for the
  shipped table showing all agree (the hole is latent, not live); and the line for a wrong failure channel showing
  validator = pass while the shipped content test is red.
- what must NOT be claimed: that a shipped row is wrong today — none is; or that the wrong failure channel goes unnoticed
  — the content test catches the one known lie (reviewer B's facts were accurate; its grading was too low).
- how the record settled it: the orchestrating agent upheld reviewer A's two holes as blocking and raised reviewer B's
  observation (the validator accepts any non-empty failure channel) to blocking as well. A repair round closed all three:
  the failure channel became a closed set, an `unsupported` row with an install route is refused, and a `preview` row must
  have every delivery field empty. The label rests on the executed check and that repair, not on which reviewer rejected.
