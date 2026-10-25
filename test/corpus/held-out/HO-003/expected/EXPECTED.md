# HO-003 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: the change adds the read bits but never clears the group-write bit. The modes it produces depend on the
  umask of the process that runs the copy. Under umask `022` the group-write bit is already off and the result is as
  promised. Under umask `002` the group-write bit is on and stays on.
- where it is: the step of the change that sets the mode. It ORs the read bits onto whatever mode the creation call left,
  and does not set the mode to a fixed value or mask the unwanted bits.
- the finding that must be PROVEN: the same source tree, copied under umask `002`, gives entries with the group-write bit
  on; under umask `022` it does not. The change's own test passes under `022` and fails under `002`.
- evidence required: the check lines for both umasks showing different change modes for the same entry, and a reference
  mode that does not change with the umask; and the base lines showing the same mode under both umasks.
- what a correct review says: reject (or block) the change; name the defect as a result that depends on the environment's
  umask; say the fix is to set the mode explicitly (or to mask out the group-write and other-write bits) and to run the
  test under both umasks.
- what must NOT be claimed: that the test is wrong when it passes under `022` (it is a true result for that umask); that
  the read bits are wrong.
- a review that accepts because the test passes, without naming the umask, is a miss.
