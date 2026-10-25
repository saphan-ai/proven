# Review output

JSON and SARIF are two writings of one review result (ADR-0020 §1). The chosen format goes to standard output, or to `-o`. Each `--text-output`, `--json-output` and `--sarif-output` file is also written in that same run. Empty lists are `[]`, not null.

Two outputs of one run must be two files. If `-o` and any `--*-output`, or two `--*-output` flags, name the same file, the run is refused at load with exit 7. The refusal names both flags and the path, and nothing is written. Each path is resolved the way the filesystem resolves it: component by component, each symbolic link followed as it is met and only then any `..` applied, never cleaned first (cleaning `link/..` first could name the wrong file); a path that passes through a directory that does not exist cannot be written and is left to fail at the write with exit 2. So `-o real/r.json` with `--sarif-output link/r.json` (where `link` points to `real`) is refused. Two existing files that are the same file (a hard link) are refused too. Two spellings of a file that does not exist yet and differ only in letter case, on a case-insensitive volume, are not refused at load: the second write finds the file the first one made, the run stops with exit 2 and names both flags, and the first output is left as it was written.

An output path whose final component is a symbolic link is refused at load with exit 7. The refusal names the flag and the path, and nothing is written. The link is not followed: the file it names is left as it was. A regular file at the path is replaced: the report is a new file renamed over it, so it gets a new inode and mode `0644` less the umask, and a hard link to the old file keeps the old content.

Any earlier component of that path that lies inside the working directory is refused at load with exit 7 when it is a symbolic link. The refusal names the flag, the path and the linked component, and nothing is written. The working directory and the prefix already walked are resolved before that comparison, so an absolute path that reaches the checkout through a different spelling is the same check. In `proven ci` the working directory is the checkout, so a directory of a relative output path that stays in the checkout, or of an absolute path that resolves to it, cannot be a link that chooses another directory. A relative path whose `..` climbs out of the working directory, even after passing through a directory inside it (`reports/../../x`), is outside it: it is the operator's, and its directories are followed at the write. A component that lies outside the working directory is the operator's choice and is not refused here. A symbolic link that is a directory on the way is still resolved when two outputs are checked for being one file. That check runs first and still exits 7, naming both flags.

A path inside the working directory is written through its directory, opened at load after these checks and before the review runs. A later rename of that directory, or a link put at its name (to a directory inside the checkout or outside it), does not change where the report goes: it is written in the directory the check saw, and if that directory was removed the write fails with exit 2 and names the path. A directory that did not exist when the run started is not looked up again at the write: that output fails with exit 2 and names the path, because the review could have made it a link. The report is a new file in that directory, and that file is renamed over the final name: the final name is replaced, never written through. A symbolic link swapped in as that name after the check is replaced, and the file it names is not written. A directory component swapped for a link that leaves the directory still cannot receive the report: the write fails with exit 2 and names the path, and nothing outside is written. A path outside the working directory is not opened through that root: a directory link on it is still followed, including one swapped in after the load check, and where the system has no open that refuses a final symbolic link, a final link on that path swapped in between the check and the open can still be followed. A final link that is seen at the write is refused on every path, and that refusal is exit 2. Where the system can open a file without following a final symbolic link (`O_NOFOLLOW`), that open is what refuses a final link on a path outside the working directory.

DEL (U+007F) and C1 (U+0080 to U+009F) in JSON and SARIF are written as `\u007f` and `\u0080` to `\u009f`. The file stays valid JSON and decodes to the same text. The JSON encoder leaves those runes raw, so both writers escape them after encoding.

## JSON

One object. Every key below is always present.

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
| `findings` | every reported finding |
| `findings[].id` | finding id |
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

## SARIF 2.1.0

One log, `version` `2.1.0`, one run. The tool driver `name` is `proven`. One result per finding. `level` is `INFO`→`note`, `WARNING`→`warning`, `ERROR`→`error`; any other severity is `note`. `ruleId` is `rule_id`, or `id` when the finding has no rule. The message is the finding's message, then its title, then its id. A finding with no path has no location; a start line below 1 gives no region; an end line before the start line is left out.

The artifact `uri` is a relative path inside the repository: no leading `/`, no authority, and no `..` segment. The path is cleaned first, so `dir/../a.go` is `a.go`. A path that is absolute, names an authority such as `//example.invalid/a.go`, or still climbs out after cleaning, is omitted. The finding is kept with no location. This log does not set `uriBaseId`.

Each result's `properties` carries `id`, `rule_id`, `title`, `severity`, `unresolved` and `evidence`. `evidence` is the same array as in the JSON.

The run's `properties` carries `class`, `class_rule`, `partial`, `failed_services`, `settings_statement`, `author_model_statement`, `reviewer_statements`, `trace_path` and `family_sentence`. Each has the same value as the JSON key of that name, so the lines the text output prints beside the findings are in a SARIF file too.
