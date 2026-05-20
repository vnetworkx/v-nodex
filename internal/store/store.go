package store

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "sync"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Store struct {
    dir           string
    eventsPath    string
    statePath     string
    peersPath     string
    snapshotsPath string
    mutex         sync.RWMutex
}

func Open(dataDir string) (*Store, error) {
    if dataDir == "" {
        return nil, errors.New("data directory required")
    }
    if err := os.MkdirAll(dataDir, 0o755); err != nil {
        return nil, err
    }
    s := &Store{
        dir:           dataDir,
        eventsPath:    filepath.Join(dataDir, "events.jsonl"),
        statePath:     filepath.Join(dataDir, "state.json"),
        peersPath:     filepath.Join(dataDir, "peers.json"),
        snapshotsPath: filepath.Join(dataDir, "snapshots.jsonl"),
    }
    if err := s.ensureFiles(); err != nil {
        return nil, err
    }
    return s, nil
}

func (s *Store) Dir() string { return s.dir }

func (s *Store) ensureFiles() error {
    files := []string{s.eventsPath, s.statePath, s.peersPath, s.snapshotsPath}
    for _, f := range files {
        if _, err := os.Stat(f); os.IsNotExist(err) {
            switch {
            case strings.HasSuffix(f, ".jsonl"):
                if err := os.WriteFile(f, []byte{}, 0o600); err != nil {
                    return err
                }
            case strings.HasSuffix(f, "state.json"):
                if err := os.WriteFile(f, []byte(`{"node_id":"","state_root":"","latest_clock":0,"event_count":0,"heads":[],"derived_state":{},"updated_at":"0001-01-01T00:00:00Z"}`), 0o600); err != nil {
                    return err
                }
            case strings.HasSuffix(f, "peers.json"):
                if err := os.WriteFile(f, []byte("[]"), 0o600); err != nil {
                    return err
                }
            }
        } else if err != nil {
            return err
        }
    }
    return nil
}

func (s *Store) Init(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return nil
    }
}

func (s *Store) AppendEvent(ctx context.Context, rec model.EventRecord) error {
    exists, err := s.ExistsEventHash(ctx, rec.EventHash)
    if err != nil {
        return err
    }
    if exists {
        return nil
    }

    s.mutex.Lock()
    defer s.mutex.Unlock()

    f, err := os.OpenFile(s.eventsPath, os.O_APPEND|os.O_WRONLY, 0o600)
    if err != nil {
        return err
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    if err := enc.Encode(rec); err != nil {
        return err
    }
    return nil
}

func (s *Store) ExistsEventHash(ctx context.Context, hash string) (bool, error) {
    events, err := s.LoadEvents(ctx)
    if err != nil {
        return false, err
    }
    for _, e := range events {
        if e.EventHash == hash {
            return true, nil
        }
    }
    return false, nil
}

func (s *Store) LoadEvents(ctx context.Context) ([]model.EventRecord, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    f, err := os.Open(s.eventsPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    out := []model.EventRecord{}
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        var rec model.EventRecord
        if err := json.Unmarshal([]byte(line), &rec); err != nil {
            return nil, fmt.Errorf("decode event: %w", err)
        }
        out = append(out, rec)
    }
    return out, scanner.Err()
}

func (s *Store) ListEvents(ctx context.Context, limit, offset int) ([]model.EventRecord, error) {
    events, err := s.LoadEvents(ctx)
    if err != nil {
        return nil, err
    }
    sort.SliceStable(events, func(i, j int) bool {
        if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
            return events[i].CreatedAt.After(events[j].CreatedAt)
        }
        if events[i].LogicalClock != events[j].LogicalClock {
            return events[i].LogicalClock > events[j].LogicalClock
        }
        if events[i].EventHash != events[j].EventHash {
            return events[i].EventHash < events[j].EventHash
        }
        return events[i].EventID < events[j].EventID
    })
    if offset > len(events) {
        return []model.EventRecord{}, nil
    }
    events = events[offset:]
    if limit > 0 && limit < len(events) {
        events = events[:limit]
    }
    return events, nil
}

func (s *Store) LastEvent(ctx context.Context) (*model.EventRecord, error) {
    events, err := s.LoadEvents(ctx)
    if err != nil {
        return nil, err
    }
    if len(events) == 0 {
        return nil, nil
    }
    sort.SliceStable(events, func(i, j int) bool {
        if events[i].LogicalClock != events[j].LogicalClock {
            return events[i].LogicalClock < events[j].LogicalClock
        }
        if !events[i].CreatedAt.Equal(events[j].CreatedAt) {
            return events[i].CreatedAt.Before(events[j].CreatedAt)
        }
        return events[i].EventHash < events[j].EventHash
    })
    last := events[len(events)-1]
    return &last, nil
}

func (s *Store) GetState(ctx context.Context) (model.StateView, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    b, err := os.ReadFile(s.statePath)
    if err != nil {
        return model.StateView{}, err
    }
    var st model.StateView
    if err := json.Unmarshal(b, &st); err != nil {
        return model.StateView{}, err
    }
    return st, nil
}

func (s *Store) UpsertState(ctx context.Context, st model.StateView) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    st.UpdatedAt = time.Now().UTC()
    return atomicWriteJSON(s.statePath, st)
}

func (s *Store) ListPeers(ctx context.Context) ([]model.Peer, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    b, err := os.ReadFile(s.peersPath)
    if err != nil {
        return nil, err
    }
    var peers []model.Peer
    if err := json.Unmarshal(b, &peers); err != nil {
        return nil, err
    }
    return peers, nil
}

func (s *Store) UpsertPeer(ctx context.Context, p model.Peer) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    peers, err := s.loadPeersLocked()
    if err != nil {
        return err
    }
    found := false
    for i := range peers {
        if peers[i].PeerID == p.PeerID {
            peers[i] = p
            found = true
            break
        }
    }
    if !found {
        peers = append(peers, p)
    }
    return atomicWriteJSON(s.peersPath, peers)
}

func (s *Store) RemovePeer(ctx context.Context, peerID string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    peers, err := s.loadPeersLocked()
    if err != nil {
        return err
    }
    filtered := make([]model.Peer, 0, len(peers))
    for _, p := range peers {
        if p.PeerID != peerID {
            filtered = append(filtered, p)
        }
    }
    return atomicWriteJSON(s.peersPath, filtered)
}

func (s *Store) AppendSnapshot(ctx context.Context, snap model.Snapshot) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    f, err := os.OpenFile(s.snapshotsPath, os.O_APPEND|os.O_WRONLY, 0o600)
    if err != nil {
        return err
    }
    defer f.Close()
    return json.NewEncoder(f).Encode(snap)
}

func (s *Store) ListSnapshots(ctx context.Context) ([]model.Snapshot, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    f, err := os.Open(s.snapshotsPath)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    var snaps []model.Snapshot
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        var snap model.Snapshot
        if err := json.Unmarshal([]byte(line), &snap); err != nil {
            return nil, err
        }
        snaps = append(snaps, snap)
    }
    return snaps, scanner.Err()
}

func (s *Store) LatestSnapshot(ctx context.Context) (*model.Snapshot, error) {
    snaps, err := s.ListSnapshots(ctx)
    if err != nil {
        return nil, err
    }
    if len(snaps) == 0 {
        return nil, nil
    }
    sort.SliceStable(snaps, func(i, j int) bool {
        if !snaps[i].CreatedAt.Equal(snaps[j].CreatedAt) {
            return snaps[i].CreatedAt.Before(snaps[j].CreatedAt)
        }
        return snaps[i].SnapshotID < snaps[j].SnapshotID
    })
    snap := snaps[len(snaps)-1]
    return &snap, nil
}

func (s *Store) loadPeersLocked() ([]model.Peer, error) {
    b, err := os.ReadFile(s.peersPath)
    if err != nil {
        return nil, err
    }
    var peers []model.Peer
    if err := json.Unmarshal(b, &peers); err != nil {
        return nil, err
    }
    return peers, nil
}

func atomicWriteJSON(path string, v any) error {
    tmp := path + ".tmp"
    b, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return err
    }
    if err := os.WriteFile(tmp, b, 0o600); err != nil {
        return err
    }
    return os.Rename(tmp, path)
}
