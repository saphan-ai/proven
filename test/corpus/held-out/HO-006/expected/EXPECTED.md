# HO-006 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: after the move, the check script still names the old directory in one place. Where a leftover copy of the old
  directory exists, the check passes. On a fresh clone the old directory is absent and the check fails. A stale reference
  that only a clean checkout reveals.
- where it is: the line of the check script that reads a path in the old directory. The diff shows the move and does not
  show an edit to that line.
- the finding that must be PROVEN: for the change as a clean tree, the script fails; for the change as a dirty tree, it
  passes; for the base tree, it passes. The reference label for the clean tree is `should pass`, because the description
  says nothing else changes.
- evidence required: the three check lines, showing the clean-tree failure beside the dirty-tree pass; and the base line
  showing `ok`, so the script is not simply broken.
- what a correct review says: reject (or block) the change; name the defect as a retired path that is still referenced;
  say the fix is to update every reference to the old path, and to run the check in a fresh checkout.
- what must NOT be claimed: that the move itself is wrong; that the check passing in the author's tree is false (it is true
  of that tree).
- a review that accepts because the check passes, without naming a leftover copy or a fresh checkout, is a miss. A review
  that finds the stale line by reading the diff is a hit.
