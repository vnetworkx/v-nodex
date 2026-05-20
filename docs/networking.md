# Networking

The architecture reserves libp2p and QUIC for future transport scaling.

The current rebuilt node keeps a transport abstraction and HTTP-based peer sync hooks so the protocol can be exercised immediately while the kernel remains authoritative.
