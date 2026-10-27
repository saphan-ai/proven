# Review output

Text, JSON, SARIF, Markdown and HTML are five writings of one review result (ADR-0020 §1, ADR-0033 point 8). The chosen format (`--text`, `--json`, `--sarif`, `--md`, `--html`, or `PROVEN_OUTPUT_FORMAT` set to `text`, `json`, `sarif`, `md` or `html`) goes to standard output, or to `-o`. Each `--text-output`, `--json-output`, `--sarif-output`, `--md-output` and `--html-output` file is also written in that same run, and `--md-output` and `--html-output` take the same output-path checks as the others. Empty lists are `[]`, not null.

Two outputs of one run must be two files. If `-o` and any `--*-output`, or two `--*-output` flags, name the same file, the run is refused at load with exit 7. The refusal names both flags and the path, and nothing is written. Each path is resolved the way the filesystem resolves it: component by component, each symbolic link followed as it is met and only then any `..` applied, never cleaned first (cleaning `link/..` first could name the wrong file); a path that passes through a directory that does not exist cannot be written and is left to fail at the write with exit 2. So `-o real/r.json` with `--sarif-output link/r.json` (where `link` points to `real`) is refused. Two existing files that are the same file (a hard link) are refused too. Two spellings of a file that does not exist yet and differ only in letter case, on a case-insensitive volume, are not refused at load: the second write finds the file the first one made, the run stops with exit 2 and names both flags, and the first output is left as it was written.

An output path whose final component is a symbolic link is refused at load with exit 7. The refusal names the flag and the path, and nothing is written. The link is not followed: the file it names is left as it was. A regular file at the path is replaced: the report is a new file renamed over it, so it gets a new inode and mode `0644` less the umask, and a hard link to the old file keeps the old content.

Any earlier component of that path that lies inside the working directory is refused at load with exit 7 when it is a symbolic link. The refusal names the flag, the path and the linked component, and nothing is written. The working directory and the prefix already walked are resolved before that comparison, so an absolute path that reaches the checkout through a different spelling is the same check. In `proven ci` the working directory is the checkout, so a directory of a relative output path that stays in the checkout, or of an absolute path that resolves to it, cannot be a link that chooses another directory. A relative path whose `..` climbs out of the working directory, even after passing through a directory inside it (`reports/../../x`), is outside it: it is the operator's, and its directories are followed at the write. A component that lies outside the working directory is the operator's choice and is not refused here. A symbolic link that is a directory on the way is still resolved when two outputs are checked for being one file. That check runs first and still exits 7, naming both flags.

A path inside the working directory is written through its directory, opened at load after these checks and before the review runs. A later rename of that directory, or a link put at its name (to a directory inside the checkout or outside it), does not change where the report goes: it is written in the directory the check saw, and if that directory was removed the write fails with exit 2 and names the path. A directory that did not exist when the run started is not looked up again at the write: that output fails with exit 2 and names the path, because the review could have made it a link. The report is a new file in that directory, and that file is renamed over the final name: the final name is replaced, never written through. A symbolic link swapped in as that name after the check is replaced, and the file it names is not written. A directory component swapped for a link that leaves the directory still cannot receive the report: the write fails with exit 2 and names the path, and nothing outside is written. A path outside the working directory is not opened through that root: a directory link on it is still followed, including one swapped in after the load check, and where the system has no open that refuses a final symbolic link, a final link on that path swapped in between the check and the open can still be followed. A final link that is seen at the write is refused on every path, and that refusal is exit 2. Where the system can open a file without following a final symbolic link (`O_NOFOLLOW`), that open is what refuses a final link on a path outside the working directory.

DEL (U+007F) and C1 (U+0080 to U+009F) in JSON and SARIF are written as `\u007f` and `\u0080` to `\u009f`. The file stays valid JSON and decodes to the same text. The JSON encoder leaves those runes raw, so both writers escape them after encoding.

## JSON

One object. Every key below is always present, except `findings[].evidence_record_url`, which is described after the table.

| field | meaning |
|---|---|
| `class` | the review class the path rules chose |
| `class_rule` | the path pattern that chose the class, or `fallback` |
| `doubled` | the class is one of the doubled classes |
| `partial` | the deadline was reached |
| `settings_statement` | set when the change edits `proven.yaml`; empty otherwise |
| `author_model_statement` | `author model not stated` when the change names no model; empty otherwise |
| `reviewer_statements` | one line per reviewer, plus a line when the special model is not configured |
| `author_models` | the models the change named; each has `agent`, `model`, `family`, `source` |
| `default_reviewer` | the configured default reviewer |
| `special_reviewer` | the configured special-review model; empty when not configured |
| `reviewer_models` | the models this review used, the default reviewer first |
| `trace_path` | the path the review wrote its trace to; it may be absolute |
| `family_sentence` | the same-family line the trace recorded; empty when it has none or cannot be read |
| `models` | every model the review used; each has `agent`, `route`, `model`, `receipt_id`, `from_receipt` |
| `failed_services` | the services that failed; each has `service` and `reason` |
| `proof_counts` | how many findings are at each level; every level is a key, with 0 when none: `proved-by-test`, `proved-by-tool`, `proved-by-data`, `agreed-by-two-models-not-proven` |
| `guardrail` | what NVIDIA's guardrail did in this review, measured per review and never typed; an object, see **The guardrail line** below |
| `findings` | every reported finding |
| `findings[].id` | finding id |
| `findings[].proof_level` | that finding's level of proof, one of the four keys of `proof_counts` |
| `findings[].rule_id` | kind of finding |
| `findings[].title` | one line naming the finding |
| `findings[].message` | what is wrong and why it matters |
| `findings[].severity` | `INFO`, `WARNING` or `ERROR` |
| `findings[].unresolved` | the evidence did not settle it |
| `findings[].models` | the reviewer models that reported this finding, the default reviewer first |
| `findings[].location` | where it is: `path`, `start_line`, `end_line` |
| `findings[].evidence[].kind` | `test` or `query` |
| `findings[].evidence[].test` | `path`, `name`, `command`, `fails_on_change`, `passes_on_base`, `output` |
| `findings[].evidence[].query` | `sql`, `engine`, `snapshot`, `columns`, `rows` |

Of one evidence entry, only the object named by its `kind` is present.

`findings[].evidence_record_url` is the one key that is not always present. It is the address of the finding's stored evidence record in the viewer page (docs/ADAPTERS.md, Evidence records and their viewer), and it is left out when the finding has none: records are not set, the record was not stored, or no link base address is set. The text and SARIF outputs do not carry it.

## SARIF 2.1.0

One log, `version` `2.1.0`, one run. The tool driver `name` is `proven`. One result per finding. `level` is `INFO`→`note`, `WARNING`→`warning`, `ERROR`→`error`; any other severity is `note`. `ruleId` is `rule_id`, or `id` when the finding has no rule. The message is the finding's message, then its title, then its id. A finding with no path has no location; a start line below 1 gives no region; an end line before the start line is left out.

The artifact `uri` is a relative path inside the repository: no leading `/`, no authority, and no `..` segment. The path is cleaned first, so `dir/../a.go` is `a.go`. A path that is absolute, names an authority such as `//example.invalid/a.go`, or still climbs out after cleaning, is omitted. The finding is kept with no location. This log does not set `uriBaseId`.

Each result's `properties` carries `id`, `rule_id`, `title`, `severity`, `unresolved` and `evidence`. `evidence` is the same array as in the JSON.

The run's `properties` carries `class`, `class_rule`, `partial`, `failed_services`, `settings_statement`, `author_model_statement`, `reviewer_statements`, `trace_path`, `family_sentence` and `guardrail`. Each has the same value as the JSON key of that name, so the lines the text output prints beside the findings are in a SARIF file too.

Each result's `properties` also carries `proof_level`, and the run's `properties` carries `proof_counts`. Both are the same values as in the JSON.

### The guardrail line

Every output states what NVIDIA's guardrail (ADR-0005) did in this review, in the same words (the reviewer lines name the reviewing models; this line names the guardrail). The counts are measured at the one place every check passes through, the guardrail callback of the pipeline, and a doubled review adds both runs' checks:

```
Guardrail: [<maker> ]<model> (<where>), checked <P> prompts and <A> answers, <B> blocked
```

`<maker>` is `NVIDIA` when the configuration says the model is NVIDIA's: NVIDIA's hosted API, or one of the NVIDIA content-safety models Proven ships as its defaults, on NVIDIA's API or the operator's own server. It is left out for any other model; a maker is never guessed from a model id. `<where>` is `NVIDIA hosted API`, `self-hosted` for the operator's model server, or the backend's own name. A count of one takes the singular (`1 prompt`, `1 answer`). When a check could not be made (the guardrail did not answer), the line ends `, <U> not checked`; such a check is in neither `<P>` nor `<A>`. Two other lines stand in its place: `Guardrail: off for this review` when the guardrail was switched off, and `Guardrail: no model call to check` when it was on and no model call was made (a sample run). `<B>` counts the checks whose verdict was unsafe, each of which stopped its call.

- text: the line is in the footer, before the trace path.
- Markdown: the line ends the foot, after the reviewing models (`<sub>reviewed by … · Guardrail: …</sub>`).
- HTML: the line ends the foot, after the reviewing models (`<footer>reviewed by … · Guardrail: …</footer>`).
- JSON: the `guardrail` object has `state` (`checked`, `off` or `no_call`), `line` (the words above), `prompts`, `answers`, `blocked`, `unanswered`, `backend`, `model` and `maker`.
- SARIF: the run's `properties` carries the same `guardrail` object.

Properties: `TestGuardrailLineIsTheSameInTextAndMarkdown`, `TestEveryFormatSaysWhenTheGuardrailWasOff`, `TestGuardrailLineSaysWhenThereWasNoModelCall`, `TestDraftCountsTheGuardrailChecksItMade`, `TestDraftWithTheGuardrailOffCountsNothing`, `TestHTMLFootCarriesTheGuardrailLine` (the HTML foot). `TestGuardrailLineIsTheSameInTextAndMarkdown`, `TestEveryFormatSaysWhenTheGuardrailWasOff` and `TestEveryFormatNamesNVIDIAForTheDefaultSelfHostedGuardrail` also run over `html`.

## Markdown

The pull-request comment (`--md`, `--md-output`, `PROVEN_OUTPUT_FORMAT=md`). It is written from the same review result as the JSON and SARIF, in this order, top to bottom:

1. **The verdict line.** `**Proven: blocked** — N findings: E error, W warning, I info`, `**Proven: clean** — 0 findings`, or, when there are no findings and the model reviewer did not answer (the `nightshift` or `nvidia` service failed), `**Proven: not reviewed** — the model reviewer did not answer (`nvidia`), so 0 findings proves nothing`: with the model down no candidate is ever raised, so the run is not called clean (`TestMarkdownDoesNotCallAReviewCleanWhenNoModelAnswered`). When the model answered some calls and not others, the heading is `**Proven: incomplete** — … did not answer every call, so 0 findings proves nothing` (`TestMarkdownCallsAReviewIncompleteWhenTheModelAnsweredOnlySomeCalls`). A count follows in brackets when it is not zero: findings left unresolved, findings an exception accepts, and opinions agreed by two models but not proven. The result carries no count of dropped claims (a claim only one model makes is not in it), so none is printed.
2. **A clean run** then has one sentence saying what was checked: the review class in words and why it was chosen (the path rule, or that no rule matched), how many tracked paths a live review skipped (symlinks) when there were any, whether every service answered (the ones that did not, by name), and that a finding needs proof to be reported, or, when no model answered, that no model read the change and the review should run again. The result carries no list of the files or the commit range reviewed, so none is printed. No table and no `<details>`.
3. **A run with findings** has the findings table (severity as a word, finding, `file:line`; the rule id stays in JSON and SARIF) and one `<details>` block per finding, in the order of the findings. The summary line is the severity, the finding's title and its place; the finding's internal id stays in JSON. The block holds the message and the evidence in the order the finding carries it: a query (the SQL in a fenced block, then the rows as a table, at most 20 rows when the comment is not too large, see below), a failing test (name, path, whether it fails on the change and passes on the base, the command and its output in fenced blocks), a tool's run on the change and on the base, and the fix when there is one (a patch only when it was proved, as the text output does).
4. **Notes**, one line each and only when there is something to say: the policy and settings statements; the author model (or that the change does not declare one) and each reviewer with its model and family, written from the review's structured fields instead of its internal role names (the owner's rule of 27 October: no raw field name on a page a judge reads, `TestMarkdownCarriesNoInternalNames`); every service that did not answer, by the name a reader knows it and the reason (a setting left empty reads "not configured for this run"), clean run or not; a partial review; an accepted finding with its owner and expiry; an exception that expired.
5. **The foot**, one line: the reviewing models, then the guardrail line (see **The guardrail line**): the model that checked, how many prompts and answers it checked and blocked, or that the guardrail was off.

**The size bound.** A comment always ends under 60,000 bytes (the platforms' limit is 65,536) and never drops all proof. When the full comment would pass it, each finding keeps its evidence cut to a share of the budget (fewer rows, shorter test and tool output), and its block says `cut: N more rows and M bytes not shown; the full evidence is in the JSON and SARIF files of this run`. The last level, the least proof, is bounded whatever the finding holds: its first piece of evidence only, with names, columns and cells shortened and one row, no fix, and the line `cut: N more pieces of evidence and fix not shown`. If even the least proof of every finding does not fit, the first findings keep their `<details>` and the rest are table rows only, with a line giving how many; a longer table is never taken at the price of every `<details>`. If the table itself does not fit, its first rows are shown, with the count of the rest. Messages, titles and note lines are bounded too. The JSON and SARIF files of the run always hold everything. Properties: `TestMarkdownKeepsProofOfEveryFindingWhenLarge`, `TestMarkdownListsTheRestByRowWhenProofCannotFit`, `TestMarkdownFitsWhenTooLargeEvenWithoutEvidence`, `TestMarkdownNeverTradesAllProofForALongerTable`, `TestMarkdownKeepsTheLeastProofOfAFindingWithManyPieces`.

The comment never starts with the poster's marker: `<!-- proven-review run=N sha=H -->` is written by the workflow or the Action, which know the run and the head commit (see [The judges' lab](LAB.md#one-comment-per-pull-request) and [The comment](ACTION.md#the-comment)). The two goldens are the comments themselves: `testdata/golden/pr-comment/clean.md` and `testdata/golden/pr-comment/findings.md` (the sample run on `testdata/fixture/discount`). `go test ./internal/cli -run TestMarkdownGolden -update-golden` rewrites them; the diff is then read like code.

### What is escaped, and how

Every value that comes from the change (a title, a path, a rule, a message, a model name, a query, a row, a test's output, a fix) is escaped by one rule, written once in `internal/cli/format_md.go` above `mdText` (ADR-0027). A value cannot close a table cell, add a table row, open a tag or a comment, end a `<details>`, break out of a fence, or start a link, an image, a heading, a list, a mention or an issue reference. In short:

- prose and table cells: whitespace folds to one space; control and bidirectional characters show as escapes, as in the text output; Markdown punctuation is backslash-escaped, and a zero-width space (`&#8203;`) follows each escaped `@` and `#`, because a renderer reads the backslash away and then links a mention or an issue reference in the text that is left (`TestMarkdownMentionCheckBitesWhenTheZeroWidthIsRemoved` removes it and shows the check turn red); `<`, `>`, `&` and quotes become HTML entities;
- a path, a name or a command is a code span whose fence is longer than any backtick run inside it; in a table cell a `|` inside it is written `\|`, elsewhere it is written as it is;
- the `<summary>` line is raw HTML, where backslashes do nothing: it takes entities, and a zero-width space (`&#8203;`) where a mention, a reference or a link would start;
- queries and outputs sit in a fenced block whose fence is longer than any backtick run inside it.

`TestMarkdownEscapesEveryValueFromTheChange` puts one hostile string in every field and reads the result as a renderer would; `TestMarkdownEscapingCheckBitesOnOneUnescapedField` writes one field (the title) unescaped and shows that the same check turns red. Not measured: how GitHub's and Gitea's own renderers display the goldens (this build had no network to a live server; the Markdown follows CommonMark and GFM).

## HTML

The report a judge opens (`--html`, `--html-output`, `PROVEN_OUTPUT_FORMAT=html`): one UTF-8 file, written from the same review result, in the words of the Markdown comment (`internal/cli/format_html.go` shares its sentences with `format_md.go`). Top to bottom:

1. **The verdict.** The same line as Markdown (`Proven: blocked — N findings: …`, with a space before the dash), in a box coloured by what it says: blocked, clean, or not reviewed when the model reviewer did not answer. When the model answered some calls and not others the heading is `Proven: incomplete`, as in Markdown (`TestHTMLCarriesTheIncompleteHeading`).
2. **The summary.** With findings, a table of severity (as a word), finding and `file:line`; each finding links to its card by an in-page anchor. Without findings, the statement of what was checked.
3. **One card per finding.** The severity as a word, the title, the place, the message, then **How we know**: one sentence per piece of evidence, in words, from the facts the evidence records (for example "A test fails on the change and passes on the base: `TestDiscounted`"). Below it a closed `<details>` element, **Evidence and fix**, which the reader opens on demand: each query with its SQL and its rows as a table, each test or tool run with its command and output in a `<pre>`, and the fix with the same sentence as the Markdown comment.
4. **Agreed by two models, not proven.** Each opinion (`Result.Opinions`) on one line, and the sentence that a claim only one model makes is not kept. The result holds no list of dropped claims, so none is printed.
5. **The notes and the foot.** The notes: the policy and settings statements, the author and reviewer lines, every service that did not answer by the name a reader knows it, a partial review, accepted and expired exceptions. The foot, one line: the models that reviewed, then the guardrail line (see **The guardrail line**: the checks counted, or that the guardrail was off). No raw field or enum name stands on the page (the owner's rule of 27 October, `TestHTMLOpinionsAndNotesAreListed`).

**One self-contained file.** The CSS is inline in one `<style>`. The page has no script, no `<link>`, no `src=`, no `href=` except an in-page `#anchor` that Proven numbers itself, and no `http://` or `https://` address: an address in a value from the change is printed as text with a zero-width space after its `:`, so it is neither a link nor a fetch. `TestHTMLSampleRunIsSelfContained` reads the sample run for each of these; `TestHTMLSelfContainedCheckBitesOnOneExternalLink` adds one external stylesheet `<link>` and shows the same check turn red.

**Escaping.** Every value that comes from the change, a model or a service goes through one of three helpers, written once above `htmlCode` in `internal/cli/format_html.go` (ADR-0027): `htmlText` for prose, cells and headings, `htmlCode` for a path, name or command, `htmlPre` for a `<pre>` body. Whitespace folds (a `<pre>` keeps its lines), control characters show as escapes, and `<`, `>`, `&` and both quotes become entities. No value is put in an attribute: the page's attributes are fixed words and the `#finding-N` anchors. `TestHTMLEscapesEveryValueFromTheChange` puts `<script>`, `"><img`, an ampersand and an `https://` address in every field and reads the page for any element, attribute, bare ampersand or address it did not make; `TestHTMLEscapingCheckBitesOnOneUnescapedField` writes one field (the title) raw and shows that check turn red.

**Not cut.** The file has no platform size limit, so the writer applies none of the Markdown comment's bounds: a fix explanation (2,000 bytes in Markdown), a note line and a service's reason (500 bytes), a title or a name (300 bytes), and the accepted and expired exception lists (20 each, then "N more") are all printed whole, and so is every query row and every test and tool output. What is bounded is only what the review result itself does not hold: a claim only one model makes, and any limit the review pipeline put on a value before the writer received it. `TestHTMLCarriesTheFullFixNotesAndExceptionLists` reads a 2,050-byte fix, long notes and 25 accepted and 25 expired exceptions off the page; `TestHTMLCompletenessCheckBitesWhenTheMarkdownBoundIsPutBack` puts the Markdown bound back on the HTML path and shows that check turn red. Not measured: how each browser prints the page (this build had no browser).

The goldens are the pages themselves: `testdata/golden/html-report/findings.html` (the sample run on `testdata/fixture/discount`) and `testdata/golden/html-report/clean.html`. `go test ./internal/cli -run TestHTMLGolden -update-golden` rewrites them; the diff is then read like code.
