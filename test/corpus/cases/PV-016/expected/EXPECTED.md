# PV-016 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in **change**
- the finding that must be PROVEN: the pages promise alerts that the code at this version cannot deliver — every
  emitted alert is considered and none is delivered, because the sender always resolves to one that sends nothing.
- evidence required: the check lines for the alerts the pages list (delivered = 0 for each); the line for the documented
  setup step (refused by the code); and, for a pages-only review, the four page-internal findings matched against the
  reference list.
- what must NOT be claimed: that the pages are consistent apart from the code (they contradict each other), or that the
  pages-only reviewer was wrong — within its subject its findings were accurate (reviewer A); the no-delivery defect is
  not visible from the pages alone. Also not: a defect in the code's other operations, which work as written.
- how the record settled it: the next round counted the first round as rejected twice and adopted all eight blocking
  items, four from each reviewer, with one overlap. The missing delivery was closed by another change that added a real
  sender; the orchestrating agent then measured delivery end to end (three considered, three delivered) before the pages
  were allowed to say the feature works. The documentation was accepted with conditions and merged several rounds later.
  The label rests on that round and that measurement, not on which reviewer rejected.
