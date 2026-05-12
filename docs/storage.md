# Storage

## Purpose

This document explains the storage model used by v-nodex.

The storage layer is built to preserve immutable history, support fast reads, and make deterministic replay practical at node scale.

## Storage model summary

v-nodex uses a hybrid model:

- **append-only event history** as the authoritative record
- **derived live state** for fast reads
- **snapshot store** for fast recovery
- **indexes** for query acceleration
- **peer metadata** for sync coordination

The main rule is:

**history is authoritative; current state is derived**

## Storage layers

### 1. Immutable event store

This stores accepted events permanently.

Properties:

- append-only
- hash-addressed or hash-linked
- replayable
- signed or signature-verifiable
- never overwritten as canonical truth

This layer is the durable source of record.

### 2. Derived live-state store

This stores the current reconstructed state.

Properties:

- mutable
- rebuildable
- disposable
- optimized for reads
- always derivable from history

This layer is a cache and view layer, not the truth layer.

### 3. Snapshot store

This stores sealed checkpoints.

Properties:

- derived from accepted history
- immutable once sealed
- used to accelerate recovery
- replay anchors for validation

### 4. Index layer

This stores query acceleration structures, such as:

- event-by-hash lookup
- entity-to-event lookup
- region-to-event lookup
- head and checkpoint indexes
- peer lookup indexes

## Data model

The core persistent entities are:

- `v_events`
- `v_state_heads`
- `v_snapshots`
- `v_peers`

These correspond to:

- immutable event records
- current head markers
- snapshot checkpoints
- known peer registry

## Write path

The write path is designed to be deterministic:

1. validate event structure
2. check event integrity
3. compute canonical hash
4. append to history
5. apply event to derived state
6. update heads and indexes
7. optionally schedule snapshot creation
8. replicate to peers

If validation fails, the event is rejected and must not mutate canonical state.

## Read path

The read path can be served from multiple sources depending on the query:

- live state cache for hot lookups
- derived state tables for structured reads
- event store for history queries
- snapshot store for checkpoint inspection

The key constraint is that every read must remain consistent with replayable truth.

## Snapshotting

Snapshots reduce replay cost.

A snapshot should contain:

- the reconstructed state at a known point
- the event root or history cursor it corresponds to
- the hash of the sealed state
- enough metadata to verify recovery

Snapshots are not a replacement for history. They are acceleration structures.

## Recommended backend roles

### PostgreSQL
Use for:
- durable event history
- metadata
- snapshot references
- peer registry
- indexed query access

### Redis
Use for:
- hot live-state caching
- short-lived query acceleration
- ephemeral sync or session data

### Rust embedded storage
Use for:
- local deterministic kernel storage
- low-level record handling
- protocol data structures

## Storage integrity

The storage system should always be able to answer:

- what was accepted?
- in what order?
- by which peer?
- under which hash?
- from which parent?
- what state does replay produce?

If it cannot answer those questions, the storage layout is incomplete.

## Recovery model

Recovery should follow this hierarchy:

1. load latest valid snapshot
2. replay events after the snapshot cursor
3. reconstruct derived live state
4. verify hashes and heads
5. resume sync

## Immutability rule

Canonical event history must never be edited in place.

If something is invalid, the correct response is:

- reject before append
- quarantine separately
- record the rejection or error path as a distinct audit artifact if needed

## Operational guidance

The storage system should be designed for:

- clear append semantics
- deterministic replay
- low-cost verification
- partial recovery after failure
- peer-verifiable history

That is what turns the storage layer into a trustworthy protocol component instead of a normal application database.
