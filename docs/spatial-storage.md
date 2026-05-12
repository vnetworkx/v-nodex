# Spatial Storage

## Purpose

This document explains the spatial vector storage mechanism used by v-nodex.

The system models space as an immutable history of vector events rather than as a mutable grid of balances or records.

## Main idea

The authoritative truth of the system is not the current state object.

The authoritative truth is the immutable event history from which current spatial state is reconstructed.

That means:

- events are stored permanently
- live space is derived
- snapshots are checkpointed views
- peers replicate the same history
- reconstruction is deterministic

## Core objects

### vSpace
The top-level spatial domain.

It contains:
- event history
- spatial partitions
- snapshots
- derived state views

### vRegion
A spatial partition.

It may be defined by:
- coordinate bounds
- spatial hash
- voxel chunk
- octree node
- logical region boundary

### vEntity
An object that exists in space and can carry or transform vector state.

### vVectorState
The derived vector state associated with an entity or region.

This is not the source of truth. It is a reconstruction.

### vEvent
An immutable spatial event.

This is the atomic storage unit of the spatial system.

## Storage model

The recommended model is:

- immutable event store
- spatial index
- derived state cache
- snapshot store

### Immutable event store
Holds every accepted event permanently.

### Spatial index
Accelerates lookup by:
- region
- entity
- time
- causality

### Derived state cache
Stores reconstructed live state for fast reads.

### Snapshot store
Stores sealed recoverable checkpoints.

## Event format

A spatial event typically contains:

- event ID
- parent IDs
- entity ID
- region ID
- vector type
- operation
- input vector
- output vector
- magnitude and direction metadata
- coordinates
- timestamp or logical order
- signer ID
- signature
- hash
- validation proof

## Vector semantics

The storage layer preserves vector-aware meanings for:

- position vectors
- free vectors
- unit vectors
- bound vectors
- zero vectors

Supported operations include:
- add
- subtract
- scale
- normalize
- rotate
- project
- constrain
- compose
- nullify

## Causality

The system must preserve causality by using:

- parent hash links
- logical order
- region-local ordering
- vector clocks or Lamport-style ordering where needed

This ensures replay can rebuild the same spatial state.

## Partitioning

Space should not be stored as one flat object.

Recommended partitions:
- octree regions
- spatial hash buckets
- layered zones
- adaptive subregions

Each region may store:
- local event history
- neighbor references
- derived snapshot metadata
- boundary-crossing events

## Replay model

The system must support deterministic replay from the same genesis and event history.

That means:
- same events
- same rules
- same order
- same final spatial state

## Why this matters

This storage model is useful when the spatial system must be:
- decentralized
- auditable
- replayable
- resistant to silent mutation
- efficient for local reads

It is a history-first design, not a mutable-state-first design.
