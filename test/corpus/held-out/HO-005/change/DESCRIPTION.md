Add a fallback to the tool.

If the requested model is busy, the tool must not fail the job.

The tool serves the answer from another model when the requested one is not in its list, and exits 0. The answer
carries no field that names the model that served it.
