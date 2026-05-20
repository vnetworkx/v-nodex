# Storage

The rebuilt node uses a local append-only file store by default.

A SQL schema is included for the future durable database-backed variant.

The storage contract is:
- append-only events
- sealed snapshots
- derived state cache
- peer registry
- immutable history replay
