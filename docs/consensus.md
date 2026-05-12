# Consensus

## Purpose

This document explains how v-nodex decides which data is valid and which history is accepted.

The consensus layer is intentionally lightweight in the current scaffold. Its job is to support deterministic acceptance, regional verification, and peer-based convergence without turning the system into a generic account chain.

## Consensus goal

Consensus must ensure that:

- only valid events enter canonical history
- all honest nodes can reconstruct the same state
- conflicting or malformed records are rejected
- recovery after partition is deterministic

## What consensus is in v-nodex

Consensus is the agreement process over:

- event validity
- event ordering
- parent linkage
- replayable state
- snapshot compatibility
- peer-reported heads

It is not only a voting layer. It is also a deterministic verification layer.

## What consensus is not

Consensus is not:

- hidden admin mutation
- arbitrary write permission
- non-deterministic merge logic
- a freeform conflict resolver that bypasses validation

## Acceptance conditions

An event should only be accepted if it satisfies the protocol’s validation rules, including:

- structural correctness
- signature validity if required
- hash consistency
- parent existence and linkage rules
- operation legality
- policy compliance
- zero-vector safety
- replay consistency

## Local finality

In the current architecture, a node can treat accepted and persisted history as locally final once it passes validation and is appended to the canonical store.

If you later add stronger quorum logic, it should build on top of this deterministic base rather than replace it.

## Regional consensus concept

The spatial storage design supports the idea of regional consensus.

That means:
- a subset of peers may validate a local region
- neighboring regions may exchange proofs and events
- consensus can be scoped to a region or partition
- global behavior emerges from consistent local validation

## Why deterministic replay matters for consensus

Replay is the simplest way to verify consensus.

If two nodes have:
- the same genesis
- the same accepted events
- the same validation logic

then they must derive the same state.

If they do not, consensus has failed.

## Conflict handling

When peers disagree, the node should not silently merge incompatible histories.

Instead it should:
- reject invalid input
- quarantine inconsistent input
- compare hashes and heads
- request missing history
- re-evaluate after replay

## Membership and trust

The project can support different trust models over time:

- open peer validation
- curated validator sets
- regional quorum groups
- stronger signature policies
- threshold validation in future extensions

The current scaffold keeps the consensus surface explicit so those models can be added later without breaking the core storage model.

## Operational perspective

From an operator’s perspective, consensus means the node can answer:

- what history do I accept?
- what event order is canonical?
- which peers do I trust enough to sync from?
- does my replayed state match the expected hash?

## Future direction

If the network later adds more formal quorums or stronger regional voting, the rules should still preserve these invariants:

- deterministic acceptance
- append-only canonical history
- replayable state
- verifiable convergence

That is the consensus contract for v-nodex.
