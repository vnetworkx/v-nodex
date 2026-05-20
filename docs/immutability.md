# Immutability

The node treats records as append-only.

Rules:
- do not rewrite event history
- do not mutate stored hashes
- do not use cached state as truth
- always rebuild from history when verifying integrity
