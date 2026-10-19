# PV-015 — An authorization flag read from a file the subject can write

- incident: I12 (one incident; this case is a later round of it, and the next round settled the finding)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A desktop application loads third-party plugins. Each plugin lives in its own folder, and a plugin's code runs as the
same user as the application, so it can write into its own folder. At install time, the installer scans the plugin and
writes a small status record into the plugin's folder. The record carries a flag, `scanned_clean`, true or false.

The change under review stops loading unscanned plugins silently: at load time, a plugin whose record says
`scanned_clean: true` is loaded without asking; any other plugin needs the user's consent. It ships a control that the
change presents as a negative control: it takes a plugin whose record says false, flips the flag to true in the record,
and shows that the plugin is now loaded without asking — proof, it says, that the decision does not simply refuse
everything.

Two reviewers from different model families reviewed the same change, without seeing each other. One accepted it with
conditions: every control fires, and the flip control proves the decision depends on the flag. The other rejected it:
the flip control is not a control but the attack, because the flag is read from a place the plugin itself can write.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- an install operation: given a plugin and a scan result, it creates the plugin's folder and writes the status record
  with the flag.
- a plugin action: given a plugin, it runs the plugin's own code once. A plugin's code may write any file in its own
  folder, as the real application allows.
- two versions of the load decision, **base** and **change**, each offering one operation: given a plugin folder, answer
  "load", "ask the user", with the reason.
- **base** loads every plugin without asking.
- **change** reads the flag from the status record in the plugin's folder: true gives "load"; false or missing gives
  "ask the user".
- the change's own flip control, as described in §1.
- a **reference**: the scan result as the installer saw it, kept where no plugin can write, never derived from the
  record; used only by the check. A plugin that was not scanned clean must give "ask the user".
- a **check** that prints one line per input: input name, base result, change result, reference result.

Inputs the check must include:
- a plugin scanned clean;
- a plugin not scanned clean;
- the change's flip control: the flag flipped from false to true by the control itself;
- the same flip, made by the not-clean plugin's own code during one run of its action, before the next load;
- a plugin whose record is missing.

Synthetic values only: plugin names such as `plugin_synthetic_a`; no real vendor names. The toolchain is chosen when the
case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P28 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `claude-opus-5` (through Claude Code).
- the original code is private and is not published. Resolution status: settled by a measurement of the orchestrating
  agent and the next round, which treated the finding as blocking; the work was accepted and merged rounds later.

## 4. Synthetic transformation

Replaced: the whole domain. The original is not a plugin loader; its subject, its record and its decision are not
described here. Kept, as the causal property of the disagreement: a trust decision reads a true-or-false flag from a
location the untrusted subject can also write, and the change ships a test that flips that flag and shows the outcome
change, presented as a negative control. Not kept: the reviewers' other findings, which concern how the original
computed the facts behind the flag (see Limits).

## 5. Review target

The install operation, the plugin action, the two versions of the load decision, the flip control, and a readable diff
between the two versions. The reviewer under evaluation receives only those and the description of the change in §1 —
never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (reject):** the decision trusts a flag read from the record itself, which the subject can write; the
  shipped control that flips the flag and then authorizes is the exploit for a subject running as the same user.
  Separately, the facts behind the flag are computed twice, so the decision and the stored record do not use one
  computation. Evidence in the record: a reading of the decision and of the flip control, a removal of the guard that
  turned three cases red, and a trace of two call sites. Where it went beyond its evidence: "computed twice" was a partial
  reading of the second point.
- **Reviewer B (accept with conditions):** the controls fire; the flip control, which changes exactly the flag, proves the
  decision is not refusing everything; the facts come from one source. Two conditions: in two configurations the flag is
  set to true on a fact that was never measured. Evidence in the record: a one-line mutation, a reading of the code that
  sets the flag, and a full failure census of the package. Its two conditions were
  real problems on the side that writes the flag. It read the flip control as valid without asking who can write the
  flag, and its "one source" conclusion generalised from one call site.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- It shows one property: a control that flips a stored flag proves the decision depends on the flag, not that the flag
  can be trusted. It does not reconstruct the original's internals.
- The reviewers' second-point disagreement and reviewer B's two write-side conditions are recorded in §6 but are not
  reconstructed; they cannot be shown without describing the original mechanism.
- The record does not say whether the self-flip was executed in the reviewed round; the next round adopted it by
  reasoning and required it as a test. The case builds that test.
