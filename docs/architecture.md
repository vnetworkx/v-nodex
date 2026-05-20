# v-nodex Architecture

v-nodex is the node/runtime layer for the Vector Network.

It handles networking, persistence, APIs, sync, snapshots, and replay orchestration.
It does not define protocol truth.

## Authority split

- Rust `v-kernelx`: validation, hashing, execution, replay, state roots
- Go `v-nodex`: transport, persistence, indexing, operator tooling

## Event flow

Ingress -> preflight -> kernel -> canonical record -> persistence -> gossip -> replay verification

## Stability rule

The node may not independently redefine protocol semantics.
