# PV-015 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in **change**
- the finding that must be PROVEN: a plugin that was not scanned clean can flip `scanned_clean` to true in its own folder
  during one run, and **change** then loads it without asking, while the reference says "ask the user". The shipped flip
  control is the same act, made by the test instead of by the plugin.
- evidence required: the check line for the self-flip input (change = load, reference = ask the user); and the lines for
  the clean and not-clean plugins without a flip, showing change and reference agree there (the decision is not simply
  broken).
- what must NOT be claimed: that the flip control proves the decision is safe; or that the controls do not fire — they do
  (reviewer B's measurements of the controls were accurate; its reading of the flip control was not).
- how the record settled it: the orchestrating agent read both reviews and measured one more defect behind the flag. The
  next round made the self-promotion blocking, adopted reviewer B's write-side condition as a second blocking item, and
  noted that on the "one source" point both reviewers were partly wrong. The work was accepted and merged rounds later.
  The label rests on that round, not on which reviewer rejected.
