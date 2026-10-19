# PV-004 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **established defect** in `change/`
- the findings that must be PROVEN: `extractCommand` in `change/` returns a command, with no refusal, for (1) a backtick
  fence whose info string contains a backtick and (2) a fence opener indented four spaces. CommonMark treats neither as
  a fence opener.
- evidence required: the check lines for both inputs (change = command returned), plus the shipped suite green, which
  shows that the suite does not cover openers.
- what must NOT be claimed: that the shipped refusal classes do not work — they refuse every closing-fence case.
  Reviewer B's run was accurate; its generalisation from the suite's name was not.
- how the record settled it: the orchestrating agent ran both inputs and both were accepted; a later round added named
  refusals for openers. The label rests on that measurement and the specification.
