# v-nodex Node Service

This service runs the main decentralized node process for v-nodex.

## What it does

The node service:

- loads configuration
- opens storage
- restores derived state
- tracks peers
- handles sync
- accepts validated events
- exposes HTTP endpoints
- creates snapshots

## What it is for

Use the node service when you need a full working node that can:

- store immutable history
- reconstruct live state
- sync with peers
- validate and append events
- serve both internal and operator endpoints

## Typical startup

Run directly:

```bash
go run ./services/node/cmd/vnodex-node
```

Or via Docker Compose:

```bash
docker compose up
```

## CLI

The node service repository also includes a CLI tool for local inspection and debugging.

Example:

```bash
go run ./services/node/cmd/vnodex-cli --help
```

## Runtime behavior

The node combines:
- config loading
- storage initialization
- peer registry setup
- sync manager setup
- HTTP server startup

## What to expect

When healthy, the node should:
- respond to health checks
- accept or reject events deterministically
- keep derived state consistent with history
- replicate accepted history to peers

## Operational caution

A node should never be trusted solely because it is running.

Its history and hashes must still be verifiable against canonical replay.
