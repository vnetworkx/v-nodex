# v-nodex API Service

This service exposes read-focused HTTP endpoints for the Vector Network node state.

## What it does

The API service is meant to:

- serve health checks
- expose derived state
- expose event history
- expose snapshots
- expose peer and sync metadata where appropriate

## What it is for

Use the API service when you want to:

- query state without using the full node runtime
- build dashboards or explorers
- inspect event history
- validate snapshot and recovery information
- integrate external systems with read access

## What it is not for

The API service is not the source of truth.

It should not:
- bypass validation
- rewrite history
- mutate canonical records
- replace the kernel

## Typical usage

Run the service:

```bash
go run ./services/api/cmd/vnodex-api
```

Or with Docker Compose:

```bash
docker compose up
```

## How it connects to the system

The API service depends on:
- the storage layer
- the protocol model
- the server handlers
- the same deterministic record history as the node

It reads from canonical data and returns query-friendly results.

## Operational notes

If a query result disagrees with replayed history, the replayed history wins.

The API exists to expose canonical data safely, not reinterpret it.
