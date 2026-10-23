# PV-016 — Documentation that describes a feature whose code path does nothing

- incident: I13 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-23 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A backup tool can send alerts to a team chat. The change under review is documentation only: four new pages that tell a
customer which alerts the tool sends, what each alert carries, how to set up a chat destination and where to put the
chat credential. No code changes.

Two reviewers from different model families reviewed the same change, without seeing each other. They had been given
different questions on purpose. The first was told to read only the shipped pages, as a customer would, and never the
code. It found that the pages contradict each other and leave out facts an operator needs, and it accepted with blocking
conditions. The second was told to compare the pages with the code. It found that the code at that version resolves the
alert sender to one that sends nothing, so none of the promised alerts can arrive, and it rejected.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- one version of the tool's code, used for both **base** and **change**, offering three operations:
  1. given a destination setting, accept or refuse it, with the reason;
  2. given an alert event, emit it and report how many alerts were considered and how many were delivered, to an
     in-memory receiver;
  3. given an alert event, list the fields the emitted alert carries.
- in that code, the sender is always resolved to one that sends nothing, with the reason "no transport yet". So every
  emitted alert is considered and none is delivered.
- **base**: the code with no alert pages. **change**: the same code plus four documentation pages: a section index,
  "what goes out", "chat setup" and "related".
- the pages must contain, by construction:
  - a promise that each listed alert is delivered to the configured destination;
  - setup steps whose destination setting the code refuses;
  - field lists that claim fields the code does not emit for at least one alert;
  - four contradictions or omissions visible from the pages alone: an alert said to carry no actor yet routed by actor;
    no page says whose environment holds the credential or which machine sends; the "delivery failed" alert is routed
    through the same single destination that failed; the index says destination ids begin with one letter while the
    setup page allows another.
- a **reference**: a list of the pages' claims, each one quoted from the pages with the code's observable result next to
  it, written by hand with the case and never derived from the reviewer; used only by the check and the evaluation.
- a **check** that prints one line per code-facing claim: claim, what the pages say, what the code did (accepted or
  refused, considered and delivered counts, fields present). The four page-internal findings cannot be checked by
  running code; the reference lists them as pairs of quoted sentences.

Credentials are synthetic only, such as `tok_synthetic_0000_not_a_secret`. The toolchain is chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P29 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original pages and code are private and are not published. Resolution status: settled by the next round, which
  adopted both reviewers' blocking items; the missing delivery was closed by another change that added a real sender.

## 4. Synthetic transformation

Replaced: the product, the kind of alerts, the chat vendor, the page names and every field name. Kept, as the causal
property of the disagreement: documentation that describes a feature in detail while the code path behind it does
nothing; pages that also contradict each other; and two reviewers assigned disjoint subjects — one forbidden the code,
one comparing pages with code — so the decisive defect was visible to only one of them.

## 5. Review target

Two variants, matching the two assigned subjects:
- **pages only**: the four pages and a readable diff against base. The reviewer is told to read as a customer and not
  to look at code.
- **pages and code**: the pages, the diff and the code. The reviewer is told to check the pages against the code.

Either reviewer receives the description of the change in §1 and its own instruction — never the check, the reference
or `expected/`. An evaluation may also give one reviewer both subjects.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions; pages only, by assignment):** four blocking contradictions or omissions, ten
  owed items and six notes, all visible from the pages. It kept the verdict at accept-with-conditions because it had not
  built the site. Evidence in the record: references to specific statements in the pages and sibling pages, the site's
  content checks (green), and link counts. Its findings were accurate. By assignment it did not read the code, so it
  could not see that nothing could be delivered. The orchestrating agent counted its four blocking items as a rejection.
- **Reviewer B (reject; pages and code, by assignment):** the code always resolves the alert sender to one that sends
  nothing, so the pages promise alerts that cannot arrive; the setup steps and several field claims also contradict the
  code. Evidence in the record: a reading of the code, including the "no transport yet" reason; targeted tests of the
  configuration and the alert catalogue, which passed. Some of its builds and tests could not run in its environment.
  Its no-delivery finding was true of the code at review time.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on change and record each line of its output; compare a review's findings with the reference list.
Run 2026-10-23T01:51:37Z (UTC). Toolchain: `go version go1.26.5 linux/amd64`, standard library only, no network.
Base and change hold the same three code files (`go.mod`, `alerts.go`, `main.go`; the check refuses to run if they
differ); change adds `docs/alerts/` with the four pages. The hand-written claim list is `claims.tsv`; the hand-written
reference, with the four page-internal pairs, is `reference.md`.

Command, from the repository root:

```
bash tests/corpus/cases/PV-016/check.sh
```

Exit code: 0. Output (24 lines, one per claim and version):

```
base | D-setup | pages say: (no such page) | code did: setting chat:t-backup-ops: refused: destination id "t-backup-ops" must begin with c-
base | D-index | pages say: (no such page) | code did: setting chat:c-backup-ops: accepted: c-backup-ops
base | P-backup_failed | pages say: (no such page) | code did: emit backup_failed: considered=1 delivered=0 reason="no transport yet"
base | P-backup_late | pages say: (no such page) | code did: emit backup_late: considered=1 delivered=0 reason="no transport yet"
base | P-restore_done | pages say: (no such page) | code did: emit restore_done: considered=1 delivered=0 reason="no transport yet"
base | P-storage_low | pages say: (no such page) | code did: emit storage_low: considered=1 delivered=0 reason="no transport yet"
base | P-delivery_failed | pages say: (no such page) | code did: emit delivery_failed: considered=1 delivered=0 reason="no transport yet"
base | F-backup_failed | pages say: (no such page) | code did: alert backup_failed: present=error,host,job,time
base | F-backup_late | pages say: (no such page) | code did: alert backup_late: present=expected_at,host,job,time
base | F-restore_done | pages say: (no such page) | code did: alert restore_done: present=host,job,time
base | F-storage_low | pages say: (no such page) | code did: alert storage_low: present=percent_free,time,volume
base | F-delivery_failed | pages say: (no such page) | code did: alert delivery_failed: present=destination,reason,time
change | D-setup | pages say: "Set `destination` to `chat:t-backup-ops`." (chat-setup.md) | code did: setting chat:t-backup-ops: refused: destination id "t-backup-ops" must begin with c-
change | D-index | pages say: "Destination ids begin with `c`, for example `chat:c-backup-ops`." (index.md) | code did: setting chat:c-backup-ops: accepted: c-backup-ops
change | P-backup_failed | pages say: "Each alert listed here is delivered to the configured destination." (what-goes-out.md) | code did: emit backup_failed: considered=1 delivered=0 reason="no transport yet"
change | P-backup_late | pages say: "Each alert listed here is delivered to the configured destination." (what-goes-out.md) | code did: emit backup_late: considered=1 delivered=0 reason="no transport yet"
change | P-restore_done | pages say: "Each alert listed here is delivered to the configured destination." (what-goes-out.md) | code did: emit restore_done: considered=1 delivered=0 reason="no transport yet"
change | P-storage_low | pages say: "Each alert listed here is delivered to the configured destination." (what-goes-out.md) | code did: emit storage_low: considered=1 delivered=0 reason="no transport yet"
change | P-delivery_failed | pages say: "Each alert listed here is delivered to the configured destination." (what-goes-out.md) | code did: emit delivery_failed: considered=1 delivered=0 reason="no transport yet"
change | F-backup_failed | pages say: "`backup_failed` carries: job, host, error, time, retry_count." (what-goes-out.md) | code did: alert backup_failed: present=error,host,job,time; claimed-but-absent=retry_count
change | F-backup_late | pages say: "`backup_late` carries: job, host, expected_at, time." (what-goes-out.md) | code did: alert backup_late: present=expected_at,host,job,time; claimed-but-absent=none
change | F-restore_done | pages say: "`restore_done` carries: job, host, time. It carries no actor." (what-goes-out.md) | code did: alert restore_done: present=host,job,time; claimed-but-absent=none
change | F-storage_low | pages say: "`storage_low` carries: volume, percent_free, threshold, time." (what-goes-out.md) | code did: alert storage_low: present=percent_free,time,volume; claimed-but-absent=threshold
change | F-delivery_failed | pages say: "`delivery_failed` carries: destination, reason, time." (what-goes-out.md) | code did: alert delivery_failed: present=destination,reason,time; claimed-but-absent=none
```

Negative control: changed one quote in the existing change-side `chat-setup.md` page, then ran the same command.
It printed `check: change D-setup: quote not found in chat-setup.md` and exited 1. The original page was restored;
the output above is the restored run. The same mutation exited 0 before the check was fixed.

Not run: macOS with Go 1.26. This run used Linux; cross-platform execution remains unverified.

## 9. Limits

- Four of the findings are prose-only: they are checked against a hand-written list, not by running code.
- Reviewer B's individual findings beyond no delivery were adopted by the next round but were not re-verified against the
  original code for this corpus.
- The case does not judge the assignment of disjoint subjects; it shows what each subject can and cannot see.
