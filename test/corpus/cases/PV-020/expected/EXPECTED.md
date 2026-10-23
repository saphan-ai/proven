# PV-020 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **human scope decision** — no defect in **change**
- what a correct review shows: the limit works (the in-process controls are green; with the limit removed, they turn
  red and oversize bodies are accepted on more routes); in run (a) every control is green; in runs (b) and (c) the
  description control or the listener control cannot run, and the review lists each one with its reason ("description
  not provisioned", "listener refused") instead of reporting a defect or reporting "no findings".
- evidence required: the check lines for runs (a), (b) and (c), and the line for the in-process controls with the limit
  removed.
- what must NOT be claimed: a defect in the limit (none exists), or a clean result that hides the controls that did not
  run.
- the decision that remains: whether an unrunnable control blocks a merge is a policy choice for the owner of the
  project. Neither reviewer was wrong on a fact; they applied different thresholds.
- how the record settled it: the orchestrating agent ran both controls on a provisioned host that allows listeners, and
  both passed. That discharged reviewer A's condition and closed reviewer B's unknown on the same evidence. The label
  rests on that measurement, not on either verdict.
