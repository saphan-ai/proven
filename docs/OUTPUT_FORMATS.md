# Review output

JSON and SARIF are two writings of one review result (ADR-0020 §1). The chosen format goes to standard output, or to `-o`. Each `--text-output`, `--json-output` and `--sarif-output` file is also written in that same run. Empty lists are `[]`, not null.

## JSON

One object. `findings` is every reported finding. `failed_services` is every service the review named as failed.

| field | meaning |
|---|---|
| `findings[].id` | finding id |
| `findings[].rule_id` | kind of finding |
| `findings[].severity` | `INFO`, `WARNING` or `ERROR` |
| `findings[].unresolved` | the evidence did not settle it |
| `findings[].location.path`, `start_line`, `end_line` | where it is |
| `findings[].evidence[].kind` | `test` or `query` |
| `findings[].evidence[].test` | `path`, `name`, `command`, `fails_on_change`, `passes_on_base`, `output` |
| `findings[].evidence[].query` | `sql`, `engine`, `snapshot`, `columns`, `rows` |
| `failed_services[].service`, `reason` | a service that failed, and why |

`class`, `class_rule`, `partial`, `doubled`, the reviewer and author fields, `models` and `trace_path` are the rest of the result. `trace_path` is the path the review wrote, and it may be absolute.

## SARIF 2.1.0

One log, `version` `2.1.0`, one run. The tool driver `name` is `proven`. One result per finding. `level` is `INFO`→`note`, `WARNING`→`warning`, `ERROR`→`error`. `ruleId` is `rule_id`, or `id` when the finding has no rule. The message is the finding's message, then its title, then its id.

Each result's `properties` carries `id`, `rule_id`, `title`, `severity`, `unresolved` and `evidence`. `evidence` is the same array as in the JSON. The run's `properties` carries `class`, `class_rule`, `partial` and `failed_services`, the same service objects as the JSON.
