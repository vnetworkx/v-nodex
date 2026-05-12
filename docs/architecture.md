# Architecture

## Purpose

This document explains how v-nodex is organized and how the main pieces interact.

v-nodex is a deterministic node runtime for the Vector Network. It is built to store immutable history, derive live state from that history, synchronize records across peers, and expose safe read/write surfaces for applications and operators.

## Design goals

The architecture is built around five non-negotiable goals:

- deterministic execution
- append-only history
- replayable state reconstruction
- peer-to-peer synchronization
- separation between truth and views

## High-level layers

The system is divided into these major layers:

1. **Protocol model**  
   Vector types, event records, hashes, and validation rules.

2. **Storage engine**  
   Append-only event persistence, derived state, snapshotting, and peer registry storage.

3. **Sync engine**  
   Peer registration, event exchange, snapshot propagation, and deterministic reconciliation.

4. **HTTP server layer**  
   Public query endpoints, event submission endpoints, sync endpoints, and debug endpoints.

5. **Node orchestration**  
   Startup, dependency wiring, config loading, and service lifecycle management.

6. **Tooling and operator interfaces**  
   CLI, API service, inspection tooling, and containerized startup.

7. **Rust kernel**  
   Strict deterministic vector operations and protocol-level validation logic.

## Core architecture rule

The architecture follows this rule:

**history is stored, state is derived, and live views are rebuildable**

That means:

- writes go into immutable event history
- reads come from reconstructed live state or replay-safe indexed views
- snapshots are only checkpoints
- caches are only accelerators

## Major runtime components

### `internal/model`

Defines the protocol objects used by the Go runtime.

It includes:

- vector operations
- event schema
- canonical hashing
- validation rules

### `internal/store`

Owns the durable local state of the node.

It is responsible for:

- loading existing data
- appending accepted events
- reconstructing live state
- writing snapshots
- persisting peer metadata
- serving query-friendly state views

### `internal/peer`

Keeps track of known peers and their sync metadata.

### `internal/sync`

Implements the node-to-node exchange model.

### `internal/server`

Exposes the HTTP routes for health, query, event handling, sync, and debug operations.

### `internal/node`

Composes all runtime pieces into a running node process.

### `services/node`

Node entrypoint and CLI tools.

### `services/api`

Read-oriented API service.

### `rust/kernel`

Provides a more strictly typed kernel implementation of the vector protocol semantics.

## Data flow

A normal write path looks like this:

1. a client submits an event
2. the server receives the request
3. the model layer validates and hashes the event
4. the store appends the event
5. the store updates derived live state
6. the node may create a snapshot later
7. the sync layer can propagate the event to peers

A normal read path looks like this:

1. a client asks for state or history
2. the server routes the query
3. the store returns the current derived view or event history
4. the response is built from immutable data and deterministic reconstruction

## Why this architecture is appropriate

This design is appropriate for decentralized state because it avoids three common failure modes:

- mutable truth
- nondeterministic replay
- hidden state changes

Instead, it uses:

- explicit records
- hash-linked history
- deterministic reconstruction
- peer-verifiable sync

## Boundaries

### What is authoritative
- validated event history
- replay rules
- canonical hashes
- snapshot roots derived from history

### What is not authoritative
- local caches
- debug views
- client-side copies
- temporary sync state
- unreadable partial output

## Evolution path

The repo is structured so that the protocol can evolve in phases:

1. deterministic local node
2. durable history and replay
3. sync and replication
4. query APIs
5. developer tooling
6. richer spatial/vector semantics
7. eventual testnet and protocol hardening

## Design principle for future changes

Any change must preserve the ability to answer this question:

**Given the same genesis, the same accepted events, and the same rules, do all honest nodes reconstruct the same state?**

If the answer is no, the change is not safe for the kernel.
