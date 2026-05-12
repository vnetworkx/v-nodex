# Immutability

## Purpose

This document explains how v-nodex keeps history immutable and replayable.

Immutability is one of the central properties of the system.

## What immutability means here

In v-nodex, immutability means:

- accepted history is append-only
- records are not rewritten in place
- canonical state is derived, not hand-edited
- snapshots are sealed checkpoints, not mutable truth
- local caches may be replaced, but history may not

## Why immutability matters

Immutability gives the system:

- tamper evidence
- auditability
- deterministic replay
- recovery after failure
- peer-verifiable consistency
- confidence in state reconstruction

## The immutability chain

A record becomes hard to dispute through several mechanisms:

1. **Validation**  
   The event must satisfy protocol rules.

2. **Hashing**  
   The event gets a canonical identity.

3. **Signing**  
   The producing identity can be verified.

4. **Append-only persistence**  
   The accepted event is added to history without rewriting prior records.

5. **Replay**  
   The same history reconstructs the same state.

6. **Snapshot anchoring**  
   Checkpoints summarize accepted history without replacing it.

## Canonical record rule

Once an event is accepted into canonical history, it is part of the permanent replay chain.

If an event is later found to be invalid, the correct protocol response is not silent editing of history. The correct response is to reject it before append or quarantine it as a non-canonical artifact.

## Cache is not truth

A cache can be replaced, cleared, rebuilt, or discarded.

That includes:
- Redis entries
- live memory state
- temporary query projections
- debug buffers

These structures are allowed to be mutable because they are not canonical truth.

## Replay safety

Replay safety means:

- the same accepted history produces the same final state
- the same canonical hash set produces the same output
- the same rules produce the same reconstruction

If a replay changes unexpectedly, the node must treat that as a correctness issue.

## Snapshot immutability

A snapshot becomes immutable once it is sealed.

It should be treated as:
- a checkpoint
- a replay accelerator
- a verification artifact

It should not be treated as a replacement for the event history behind it.

## Peer immutability

A peer may disagree, disappear, or reconnect, but that does not change canonical history.

Peers are just replication participants. They do not own truth.

## Practical consequence

The practical consequence of immutability is that the node can always answer:

- what happened?
- in what order?
- under which hash?
- with which proof?
- what does replay produce?

That is the core trust model of the system.
