# Developer Tools

## Purpose

This document explains the tools provided to help developers inspect, test, and operate v-nodex during development.

These tools are intended to make the protocol easier to understand, verify, and troubleshoot.

## Tooling goals

Developer tools should make it easy to:

- run a node locally
- inspect events and state
- verify replay behavior
- create and inspect snapshots
- inspect peers
- trace sync behavior
- diagnose protocol errors

## CLI

The main developer-facing command line tool is the node CLI.

Typical responsibilities:

- show runtime status
- list peers
- inspect recent events
- query entity or region state
- create snapshots
- support local development workflows

## Debug endpoints

The debug endpoints are intended for development and controlled inspection.

They are useful for:

- viewing live state
- checking derived values
- listing known peers
- examining recent history

These endpoints should remain read-focused and should not become a back door for state mutation.

## Replay tools

Replay tooling should answer:

- can the node reconstruct state from history alone?
- do snapshot hashes match replayed hashes?
- does the event order produce the expected final state?

Recommended replay tooling tasks:

- load event history
- rebuild live state
- compare hash outputs
- print divergence information if hashes differ

## Snapshot tools

Snapshot tools are useful for:

- verifying checkpoint integrity
- confirming that recovery is possible
- making replay faster in large histories

A good snapshot tool should show:

- snapshot cursor
- snapshot hash
- associated region or state scope
- time created
- verification result

## Event inspection

Developer tools should allow inspection of:
- event ID
- parent relationship
- operation type
- input/output vectors
- signatures
- validation status
- replay position

This is essential for debugging deterministic systems.

## Suggested operator workflow

A practical local workflow looks like:

1. start the stack
2. submit a small set of events
3. inspect the resulting state
4. create a snapshot
5. restart the node
6. confirm it reconstructs correctly
7. check peer sync after reconnecting a second node

## Tool design principle

Developer tools should never be treated as the source of truth.

Their purpose is to:
- expose
- inspect
- verify
- assist

They are not supposed to override canonical history or bypass validation.

## When to extend tooling

Add tooling when you need:

- better replay visibility
- richer state inspection
- easier local testing
- snapshot validation
- peer debugging
- region debugging

The better the tooling, the easier it becomes to trust the protocol implementation.
