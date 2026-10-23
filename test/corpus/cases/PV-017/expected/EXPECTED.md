# PV-017 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in the guard test shipped with `change/`
- the finding that must be PROVEN: with `{project}` removed from a project-bearing template in every catalogue, the
  change's guard still passes; only its logged count of examined templates falls.
- evidence required: the check line for that input showing guard = pass, a lower count than for the shipped catalogues,
  and reference = fail naming the template; and the line for the shipped catalogues showing guard and reference both
  pass (the renderer and catalogues are not broken today).
- what must NOT be claimed: that the renderer is wrong, or that a shipped notice lacks its project line — none does
  (reviewer B's measurements were accurate and it said so). Nor that reviewer A's reject found a defect: it measured
  only that its copy lacked the commit, which was accurate and says nothing about the code.
- how the record settled it: a later round first reproduced the finding on the base (placeholder removed, suite green,
  count falling), then replaced the self-selected population with one taken from a separate declaration of which
  templates must carry the placeholder; its review re-measured both. The label rests on that reproduction, not on which
  reviewer rejected.
