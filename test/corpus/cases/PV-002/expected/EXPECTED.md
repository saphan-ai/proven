# PV-002 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in **change** (round 1) and in **later change** (round 2)
- the findings that must be PROVEN:
  1. round 1: with `<template/>` followed by a required frame, change counts the frame and passes, while the reference
     puts the frame into template content, so it is not in the page;
  2. round 2: with `<svg><foreignObject><template/>` followed by a required frame, later change counts the frame and
     passes, while the reference puts the frame into template content.
- evidence required: the check lines for those two inputs; the line for an explicit `<template>` … `</template>`, where
  change and the reference agree that the frame does not count (the inert rule works on the path the fixtures walk);
  the line for `<svg><rect/></svg>`, where later change and the reference agree; the line for a single real frame, where
  all agree; and the shipped fixtures green on both versions, which shows they cannot express either defect.
- what must NOT be claimed: that template inertness or first-duplicate reading does not work — both work on every
  shipped fixture, and the unplug controls really turn red (reviewer A's measurements were accurate in both rounds;
  they were incomplete); that a browser was run.
- how the record settled it: round 1 — the orchestrating agent probed the guard with an explicit template (inert) and
  `<template/>` before a frame (counted), ruled for the reject, and the next round repaired self-closing handling
  (that repair is later change). Round 2 — no ruling was recorded; the reject stood and the work was parked. When this corpus was
  prepared, the guard-side half was re-probed and confirmed; the browser half follows from the standard's rules for
  HTML integration points. The label rests on the standard and those measurements, not on which reviewer rejected.
