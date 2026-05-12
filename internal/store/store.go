package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"vnodex/internal/model"
)

type Store struct {
	root      string
	mu        sync.RWMutex
	events    map[string]model.Event
	ordered   []string
	state     model.LiveState
	peers     map[string]model.PeerInfo
	snapshots []model.Snapshot
}

func New(root string) (*Store, error) {
	if root == "" {
		root = "./data"
	}
	dirs := []string{
		filepath.Join(root, "events"),
		filepath.Join(root, "state"),
		filepath.Join(root, "snapshots"),
		filepath.Join(root, "indexes"),
		filepath.Join(root, "peers"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	s := &Store{
		root:   root,
		events: make(map[string]model.Event),
		peers:  make(map[string]model.PeerInfo),
		state: model.LiveState{
			Entities: make(map[string]model.EntityState),
			Regions:  make(map[string]model.RegionState),
			Heads:    make(map[string]string),
		},
	}
	if err := s.loadAll(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Root() string { return s.root }

func (s *Store) loadAll() error {
	if err := s.loadPeers(); err != nil {
		return err
	}
	if err := s.loadEvents(); err != nil {
		return err
	}
	if err := s.Rebuild(); err != nil {
		return err
	}
	if err := s.loadSnapshots(); err != nil {
		return err
	}
	return nil
}

func (s *Store) loadPeers() error {
	base := filepath.Join(s.root, "peers")
	return filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		var peer model.PeerInfo
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &peer); err != nil {
			return err
		}
		if peer.PeerID != "" {
			s.peers[peer.PeerID] = peer
		}
		return nil
	})
}

func (s *Store) loadEvents() error {
	base := filepath.Join(s.root, "events")
	return filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var event model.Event
		if err := json.Unmarshal(b, &event); err != nil {
			return err
		}
		s.events[event.EventHash] = event
		s.ordered = append(s.ordered, event.EventHash)
		return nil
	})
}

func (s *Store) loadSnapshots() error {
	base := filepath.Join(s.root, "snapshots")
	return filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".json") {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var snap model.Snapshot
		if err := json.Unmarshal(b, &snap); err != nil {
			return err
		}
		s.snapshots = append(s.snapshots, snap)
		return nil
	})
}

func (s *Store) Rebuild() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = model.LiveState{Entities: map[string]model.EntityState{}, Regions: map[string]model.RegionState{}, Heads: map[string]string{}}
	ordered := make([]model.Event, 0, len(s.events))
	for _, e := range s.events {
		ordered = append(ordered, e)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if !ordered[i].TimeCreated.Equal(ordered[j].TimeCreated) {
			return ordered[i].TimeCreated.Before(ordered[j].TimeCreated)
		}
		if ordered[i].LogicalOrder != ordered[j].LogicalOrder {
			return ordered[i].LogicalOrder < ordered[j].LogicalOrder
		}
		return ordered[i].EventHash < ordered[j].EventHash
	})
	for _, e := range ordered {
		if err := s.applyEventLocked(e); err != nil {
			return err
		}
	}
	return s.persistLiveStateLocked()
}

func (s *Store) Append(event model.Event) (model.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.TimeCreated.IsZero() {
		event.TimeCreated = time.Now().UTC()
	}
	event.Metadata = model.CanonicalizeMetadata(event.Metadata)
	computedHash, _, err := model.CanonicalEventHash(event.EventCore)
	if err != nil {
		return model.Event{}, err
	}
	if event.EventHash == "" {
		event.EventHash = computedHash
	}
	if event.EventHash != computedHash {
		return model.Event{}, fmt.Errorf("event hash mismatch: got %s want %s", event.EventHash, computedHash)
	}
	if err := model.ValidateEvent(event); err != nil {
		return model.Event{}, err
	}
	if _, exists := s.events[event.EventHash]; exists {
		return event, nil
	}
	if err := s.applyEventLocked(event); err != nil {
		return model.Event{}, err
	}
	if err := s.persistEventLocked(event); err != nil {
		return model.Event{}, err
	}
	s.events[event.EventHash] = event
	s.ordered = append(s.ordered, event.EventHash)
	if err := s.persistLiveStateLocked(); err != nil {
		return model.Event{}, err
	}
	return event, nil
}

func (s *Store) applyEventLocked(event model.Event) error {
	if event.ParentIDs != nil {
		for _, pid := range event.ParentIDs {
			if pid == "" {
				continue
			}
			if _, ok := s.events[pid]; !ok {
				return fmt.Errorf("missing parent event %s", pid)
			}
		}
	}
	entity := s.state.Entities[event.EntityID]
	if entity.EntityID == "" {
		entity = model.EntityState{EntityID: event.EntityID, SpaceID: event.SpaceID, RegionID: event.RegionID, VectorType: event.VectorType, Metadata: []model.KeyValue{}}
	}
	before := entity.Vector.Clone()
	after := applyOperation(before, event)
	entity.SpaceID = event.SpaceID
	entity.RegionID = event.RegionID
	entity.VectorType = event.VectorType
	entity.Position = entity.Position.Clone()
	if len(event.SpaceCoordinatesAfter) > 0 {
		entity.Position = event.SpaceCoordinatesAfter.Clone()
	}
	if len(event.OutputVector) > 0 {
		entity.Vector = event.OutputVector.Clone()
	} else {
		entity.Vector = after.Clone()
	}
	if len(event.InputVector) > 0 && event.Operation == model.OpTransfer && event.TargetEntityID != "" {
		target := s.state.Entities[event.TargetEntityID]
		if target.EntityID == "" {
			target = model.EntityState{EntityID: event.TargetEntityID, SpaceID: event.SpaceID, RegionID: event.RegionID, VectorType: event.VectorType, Metadata: []model.KeyValue{}}
		}
		target.Vector = target.Vector.Add(event.OutputVector)
		target.SpaceID = event.SpaceID
		target.RegionID = event.RegionID
		s.state.Entities[event.TargetEntityID] = target
	}
	entity.Velocity = entity.Vector.Sub(before)
	entity.Certified = event.Certified
	entity.LastEventHash = event.EventHash
	entity.LastLogicalOrder = event.LogicalOrder
	s.state.Entities[event.EntityID] = entity

	region := s.state.Regions[event.RegionID]
	region.RegionID = event.RegionID
	region.SpaceID = event.SpaceID
	region.EventHashes = appendUnique(region.EventHashes, event.EventHash)
	region.EntityIDs = appendUnique(region.EntityIDs, event.EntityID)
	region.LogicalHeight = max64(region.LogicalHeight, event.LogicalOrder)
	s.state.Regions[event.RegionID] = region
	s.state.Heads[event.EntityID] = event.EventHash
	s.state.LatestHash = event.EventHash
	s.state.LatestOrder = max64(s.state.LatestOrder, event.LogicalOrder)
	return nil
}

func applyOperation(before model.Vector, event model.Event) model.Vector {
	switch event.Operation {
	case model.OpCreate:
		if len(event.OutputVector) > 0 {
			return event.OutputVector.Clone()
		}
		return before
	case model.OpTransfer, model.OpSubtract:
		return before.Sub(event.InputVector)
	case model.OpDrain:
		return before.Scale(1.0 - clamp(event.MagnitudeAfter, 0, 1))
	case model.OpProject:
		return event.OutputVector.Clone()
	case model.OpReconstruct, model.OpQuery, model.OpRecord:
		return before
	case model.OpAdd, model.OpCompose:
		return before.Add(event.InputVector)
	case model.OpScale:
		return before.Scale(clamp(event.MagnitudeAfter, -1e9, 1e9))
	case model.OpNormalize:
		if out, err := before.Normalize(); err == nil {
			return out
		}
		return before
	case model.OpRotate:
		if out, err := before.Rotate2D(event.MagnitudeAfter); err == nil {
			return out
		}
		return before
	case model.OpConstrain:
		return before.Constrain(model.ZeroVector(len(before)), event.OutputVector)
	case model.OpNullify:
		return before.Nullify()
	default:
		if len(event.OutputVector) > 0 {
			return event.OutputVector.Clone()
		}
		return before
	}
}

func (s *Store) persistEventLocked(event model.Event) error {
	path := s.eventPath(event.EventHash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeJSONAtomic(path, event)
}

func (s *Store) persistLiveStateLocked() error {
	path := filepath.Join(s.root, "state", "current.json")
	state := cloneState(s.state)
	state.Snapshot = nil
	return writeJSONAtomic(path, state)
}

func (s *Store) eventPath(hash string) string {
	if len(hash) < 4 {
		hash = fmt.Sprintf("%04s", hash)
	}
	return filepath.Join(s.root, "events", hash[:2], hash[2:4], hash+".json")
}

func (s *Store) GetEvent(hash string) (model.Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ev, ok := s.events[hash]
	return ev, ok
}

func (s *Store) ListEvents(limit int) []model.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ordered := make([]model.Event, 0, len(s.ordered))
	for _, hash := range s.ordered {
		if ev, ok := s.events[hash]; ok {
			ordered = append(ordered, ev)
		}
	}
	if limit > 0 && len(ordered) > limit {
		return ordered[len(ordered)-limit:]
	}
	return ordered
}

func (s *Store) State() model.LiveState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneState(s.state)
}

func (s *Store) Entity(entityID string) (model.EntityState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.state.Entities[entityID]
	return e, ok
}

func (s *Store) Region(regionID string) (model.RegionState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.state.Regions[regionID]
	return r, ok
}

func (s *Store) Snapshot(scope string) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := model.Snapshot{
		SnapshotID: fmt.Sprintf("snap-%d-%s", time.Now().UTC().UnixNano(), shortHash(scope)),
		SpaceID:    s.state.SpaceID,
		Scope:      scope,
		RootHash:   s.state.LatestHash,
		EventHash:  s.state.LatestHash,
		StateRoot:  cloneState(s.state),
		CreatedAt:  time.Now().UTC(),
		Sealed:     true,
	}
	snap.StateRoot.Snapshot = &snap
	path := filepath.Join(s.root, "snapshots", snap.SnapshotID+".json")
	if err := writeJSONAtomic(path, snap); err != nil {
		return model.Snapshot{}, err
	}
	s.snapshots = append(s.snapshots, snap)
	if region, ok := s.state.Regions[scope]; ok {
		region.SnapshotHash = snap.SnapshotID
		s.state.Regions[scope] = region
		_ = s.persistLiveStateLocked()
	}
	return snap, nil
}

func (s *Store) Snapshots() []model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Snapshot, len(s.snapshots))
	copy(out, s.snapshots)
	return out
}

func (s *Store) Peers() []model.PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.PeerInfo, 0, len(s.peers))
	for _, p := range s.peers {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PeerID < out[j].PeerID })
	return out
}

func (s *Store) RegisterPeer(peer model.PeerInfo) error {
	if peer.PeerID == "" {
		return errors.New("peer_id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	peer.LastSeen = time.Now().UTC()
	s.peers[peer.PeerID] = peer
	return writeJSONAtomic(filepath.Join(s.root, "peers", peer.PeerID+".json"), peer)
}

func (s *Store) EventsSince(cursor uint64, limit int) []model.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ordered := make([]model.Event, 0)
	for _, hash := range s.ordered {
		ev := s.events[hash]
		if ev.LogicalOrder > cursor {
			ordered = append(ordered, ev)
		}
	}
	if limit > 0 && len(ordered) > limit {
		return ordered[:limit]
	}
	return ordered
}

func (s *Store) RejectedEvent(reason string, event model.Event) model.Event {
	event.Accepted = false
	event.RejectedReason = reason
	return event
}

func cloneState(in model.LiveState) model.LiveState {
	out := model.LiveState{
		SpaceID:     in.SpaceID,
		Entities:    make(map[string]model.EntityState, len(in.Entities)),
		Regions:     make(map[string]model.RegionState, len(in.Regions)),
		Heads:       make(map[string]string, len(in.Heads)),
		LatestHash:  in.LatestHash,
		LatestOrder: in.LatestOrder,
	}
	for k, v := range in.Entities {
		out.Entities[k] = cloneEntity(v)
	}
	for k, v := range in.Regions {
		out.Regions[k] = cloneRegion(v)
	}
	for k, v := range in.Heads {
		out.Heads[k] = v
	}
	// Snapshot pointers are intentionally not deep-copied into persisted cache state
	// to avoid recursive serialization cycles.
	out.Snapshot = nil
	return out
}

func cloneEntity(in model.EntityState) model.EntityState {
	out := in
	out.Vector = in.Vector.Clone()
	out.Position = in.Position.Clone()
	out.Velocity = in.Velocity.Clone()
	if len(in.Metadata) > 0 {
		out.Metadata = append([]model.KeyValue(nil), in.Metadata...)
	}
	return out
}

func cloneRegion(in model.RegionState) model.RegionState {
	out := in
	out.EntityIDs = append([]string(nil), in.EntityIDs...)
	out.EventHashes = append([]string(nil), in.EventHashes...)
	out.NeighborRegionHashes = append([]string(nil), in.NeighborRegionHashes...)
	return out
}

func writeJSONAtomic(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

func max64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:6])
}

func (s *Store) rebuildState() error {
	return s.Rebuild()
}
