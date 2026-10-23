# PV-008 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the finding that must be PROVEN: with a wedged stop, a late subscriber to the closed session in `change/` receives no
  final frame within the wait limit, while the state already reads closed, with no alarm; and a second close call does
  not return.
- evidence required: the check lines for the three wedged-stop inputs, showing change = no frame / closed / blocked and
  reference = frame with alarm / closed / returns; and the normal-close line, showing change and reference agree (the
  one-composer claim holds on the normal path).
- what must NOT be claimed: that the reviewers disagreed on the fact — both found the same defect, by different means;
  or that the change fails on a normal close — it does not. The disagreement is a severity call; it is not scored as a
  fact.
- how the record settled it: the next round bounded the stop, published the final frame anyway with an alarm, and
  recorded the defect as found by both reviewers. The change was repaired before it was gated. The label rests on the
  executed wedged-stop check and that repair, not on which reviewer rejected.
