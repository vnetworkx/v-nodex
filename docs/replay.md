# Replay

Replay is deterministic reconstruction from immutable records.

The node replay engine:
- sorts records canonically
- validates ancestry
- computes history roots
- verifies state root stability
- triggers snapshot sealing
