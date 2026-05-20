# v-nodex

v-nodex is the Go node/runtime layer for the Vector Network.

It is responsible for:

- config loading
- peer networking
- record transport
- local persistence
- replay orchestration
- snapshot management
- HTTP APIs
- operator tooling

Rust `v-kernelx` remains the protocol authority.

## Current process

1. Go receives an event or peer record.
2. Go forwards protocol decisions to Rust kernel.
3. Rust returns canonical records.
4. Go persists records, updates derived views, and syncs peers.

## Quick start

```bash
cp configs/node.yaml.example configs/node.yaml
go run ./services/node/cmd/vnodex-node --config configs/node.yaml
```

The node uses a local append-only store by default and talks to the Rust kernel over HTTP.

## Layout

See `docs/architecture.md` for the updated canonical architecture.
