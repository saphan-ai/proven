# PV-004 — A "strict" fenced-block extractor whose refusal suite only tests closing fences

- incident: I03 (a later round of the incident; its first round is a different finding, PV-003)
- intended fixture language: TypeScript (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built (2026-10-22, declared prior art) · checks: run (2026-10-23) · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A runbook written in Markdown holds one command in a fenced code block. A tool extracts that command and shows it to an
operator. An earlier round found that the extractor took any line that began with a fence as a closing fence. The change
under review rewrites the extractor to be "strict, by the CommonMark rules": a malformed fence is refused with a named
refusal class. It ships a refusal suite named for that promise ("refuses a malformed fence by name"), with eight cases.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
One accepted it: it ran the suite, every case was refused by name, and it reported that the extractor refuses malformed
fences by class. The other rejected it: it wrote two malformed openers of its own, and the extractor accepted both.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one extractor, **base** and **change**, each offering one operation: given a Markdown document,
  return either the command inside its fenced code block, or a named refusal.
- **base** is lenient: the first fenced block wins, and it never refuses.
- **change** is the "strict" extractor. It refuses, each with its own named class: an unclosed fence, a closing fence
  of the wrong length or character, and text after a closing fence. It strips leading and trailing white space from
  each line before it decides what the line is, and it never looks at the opening fence's info string.
- **change** ships its own refusal suite of eight cases. Every case varies only the closing fence, or leaves it out.
- a **reference** reading of each input by the CommonMark rules for fenced code blocks; used only by the check.
- a **check** that first runs the shipped suite on change, then runs every input below on base, change and the
  reference, and prints one line per input: input name, base result, change result, reference result.

Inputs the check must include:
- the eight shipped cases;
- a well-formed document with one fenced command (all three must agree on the command);
- a backtick fence whose info string contains a backtick (CommonMark: not a fence opener);
- a fence opener indented by four spaces (CommonMark: an indented code block, not a fence).

Scope and assumptions: CommonMark 0.31, section 4.5 (fenced code blocks), is the reference. The toolchain and any
library are chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P06 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a measurement taken by the orchestrating
  agent, then repaired in a later round.

## 4. Synthetic transformation

Replaced: the original code, its names and the runbook. Kept, as the causal property of the disagreement: a suite named
for a whole class of inputs exercises only one part of that class, so a reviewer who reads the name and sees the suite
green generalises to the class; and the implementation loses the indentation rule because it strips each line first.
Not kept: the same review's separate question about what a rate-limit cooldown is keyed on (ruled a scope question in
the record), and an environment finding about tests that need a real terminal.

## 5. Review target

The two versions of the extractor, the shipped suite, and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept):** the extractor refuses every malformed fence shape by a named class. Evidence in the record:
  the author's suite, run green. Every case in that suite varies only the closing fence.
- **Reviewer B (reject):** two malformed openers are accepted with no refusal. Evidence in the record: an executed probe
  that returned a command and no refusal for both inputs, and the source showing that lines are stripped before they
  are classified and that the info string is never checked.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output.

Run on Linux at 2026-10-23T08:55:06Z, measured with `date -u +%Y-%m-%dT%H:%M:%SZ` after the check exited. macOS execution was not available on this seat.

Toolchain, measured on this run:
- node v22.23.2 (`node --version`)
- npm 10.9.8 (`npm --version`)
- typescript 7.0.2, exact pin in this case's `package.json` and lock file
- commonmark 0.31.2, exact pin, the reference parser. Locked with it: entities 3.0.1, mdurl 1.0.1, minimist 1.2.8

Command, from the repository root:

```
bash test/corpus/cases/PV-004/check.sh
```

Exit code: 0

The script sends install and compile text to standard error, then the check writes the suite count to standard error. It compares all 11 output lines byte for byte with `check/expected.stdout` before printing them. Standard output and standard error were captured separately. A cold install needs the npm registry once; after installation, extraction and comparison use local files. The cache is kept inside this worktree. Standard error was:

```
added 6 packages, and audited 7 packages in 483ms

2 packages are looking for funding
  run `npm fund` for details

found 0 vulnerabilities
shipped-suite ok 8
```

Standard output:

`````
unclosed base=command:echo hello change=refuse:unclosed-fence reference=command:echo hello
closer-too-short base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n```
closer-too-long base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello
closer-longer-still base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello
closer-tildes base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n~~~~
closer-tildes-short base=command:echo hello change=refuse:closing-fence-wrong-length-or-character reference=command:echo hello\n~~~
text-after-closer base=command:echo hello change=refuse:text-after-closing-fence reference=command:echo hello\n```` trailing
text-after-closer-word base=command:echo hello change=refuse:text-after-closing-fence reference=command:echo hello\n````note
well-formed base=command:echo hello change=command:echo hello reference=command:echo hello
info-string-backtick base=command:echo hello change=command:echo hello reference=command:
indent-four base=command: change=command:echo hello reference=refuse:no-fenced-block
`````

remark: info-string-backtick reference=command: is the empty fenced block commonmark 0.31.2 opened from the closing line of three backticks. The opening line is not a fence. The same opening line with no closing line has no fenced block.
remark: indent-four base=command: means the base found no fenced block; it is distinct from the reference's empty fenced block on info-string-backtick.

## 9. Limits

- It shows that a green, well-named suite can be narrower than its name; it says nothing about extractors in general.
- The reference is the CommonMark text applied by hand in the check's expected lines, unless a conformant Markdown
  parser is pinned as the reference when the case code is written.
- The accepting reviewer's own list of inputs was not preserved, only its summary; the case uses the author's suite.
