# Related

## Routing

Routing rules can select an alert by its actor, for example by the actor who started a restore.

## When delivery fails

When an alert cannot be delivered, a `delivery_failed` alert is sent. There is one destination for all alerts, so
the `delivery_failed` alert goes to the same destination as every other alert.
