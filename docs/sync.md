# Sync

## Purpose

This document explains how nodes synchronize data in v-nodex.

The sync layer is responsible for moving accepted history between peers, reconciling heads, and keeping derived state aligned across nodes.

## Sync goals

The sync system must:

- exchange events deterministically
- propagate snapshots when useful
- track peer state
- detect missing history
- recover from partial partitions
- avoid accepting invalid records

## Sync model

The basic sync unit is the accepted event history, not arbitrary mutable state.

Nodes exchange:

- event batches
- snapshot checkpoints
- state heads
- peer metadata
- acknowledgement information

## Sync lifecycle

A typical sync cycle looks like this:

1. a node discovers or remembers a peer
2. the node asks what the peer has
3. the peer responds with available heads or events
4. the receiving node requests missing history
5. the sender transmits events and optional snapshots
6. the receiver validates everything
7. accepted events are appended locally
8. the receiver updates derived state
9. both sides move their heads forward

## Event-first replication

The sync system prefers event replication first.

That means:
- events are the primary unit of exchange
- state is reconstructed locally
- snapshots are only accelerators
- peers must be able to replay the same history

## Peer behavior

A peer should be treated as:

- a candidate source of history
- a validation participant
- a sync partner
- a local consistency checker

A peer is not automatically trusted.

The node must still verify:
- hashes
- signatures
- parent relationships
- logical ordering
- structural validity

## Sync endpoints

The runtime includes HTTP handlers for peer communication. Typical behaviors include:

- pull current history from another node
- push local events to another node
- exchange acknowledgements
- register new peers
- query sync state

The exact route names are defined in `internal/server/server.go` and `internal/sync/sync.go`.

## Missing history recovery

If a node is missing part of the accepted sequence:

1. it identifies the missing cursor or hash range
2. it requests the gap from peers
3. it applies events in order
4. it rebuilds state after the gap is closed

This is why history is stored append-only: the node can always reconstruct.

## Conflict handling

If a peer sends:
- invalid hashes
- broken parent references
- malformed events
- contradictory data

the receiving node must reject those records.

In a decentralized system, rejection is a normal and necessary outcome.

## Snapshot sync

Snapshots may be exchanged when:

- a peer is far behind
- replay cost is high
- recovery is needed after restart
- a region wants to bootstrap quickly

Snapshots should always be verifiable against history.

## Determinism requirement

Sync is only valid if all honest nodes can arrive at the same result from the same accepted input.

That means:
- ordered application matters
- event validation matters
- hash calculation matters
- replay rules matter

If sync introduces nondeterminism, the protocol is broken.

## Practical operators’ view

For operators, sync means:

- peers are connected
- accepted history is flowing
- local state is not drifting
- snapshots are helping recovery
- invalid data is rejected

## Practical developer view

For developers, sync means:

- one node can be treated as a source of accepted history
- another node can reproduce the same state
- divergence can be detected by comparing state hashes
- replays can be used to debug consistency issues

## Best practice

Do not rely on the latest local cache as proof that a node is in sync.

Use:
- event history
- state hashes
- snapshot cursors
- replay verification
- peer acknowledgements

Those are the real sync signals.
