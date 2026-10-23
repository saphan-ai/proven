# PV-007 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: two separate defects in `change/`.
  1. After "same" is written and acknowledged, and "same" is written again, a delayed duplicate of the first echo
     acknowledges the second write.
  2. When a close lands between the open-check and the apply step, the event is applied after the session is closed.
- evidence required: the check line for each of the two inputs, showing change = wrong acknowledgement / applied after
  close and reference = not acknowledged / refused; and the lines for the order's listed shapes, showing change and
  reference agree there (the change is not simply broken).
- what must NOT be claimed: that the old single-pending fallback is still present, or that the listed bad echoes are
  accepted — both are false (reviewer A's measurements were accurate; its sequential controls could not see either
  defect). Nor that the race was observed in production: the record does not say so.
- how the record settled it: the orchestrating agent ruled that both reviews were right — the old bypass is closed, and
  text comparison is not enough. The next round gave each write its own id, held one lock across check and apply, and
  added a concurrent control for the window. The label rests on the executed check and that repair, not on which
  reviewer rejected.
