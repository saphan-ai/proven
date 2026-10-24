# Review output

JSON and SARIF are two writings of one review result (ADR-0020 §1). The chosen format goes to standard output, or to `-o`. Each `--text-output`, `--json-output` and `--sarif-output` file is also written in that same run. Empty lists are `[]`, not null.

Two outputs of one run must be two files. If `-o` and any `--*-output`, or two `--*-output` flags, name the same file, the run is refused at load with exit 7. The refusal names both flags and the path, and nothing is written. Each path is resolved the way the filesystem resolves it: component by component, each symbolic link followed as it is met and only then any `..` applied, never cleaned first (cleaning `link/..` first could name the wrong file); a path that passes through a directory that does not exist cannot be written and is left to fail at the write with exit 2. So `-o real/r.json` with `--sarif-output link/r.json` (where `link` points to `real`) is refused. Two existing files that are the same file (a hard link) are refused too. Two spellings of a file that does not exist yet and differ only in letter case, on a case-insensitive volume, are not caught.

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

Each result's `properties` carries `id`, `rule_id`, `title`, `severity`, `unresolved` and `evidence`. `evidence` is the same array as in the JSON.

The run's `properties` carries `class`, `class_rule`, `partial`, `failed_services`, `settings_statement`, `author_model_statement`, `reviewer_statements`, `trace_path` and `family_sentence`. Each has the same value as the JSON key of that name, so the lines the text output prints beside the findings are in a SARIF file too.
