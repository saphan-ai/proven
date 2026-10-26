# REVIEW — corpus-pv-016 · round 6

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T22:13:21Z

## Summary

PV-016 reproduces all 24 recorded lines; the round-2 path repair is complete. No blocking conditions remain.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

### Ground

At measured HEAD `ffd3e35c8a96f4b31a0a12b08290b750ea8ab8ba`, `bash test/corpus/cases/PV-016/check.sh` exited 0 with Go 1.26.5 linux/amd64. All 24 lines match the recorded CASE.md §8 output byte for byte: `diff -u` of an `awk` extraction of the fenced block against a fresh run returned 0. An initial extraction pattern selected no lines and returned 1; the corrected pattern returned 0. The check builds both variants, checks equality of their three Go files, and checks change-side quote presence. Its observations concern this synthetic fixture, not a live chat service.

Fresh check output (24 lines):

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

### Differential

`git diff 3a84be4..ffd3e35` touches only `test/corpus/cases/PV-016/CASE.md:102` and `test/corpus/cases/PV-016/check.sh:2,6`: the three `tests/corpus` references become `test/corpus`. `git diff --name-only 9303b2d...ffd3e35` names 14 files within PV-016 plus `test/corpus/manifest.json`, nothing elsewhere. `git diff --check` exited 0. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-016` found no matches. The round-2 retired-path condition is ADDRESSED: the documented command runs and matches §8. No bar was weakened, no exception added, and no dependency or performance change was made by this path repair.

### Unplug

False-world answer: a functioning sender would show `delivered=1` rather than `0` in the five `P-*` rows; a valid setup setting would say accepted in `D-setup`. The observed outputs differ. `base/alerts.go` resolves to `nullSender` with reason "no transport yet", so the fixture cannot measure external delivery. CASE.md §8 records a historical quote-mutation control that exits 1; I did not independently run it. An attempted disposable-copy mutation command was refused by the command tool because it included a denied cleanup act. I did not re-spell that act or modify the reviewed files.

### Closure

Read `expected/EXPECTED.md` last. Its expected outcome is an established change-side defect: the pages promise delivery but all five listed alerts report `considered=1 delivered=0`; documented setup is refused; `retry_count` and `threshold` are claimed but absent. The reference lists four page-internal inconsistencies/omissions; the executable check by itself does not prove them. The opposite outcome of successful fixture delivery cannot give these zero-delivery rows under the fixture's sender contract. This is not independent production-system evidence; the other code operations work as specified.

### Remainder

No BLOCKING or OWED condition for this repair. Declined to judge live chat delivery, private provenance, licences, and platforms other than this linux/amd64 run. Noticed outside re-review: synthetic null-sender output cannot establish live transport behavior; CASE.md states this design explicitly.

### Profile

Docs: CASE.md §8's command exists and runs. The repository's own check is a suite instrument and is more circular than an external integration check; the `diff -u` comparison independently checks recorded text but not real-world delivery. The path search is a project-source check. No coverage, mutation, lint or type metric, command or baseline was declared for the path repair, so none is reported as met. No weakening moves or exception touched.

### Focus

FILE·LINE·CONTROL: `test/corpus/cases/PV-016/CASE.md:102` documents an existing runnable command; `test/corpus/cases/PV-016/check.sh:2,6` targets that same directory. The previous path condition is ADDRESSED. No new finding admitted.

### Conditions

None. Verdict: accept.

### Unrun

Independent quote-mutation negative control: not run; the proposed disposable-copy command was refused on its cleanup act. macOS/arm64, live chat transport and the private original incident: not run or accessible on this seat. No baseline/command was declared for coverage, mutation, lint or type metrics; not run. The check itself and the exact §8 comparison were run.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-016 reproduces all 24 recorded lines; the round-2 path repair is complete. No blocking conditions remain.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T22:10:00Z

## REDACTED

none
