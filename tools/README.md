# Tools

This folder is reserved for helper tools, scripts, and auxiliary development utilities.

## What belongs here

Examples include:

- replay helpers
- event inspectors
- snapshot verification tools
- migration helpers
- local debugging scripts
- packaging utilities

## What does not belong here

Do not put core protocol logic here.

Core protocol logic belongs in:
- `internal/`
- `rust/kernel/`
- `sql/`
- `services/`
- `docs/`

## Tooling goal

The tools in this folder should help developers:
- inspect
- verify
- test
- replay
- debug
- package

They should not redefine the authoritative state model.
