# REVIEW — corpus-pv-016 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:01:42Z

## Summary

PV-016 reproduces the delivery defect at 5af73bc. A missing page quote is silently misreported; macOS execution is unrun. See conditions.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

### Ground

`git rev-parse --short=7 HEAD` printed `5af73bc`. Scope: `fe146e0...5af73bc`. The diff touches PV-016 outside `expected/` and its own manifest status line only. `git diff --check fe146e0...5af73bc` exited 0. Both variants hold identical Go source and modules; `check.sh:12-17` enforces this. Both modules declare Go 1.26; `go version` printed `go version go1.26.5 linux/amd64`. No dependency or numeric project quality bar was added. The check is a suite instrument, so the source reading and printed observations carry the conclusion, not exit 0 alone.

### Differential

The base has no pages. The change adds four pages and keeps the code unchanged. `change/docs/alerts/what-goes-out.md:3` promises delivery. `base/alerts.go:58-65` always resolves a sender that sends nothing; `base/alerts.go:84-89` counts one considered and zero delivered. Thus the code result is equal across variants, while only the change makes the false promise. `base/alerts.go:36-38` refuses the documented setup setting. The catalogue at `base/alerts.go:19-24` omits the documented `retry_count` and `threshold` fields. The four page-only contradictions or omissions appear in the pages and their quoted pairs in `reference.md`.

Command: `bash tests/corpus/cases/PV-016/check.sh`; exit 0. Full output:

```text
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

### Unplug

I copied PV-016 under the worktree, changed the one delivery-promise sentence to `No alert listed here is delivered to the configured destination.`, and ran the copied check. It exited 0. All five change-side `P-*` lines then said `pages say: (no such page)`, although `what-goes-out.md` existed. The complete output changed, so a reader of that column can distinguish the opposite world; an exit-code-only reader cannot. I removed the copy and generated binaries. Reviewed files were not edited.

### Closure

Verdict: **accept-with-conditions**. The present case establishes the expected `established defect` outcome: five change-side lines quote the promise and show zero delivered. The opposite outcome, where the pages do not promise delivery, changes the page-claim column. The inspected sender independently explains zero delivery. The documented setup is refused, and two claimed fields are absent. No check was bent to force the numbers: both variants run the same code and inputs. The check expressly prints evidence and does not assert an outcome. Its exit status alone is not proof. The hand-written reference agrees with the inspected pages and run.

### Remainder

**OWED — `check.sh:27-33`.** A missing quote in an existing change-side page is reported as `(no such page)` with exit 0. The single-sentence mutation above is the control. Remedy: distinguish missing page from missing quote, and fail when a listed change-side claim disappears. This prevents an exit-0 run from silently outliving its page claim.

**NOTE — `CASE.md:55`, `check.sh:26-66`.** The case says `reference.md` is used by the check, but the check reads only `claims.tsv`. Remedy: correct that description or wire the reference into the check if intended.

No private host or seat name, internal absolute path, private repository name, real token, or email address appeared in the diff. The credential example is explicitly synthetic. `host` is an alert field, not a host identifier. Licences and behavior outside this diff were declined as out of scope.

### Profile

Docs and Go fixture. `go -C`, `diff -q`, `grep -qF`, Bash substitution, `mkdir`, and `printf` were inspected for portability. Go 1.26 is declared. Bash `${claimed//,/ }` is supported by macOS Bash 3.2. No Linux-specific call was found. That is a static assessment, not a macOS run. No skip, suppression, threshold change, or test deletion was found.

### Focus

1. Same code in base and change; change adds pages. Five promised alerts each print `considered=1 delivered=0`.
2. The full output and code support the expected outcome. The opposite world changes the page column, but not exit status.
3. The current check prints real disagreements. The mutation shows its claim-presence guard is incomplete.
4. No private material found.
5. The diff is confined to PV-016 outside `expected/` and its manifest entry.
6. Linux run passed. macOS calls were inspected, but execution is unrun.

### Conditions

1. Fix `check.sh:27-33` so a missing listed quote in an existing change-side page fails by name. The positive check must still print 24 claims; the one-quote mutation must exit nonzero.
2. Run `bash tests/corpus/cases/PV-016/check.sh` on macOS with Go 1.26 and record output and exit status before claiming cross-platform execution.

### Unrun

macOS execution: no macOS host was available to this review. The full Go test suite, race detector, vet, and coverage were not run: the order's acceptance instrument is this case's check, and the change adds no Go tests, shared-state code, or numeric bar. Model and dollar cost were unavailable as a completed Saphan Studio receipt; they must come from that record rather than an estimate.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-016 reproduces the delivery defect at 5af73bc. A missing page quote is silently misreported; macOS execution is unrun. See conditions.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T17:55:46Z

## REDACTED

none
