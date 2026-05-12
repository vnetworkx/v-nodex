# API

## Purpose

This document explains the public and internal API surfaces exposed by v-nodex.

The API layer is designed for reading state, inspecting history, submitting events where appropriate, and coordinating sync or debug workflows.

## API design principles

The API should be:

- deterministic
- explicit
- read-friendly
- validation-aware
- safe to inspect
- consistent with canonical history

It should never pretend to be the source of truth.

## Main API categories

### 1. Health endpoints
Used for:
- liveness checks
- readiness checks
- simple diagnostics

Example use:
- verify a node is up
- confirm services have started

### 2. State endpoints
Used for:
- current derived state
- entity views
- region views
- heads and checkpoints

These are read surfaces over reconstructed state.

### 3. Event endpoints
Used for:
- event submission
- event lookup
- event inspection
- history browsing

### 4. Snapshot endpoints
Used for:
- reading checkpoint metadata
- recovery inspection
- replay acceleration checks

### 5. Peer endpoints
Used for:
- peer enumeration
- sync observation
- last-seen metadata
- topology inspection

### 6. Sync endpoints
Used for:
- pulling accepted history
- pushing accepted history
- acknowledging replication progress

### 7. Debug endpoints
Used for:
- internal state inspection
- developer diagnostics
- replay debugging
- visibility during early development

## Read vs write surfaces

The API should distinguish between:

- **read-only query routes**
- **controlled mutation routes**
- **sync routes**
- **debug routes**

This separation prevents accidental misuse of the system.

## Data returned by the API

The API may return:

- vector state
- event records
- snapshot metadata
- peer metadata
- region summaries
- derived entity state
- health information

## Important API rule

If an endpoint exposes mutable or authoritatively accepted state, that endpoint must still respect the validation and append-only rules of the kernel.

The API does not get to bypass the storage contract.

## Example usage patterns

### Operator health check
- confirm node is alive
- confirm storage is reachable
- confirm derived state can be served

### Developer inspection
- query recent events
- check a specific entity
- inspect peer list
- verify a snapshot cursor

### Sync partner
- request missing history
- exchange accepted events
- compare heads

## Stability guidance

If you change request or response shapes, update:

- `docs/api.md`
- service README files
- server handlers
- any client tooling

The API is part of the developer contract, so changes should be explicit and version-aware.

## Best practice

Prefer responses that are:
- easy to validate
- easy to replay
- easy to compare across nodes
- clearly tied to canonical history

That keeps the API aligned with the protocol instead of drifting into an ad hoc application interface.
