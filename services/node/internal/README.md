# Node Internal Notes

This folder is reserved for node-service-specific internal code if needed later.

## Current role

In the current scaffold, shared functionality lives under `internal/` at the repository root.

That shared location is used so that:
- the node service
- the API service
- the CLI
- and other tooling

can rely on the same underlying protocol and storage logic.

## When to add code here

Add node-specific internal code here only if it is not reusable by the rest of the repo.

Good candidates would be:
- node-local orchestration helpers
- service-specific startup helpers
- node-only diagnostics
- node-only configuration adapters

## Guiding rule

Prefer shared protocol code in the root `internal/` packages whenever possible.
