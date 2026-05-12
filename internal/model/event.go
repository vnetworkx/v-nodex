package model

import "time"

type VectorType string

const (
	VectorTypePosition VectorType = "POSITION"
	VectorTypeFree     VectorType = "FREE"
	VectorTypeBound    VectorType = "BOUND"
	VectorTypeUnit     VectorType = "UNIT"
	VectorTypeZero     VectorType = "ZERO"
)

type Operation string

const (
	OpCreate      Operation = "CREATE"
	OpCertify     Operation = "CERTIFY"
	OpTransfer    Operation = "TRANSFER"
	OpDrain       Operation = "DRAIN"
	OpProject     Operation = "PROJECT"
	OpReconstruct Operation = "RECONSTRUCT"
	OpQuery       Operation = "QUERY"
	OpRecord      Operation = "RECORD"
	OpAdd         Operation = "ADD"
	OpSubtract    Operation = "SUBTRACT"
	OpScale       Operation = "SCALE"
	OpNormalize   Operation = "NORMALIZE"
	OpRotate      Operation = "ROTATE"
	OpConstrain   Operation = "CONSTRAIN"
	OpCompose     Operation = "COMPOSE"
	OpNullify     Operation = "NULLIFY"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type EventCore struct {
	EventID                string     `json:"event_id"`
	ParentIDs              []string   `json:"parent_ids"`
	PrevHash               string     `json:"prev_hash"`
	SpaceID                string     `json:"space_id"`
	RegionID               string     `json:"region_id"`
	EntityID               string     `json:"entity_id"`
	TargetEntityID         string     `json:"target_entity_id"`
	VectorType             VectorType `json:"vector_type"`
	Operation              Operation  `json:"operation"`
	InputVector            Vector     `json:"input_vector"`
	OutputVector           Vector     `json:"output_vector"`
	MagnitudeBefore        float64    `json:"magnitude_before"`
	MagnitudeAfter         float64    `json:"magnitude_after"`
	DirectionBefore        Vector     `json:"direction_before"`
	DirectionAfter         Vector     `json:"direction_after"`
	SpaceCoordinatesBefore Vector     `json:"space_coordinates_before"`
	SpaceCoordinatesAfter  Vector     `json:"space_coordinates_after"`
	TimeCreated            time.Time  `json:"time_created"`
	LogicalOrder           uint64     `json:"logical_order"`
	SignerID               string     `json:"signer_id"`
	SignerPublicKey        string     `json:"signer_public_key"`
	ValidationProof        string     `json:"validation_proof"`
	Metadata               []KeyValue `json:"metadata"`
}

type Event struct {
	EventCore
	EventHash      string `json:"event_hash"`
	Signature      string `json:"signature"`
	Certified      bool   `json:"certified"`
	Accepted       bool   `json:"accepted"`
	RejectedReason string `json:"rejected_reason,omitempty"`
}

type EntityState struct {
	EntityID         string     `json:"entity_id"`
	SpaceID          string     `json:"space_id"`
	RegionID         string     `json:"region_id"`
	Vector           Vector     `json:"vector"`
	VectorType       VectorType `json:"vector_type"`
	Position         Vector     `json:"position"`
	Velocity         Vector     `json:"velocity"`
	Certified        bool       `json:"certified"`
	AuthRatio        float64    `json:"auth_ratio"`
	LastEventHash    string     `json:"last_event_hash"`
	LastLogicalOrder uint64     `json:"last_logical_order"`
	Metadata         []KeyValue `json:"metadata"`
}

type RegionState struct {
	RegionID             string   `json:"region_id"`
	SpaceID              string   `json:"space_id"`
	EntityIDs            []string `json:"entity_ids"`
	EventHashes          []string `json:"event_hashes"`
	NeighborRegionHashes []string `json:"neighbor_region_hashes"`
	SnapshotHash         string   `json:"snapshot_hash"`
	LogicalHeight        uint64   `json:"logical_height"`
}

type Snapshot struct {
	SnapshotID string    `json:"snapshot_id"`
	SpaceID    string    `json:"space_id"`
	Scope      string    `json:"scope"`
	RootHash   string    `json:"root_hash"`
	EventHash  string    `json:"event_hash"`
	StateRoot  LiveState `json:"state_root"`
	CreatedAt  time.Time `json:"created_at"`
	Sealed     bool      `json:"sealed"`
}

type LiveState struct {
	SpaceID     string                 `json:"space_id"`
	Entities    map[string]EntityState `json:"entities"`
	Regions     map[string]RegionState `json:"regions"`
	Heads       map[string]string      `json:"heads"`
	LatestHash  string                 `json:"latest_hash"`
	LatestOrder uint64                 `json:"latest_order"`
	Snapshot    *Snapshot              `json:"snapshot,omitempty"`
}

type PeerInfo struct {
	PeerID      string     `json:"peer_id"`
	Address     string     `json:"address"`
	RegionScope string     `json:"region_scope"`
	LastSeen    time.Time  `json:"last_seen"`
	Score       float64    `json:"score"`
	Metadata    []KeyValue `json:"metadata"`
}

type SyncEnvelope struct {
	NodeID   string    `json:"node_id"`
	HeadHash string    `json:"head_hash"`
	Events   []Event   `json:"events"`
	Snapshot *Snapshot `json:"snapshot,omitempty"`
	Cursor   uint64    `json:"cursor"`
}

type ConsensusDecision struct {
	RegionID     string `json:"region_id"`
	EventHash    string `json:"event_hash"`
	Quorum       int    `json:"quorum"`
	Acknowledged int    `json:"acknowledged"`
	Finalized    bool   `json:"finalized"`
	Reason       string `json:"reason,omitempty"`
}
