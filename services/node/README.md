# vnodex-node

Node runtime entrypoint.

This binary:
- loads config
- starts the kernel bridge
- opens the store
- launches the HTTP API
- starts sync and snapshot orchestration

Example:
```bash
go run ./services/node/cmd/vnodex-node --config configs/node.yaml
```
