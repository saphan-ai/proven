# What goes out

Each alert listed here is delivered to the configured destination.

## The alerts

- `backup_failed` carries: job, host, error, time, retry_count.
- `backup_late` carries: job, host, expected_at, time.
- `restore_done` carries: job, host, time. It carries no actor.
- `storage_low` carries: volume, percent_free, threshold, time.
- `delivery_failed` carries: destination, reason, time.

Alerts reach the chat within a minute of the event.
