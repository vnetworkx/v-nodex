package snapshot

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/replay"
    "github.com/vnetworkx/v-nodex/internal/store"
)

type Manager struct {
    store  *store.Store
    replay *replay.Engine
}

func New(store *store.Store, replay *replay.Engine) *Manager {
    return &Manager{store: store, replay: replay}
}

func (m *Manager) Create(ctx context.Context, state model.StateView) (model.Snapshot, error) {
    report, _, err := m.replay.Replay(ctx)
    if err != nil {
        return model.Snapshot{}, err
    }
    snap := model.Snapshot{
        SnapshotID:   fmt.Sprintf("snapshot-%d", time.Now().UTC().UnixNano()),
        StateRoot:    report.StateRoot,
        HistoryRoot:  report.HistoryRoot,
        HeadHashes:   report.Heads,
        RegionRoot:   "",
        LogicalClock: report.LatestClock,
        CreatedAt:    time.Now().UTC(),
    }
    payload := map[string]any{
        "state":  state,
        "report": report,
        "sealed": true,
    }
    b, _ := json.Marshal(payload)
    snap.Payload = b

    if err := m.store.AppendSnapshot(ctx, snap); err != nil {
        return model.Snapshot{}, err
    }
    if err := m.writeSnapshotFile(snap); err != nil {
        return model.Snapshot{}, err
    }
    return snap, nil
}

func (m *Manager) Latest(ctx context.Context) (*model.Snapshot, error) {
    return m.store.LatestSnapshot(ctx)
}

func (m *Manager) Verify(ctx context.Context) (bool, error) {
    latest, err := m.store.LatestSnapshot(ctx)
    if err != nil || latest == nil {
        return false, err
    }
    report, _, err := m.replay.Replay(ctx)
    if err != nil {
        return false, err
    }
    return latest.StateRoot == report.StateRoot && latest.HistoryRoot == report.HistoryRoot, nil
}

func (m *Manager) writeSnapshotFile(snap model.Snapshot) error {
    dir := filepath.Join(m.store.Dir(), "snapshots")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    path := filepath.Join(dir, snap.SnapshotID+".json")
    b, _ := json.MarshalIndent(snap, "", "  ")
    return os.WriteFile(path, b, 0o600)
}
