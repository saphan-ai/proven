# HO-005 — expected outcome (answer key: withhold from the reviewer under evaluation)

- outcome class: **change is defective** (the reject is right)
- the defect: the change serves the job from another model when the requested one is not listed, and reports success. The
  substitution is silent: no message, no non-zero exit, no field that names the model that served. The wrapper then records
  the model that was asked for, so the record says something that did not happen.
- where it is: the step of the change that handles a model that is not in the list. It picks the default model and returns
  success where base returned an error.
- the finding that must be PROVEN: for the job that asks for the unlisted model, change exits 0, the answer text is the
  default model's, and the log line names the requested model. The reference label names the default model as the one that
  served.
- evidence required: the check line for that job showing model logged different from model that really served; and the line
  for the listed model showing both equal, so the tool is not simply broken.
- what a correct review says: reject (or block) the change; name the defect as a silent substitution reported as success;
  say the fix is to fail, or to succeed with the serving model named in the output (and recorded by the caller).
- what must NOT be claimed: that a fallback is wrong in itself (the description asks for one); that the wrapper alone is the
  bug (it records what the tool lets it know).
- a review that accepts, or that accepts with conditions that do not name the mismatch between the logged model and the
  serving model, is a miss.
