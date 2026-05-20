# DAG

The DAG is the node-side ancestry index over immutable records.

It provides:
- parent lookups
- head tracking
- topological ordering
- duplicate suppression
- cycle detection

The canonical event semantics still live in Rust.
