# PV-016 reference

Written by hand with the case, from the pages and the output of `check.sh` on 2026-10-22. It is used only by the
check and the evaluation. A reviewer never receives it. Page paths are under `change/docs/alerts/`.

## Code-facing claims (each quoted from the pages, with the code's observable result)

| id | page | quoted claim | what the code did |
|---|---|---|---|
| D-setup | chat-setup.md | Set `destination` to `chat:t-backup-ops`. | refused: id must begin with c- |
| D-index | index.md | Destination ids begin with `c`, for example `chat:c-backup-ops`. | accepted |
| P-* (five alerts) | what-goes-out.md | Each alert listed here is delivered to the configured destination. | considered=1 delivered=0, reason "no transport yet", for all five alerts |
| F-backup_failed | what-goes-out.md | `backup_failed` carries: job, host, error, time, retry_count. | retry_count is not emitted |
| F-storage_low | what-goes-out.md | `storage_low` carries: volume, percent_free, threshold, time. | threshold is not emitted |
| F-backup_late, F-restore_done, F-delivery_failed | what-goes-out.md | the field lists as written | every claimed field is emitted |

## Page-internal findings (not checked by running code; each is a pair of quoted sentences)

1. what-goes-out.md: "It carries no actor." against related.md: "Routing rules can select an alert by its actor, for
   example by the actor who started a restore."
2. No page says whose environment holds the credential or which machine sends. chat-setup.md only says: "Put the chat
   credential in the `credential` setting of the alert configuration".
3. related.md: "the `delivery_failed` alert goes to the same destination as every other alert." The destination that
   failed is the destination the failure alert is sent to.
4. index.md: "Destination ids begin with `c`" against chat-setup.md: "A destination id may begin with `t`."
