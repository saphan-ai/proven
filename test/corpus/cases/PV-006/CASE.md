# PV-006 — A placeholder check that enumerates shapes instead of refusing the class

- incident: I05 (one incident; this case is a later review round of it, and its only pair in the corpus)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-22 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A web service shows a "please wait" notice when a visitor sends too many requests. The notice text comes from a message
catalog, and the product fills a placeholder such as `{wait}` with a duration. The same answer also carries a
retry-after value, in seconds. Earlier rounds had let a raw placeholder reach the visitor, and each repair added a
control for the shape that had escaped. The change under review is test-only. It rebuilds the controls: the text
control now uses the product's own placeholder pattern (a token in braces), and a new control compares the duration
shown with the retry-after value.

Two reviewers from different model families reviewed the same change, without seeing each other. Their orders differed.
One was told to re-run the earlier attacks and probe the new controls. It found two narrow holes in the controls and
accepted with conditions. The other was told to find the next raw token that reaches a visitor. It rejected the change.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- one **product** renderer, the same in both versions: given a catalog entry and a wait time, it returns the notice
  text and the retry-after value. It fills only tokens in braces; any other text is copied as it is.
- two versions of the controls, **base** and **change**, each offering two operations:
  1. given a rendered notice, answer pass or fail, and name any raw placeholder found;
  2. given the duration shown in the notice and the retry-after value, answer pass or fail.
- **base**: the text control refuses only the placeholder names it lists, each written in braces. It has no duration
  control (it always passes).
- **change**: the text control refuses any token in braces (an opening brace, one or more characters that are not
  braces, a closing brace), the same pattern the product uses to fill. The duration control passes when the duration
  shown is equal to the retry-after value or up to two seconds short of it.
- a **reference**: for each input, the exact notice text a visitor must see, written by hand, and an exact equality
  test for the duration. It is never derived from the controls and is used only by the check.
- a **check** that runs every input below against base, change and the reference, and prints one line per input:
  input name, base result, change result, reference result.

Inputs the check must include:
- a well-formed catalog entry, fully filled — all three must agree;
- the earlier rounds' attacks: an unfilled token in braces whose name the base list holds, and one whose name it does
  not;
- the catalog entry with an extra token of another shape appended: a dollar sign and a name, with no braces (for
  example `$wait`);
- a shown duration of 0 seconds against a retry-after value of 1 second;
- a shown duration equal to the retry-after value.

Scope and assumptions: plain text only, one language, no browser. The toolchain is chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P10 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a ruling of the orchestrating agent
  after three defeated rounds, and by the next round, which replaced the enumerated controls.

## 4. Synthetic transformation

Replaced: the page, its text, its catalog and its names. Kept, as the causal property of the disagreement: a control
that refuses placeholders by an enumerated pattern (tokens in braces); a catalog string that gains a token of another
shape, which the product shows raw while every control stays green; and a duration control that accepts zero against
one. Both reviewers found real ways past the controls; they differed on whether those ways block the change.
Not kept: a source scan that counted the product's fill calls, an attribute-quoting hole and a decoy-anchor hole in two
other controls, and a token in a visible HTML attribute (see Limits).

## 5. Review target

The product renderer, the two versions of the controls, and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check, the reference or
`expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the rebuilt controls are different in kind and catch every earlier attack.
  Two other predicates are narrower than they claim, and the source scan misses some call shapes; these are conditions,
  not a reject. Evidence in the record: the earlier attacks re-run, all now red; two planted inputs that beat the two
  narrow predicates with the suite green; several call shapes the scan did not see. It did not try a token outside the
  brace pattern.
- **Reviewer B (reject):** a raw token of another shape reaches visitors while every control stays green; the source
  scan can be blinded; and the duration control accepts a shortfall of up to two seconds, zero against one included.
  Evidence in the record: a `$`-token appended to the catalog, with the text control passing while its log showed the
  raw token in the notice; a scan bypass; a temporary control showing the zero-against-one case passes. All mutations
  were restored. One arm of the claim, a second language, was later shown moot: only one language was served.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output.

- toolchain: Go, standard library only, `go version go1.26.5 linux/amd64`. The case is its own module (`go.mod` in
  this directory, module `pv006`), so it stays out of the repository's module.
- run: 2026-10-22T15:35:43Z (UTC)
- command, from the repository root: `bash tests/corpus/cases/PV-006/check.sh` (it runs `go -C tests/corpus/cases/PV-006 run ./check`)
- exit code: 0

```
well-formed base=pass change=pass reference=pass
listed-token base=fail({limit}) change=fail({limit}) reference=fail
unlisted-token base=pass change=fail({window}) reference=fail
dollar-token base=pass change=pass reference=fail
duration-0-vs-1 base=pass change=pass reference=fail
duration-equal base=pass change=pass reference=pass
```

Each line is: input name, base result, change result, reference result. A failing text result names the raw
placeholder found. The inputs are one catalog entry, `Too many requests. Please wait {wait} before trying again.`, with
a wait of 30 seconds, plus a suffix per input (` Limit: {limit}.`, ` Window: {window}.`, ` $wait`).

## 9. Limits

- It shows one property of one kind of control: an enumerated pattern cannot refuse a class it does not name. It does
  not show that every enumerated control fails.
- The original review also found holes in two other controls, blind spots in a source scan, and a token in a visible
  attribute. Those are not reconstructed here.
- That the enumerated approach could not be patched once more is the orchestrating agent's inference from three
  defeated rounds, not a measurement; the case does not test it.
- The base version is constructed: the record does not say what the control looked like before the change, so base is a
  plain stand-in that keeps the property under test.
