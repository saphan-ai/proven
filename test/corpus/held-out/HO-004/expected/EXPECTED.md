# HO-004 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: the change reads every 422 from the create call as "already exists". The API also answers 422 for an input
  it refuses. For a refused email the user is never created, the script goes on, and the team call fails with 404 at a step
  that has no link to the cause. A status code that stands for two causes is read as one.
- where it is: the step of the change that handles the answer to the create call, which branches on the status code alone
  and does not read the body that tells the two reasons apart.
- the finding that must be PROVEN: for the input with the refused email, base stops at the create step and change stops at
  the team step with 404. The reference label is `should fail at create`. For the existing user, change goes on to
  success, as intended.
- evidence required: the check lines for the refused-email input showing different last steps for base and change; and the
  line for the existing user showing the change succeeds, so the fix for re-runs is not simply broken.
- what a correct review says: reject (or block) the change; name the defect as a conflated status; say the fix is to read
  the reason in the 422 body (or to check that the user exists) before going on.
- what must NOT be claimed: that tolerating an existing user is wrong (it is the stated goal); that the 404 is the bug
  (it is the symptom).
- a review that accepts, or that accepts with conditions that do not name the second cause of 422, is a miss.
