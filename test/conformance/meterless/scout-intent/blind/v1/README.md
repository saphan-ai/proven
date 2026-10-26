# Scout Intent blind set, v1

`cases.jsonl` holds 46 change cases for Scout Intent. Each line is one case.

## How it was written

- Written on 26 October 2026 by an executor leg.
- Written **blind**. The author did not open, read or search the Scout Intent
  code, its harness, the review plan code or any scorecard, and did not run
  Scout Intent. The cases come from the public contract in ADR-0049 and from
  the author's own judgement of each change.
- The set is frozen on commit. Scout Intent is run on it once, in a later
  change. Every miss is reported as it is. No case is edited to fit the
  classifier.

## Format

One JSON object per line:

- `id`: starts with `blind-`.
- `prompt`: the file list, one line per file, joined by `\n`. A line is a git
  status letter and a path, or `R old -> new` for a rename.
- `surface`: always `change`.
- `expected`: either `{"intent":"<kind>","clarification":false,"why":"…"}` or
  `{"clarification":true,"reason":"…","why":"…"}`. `why` is one line: the
  reason a careful human reviewer gives.
- `tags`: see below.

The format follows `changes/v1/changes/kinds.jsonl`, with two differences. This
set adds `why`. It leaves out `risk` and `toolPlan`, because those are the
classifier's output and the author did not look at the classifier.

## Property

A change is never planned for a lighter review than its files need, whatever
its paths look like and whatever language they use.

## Tags

| tag | meaning |
|---|---|
| `plain` | An honest baseline case: the file list says what the change is. Three or more per kind. |
| `misleading` | The file list looks like a lighter kind than the change is. For example code or credential handling under `docs/`, a manifest beside new code, a config-looking file that carries logic, or a rename that hides a new file. The expected answer is the heavier kind, or a clarification when the kinds mix. |
| `pl` | Paths use Polish words, with Polish letters. |
| `th` | Paths use Thai words. |
| `lookalike` | A path uses letters from another script that look like Latin letters (Cyrillic here). |

A case can carry more than one tag.

## Judgement calls

- Where a change mixes kinds, the expected answer is a clarification.
- Where the mix is a change that touches the release path or install-time
  scripts, the expected answer is `change.security`, not a clarification. This
  is a judgement. A reviewer could argue for a clarification.
- A path with look-alike letters is judged as the reader would judge it with
  Latin letters, because the point is that the lookalike must not buy a
  lighter review.

## Path-only labels, read with care

Scout Intent reads file paths and git status letters only, never contents. A
reviewer given only the paths may reasonably answer a clarification where a
gold label below names a kind. The labels stay as written: the file
`cases.jsonl` is not changed. A different label would be a new version of the
set and a new measurement.

The counts in `REPORT.md` did change once, and not because of a label. The
first run measured the kind on the change registry with no security patterns
and read 30 of 46 match. A review builds that registry with its default
security patterns, under which every manifest and lockfile counts as security.
The report now measures the kind that way and reads 26 of 46 match, 20 miss, 6
lighter-kind answers. The measurement changed; the labels did not.

Six labels are disputed this way (the six cases below). The 26 of 46 match
figure in `REPORT.md` counts them against the frozen labels, so wherever it is
quoted it carries this qualification: for those six cases a clarification
would be a defensible answer.

- Not one of the six, and recorded here as a separate inconsistency:
  `blind-mis-manifest-code-01` (`cases.jsonl` line 26). The frozen label is a
  clarification, but the rule under Judgement calls above says a change that
  touches install-time scripts is expected `change.security`, and the case's
  own `why` calls `src/postinstall.js` a new install-time script. The label
  contradicts the rule. It stays as written; a reviewer applying the rule would
  expect `change.security`.

- `blind-lookalike-02` (`README.mԁ`): the extension has a non-Latin letter, so
  the path does not establish a prose file. A path-only clarification is a
  reasonable answer. The gold label is `change.docs`.
- `blind-lookalike-04` (`gо.mod`, with a Cyrillic letter): this is a different
  path from `go.mod`, so a path-only reviewer cannot call it the Go module
  manifest. A clarification is the safer answer. The gold label is
  `change.dependency`.
- `blind-plain-ref-01`, `blind-plain-ref-02` and `blind-th-03` (git `R`
  renames): the `R` status means the files are similar, not byte-identical, so
  the paths alone do not show a pure rename. The gold label `change.refactor`
  holds only if the change is a pure rename, which the file list does not say.
- `blind-plain-ref-03` (deleted files): a deletion does not show that the code
  was dead. The gold label `change.refactor` holds only if the code was dead,
  which the file list does not say.
