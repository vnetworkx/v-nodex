package model

import (
    "encoding/json"
    "time"
)

type OperationType string

const (
    OpOriginCreate  OperationType = "ORIGIN_CREATE"
    OpTransfer      OperationType = "TRANSFER"
    OpDrain        OperationType = "DRAIN"
    OpProject      OperationType = "PROJECT"
    OpReconstruct  OperationType = "RECONSTRUCT"
    OpQuery       OperationType = "QUERY"
    OpReplay      OperationType = "REPLAY"
    OpSnapshot    OperationType = "SNAPSHOT"
    OpPeerIngest  OperationType = "PEER_INGEST"
)

type EventRequest struct {
    EventID      string          `json:"event_id"`
    ParentHashes []string        `json:"parent_hashes,omitempty"`
    RegionID     string          `json:"region_id"`
    EntityID     string          `json:"entity_id"`
    Operation    OperationType    `json:"operation"`
    Params       json.RawMessage `json:"params,omitempty"`
    ActorPK      string          `json:"actor_pk"`
    Signature    string          `json:"signature,omitempty"`
}

type EventRecord struct {
    EventID      string          `json:"event_id"`
    ParentHashes []string        `json:"parent_hashes,omitempty"`
    RegionID     string          `json:"region_id"`
    EntityID     string          `json:"entity_id"`
    Operation    OperationType    `json:"operation"`
    Params       json.RawMessage `json:"params,omitempty"`
    ActorPK      string          `json:"actor_pk"`
    Signature    string          `json:"signature,omitempty"`
    EventHash    string          `json:"event_hash"`
    PayloadHash  string          `json:"payload_hash"`
    StateRoot    string          `json:"state_root"`
    RegionRoot   string          `json:"region_root"`
    LogicalClock int64           `json:"logical_clock"`
    Accepted     bool            `json:"accepted"`
    Certified    bool            `json:"certified"`
    Reason       string          `json:"reason,omitempty"`
    CreatedAt    time.Time       `json:"created_at"`
}

type VectorState struct {
    WalletID   string          `json:"wallet_id,omitempty"`
    PublicKey  string          `json:"public_key,omitempty"`
    SpaceID    string          `json:"space_id,omitempty"`
    Dimensions json.RawMessage `json:"dimensions,omitempty"`
}

type StateView struct {
    NodeID      string          `json:"node_id"`
    StateRoot   string          `json:"state_root"`
    LatestClock  int64           `json:"latest_clock"`
    EventCount   int64           `json:"event_count"`
    Heads       []string        `json:"heads,omitempty"`
    DerivedState json.RawMessage `json:"derived_state,omitempty"`
    UpdatedAt    time.Time       `json:"updated_at"`
}

type Peer struct {
    PeerID    string    `json:"peer_id"`
    BaseURL   string    `json:"base_url,omitempty"`
    Multiaddr string    `json:"multiaddr,omitempty"`
    Healthy   bool      `json:"healthy"`
    Score     int       `json:"score"`
    LastSeen  time.Time `json:"last_seen"`
}

type Snapshot struct {
    SnapshotID   string          `json:"snapshot_id"`
    StateRoot    string          `json:"state_root"`
    HistoryRoot  string          `json:"history_root,omitempty"`
    HeadHashes   []string        `json:"head_hashes"`
    RegionRoot   string          `json:"region_root"`
    LogicalClock int64           `json:"logical_clock"`
    Payload      json.RawMessage `json:"payload,omitempty"`
    CreatedAt    time.Time       `json:"created_at"`
}

type KernelStatus struct {
    OK      bool   `json:"ok"`
    Version string `json:"version,omitempty"`
}

type ReplayReport struct {
    EventCount    int           `json:"event_count"`
    LatestClock   int64         `json:"latest_clock"`
    StateRoot     string        `json:"state_root"`
    HistoryRoot   string        `json:"history_root"`
    CanonicalHash string        `json:"canonical_hash"`
    Heads         []string      `json:"heads"`
    OrderedHashes []string      `json:"ordered_hashes"`
    Verified      bool          `json:"verified"`
    StartedAt     time.Time     `json:"started_at"`
    FinishedAt    time.Time     `json:"finished_at"`
}

type SnapshotReport struct {
    Snapshot Snapshot `json:"snapshot"`
    Verified bool     `json:"verified"`
}

type SyncSummary struct {
    PeersVisited    int    `json:"peers_visited"`
    EventsFetched   int    `json:"events_fetched"`
    EventsImported  int    `json:"events_imported"`
    SnapshotsPulled  int    `json:"snapshots_pulled"`
    StateRoot       string `json:"state_root"`
    HistoryRoot     string `json:"history_root"`
}
