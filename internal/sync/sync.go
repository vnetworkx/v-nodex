package sync

import (
    "context"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/p2p"
)

type EventStore interface {
    LoadEvents(ctx context.Context) ([]model.EventRecord, error)
    AppendEvent(ctx context.Context, rec model.EventRecord) error
    ExistsEventHash(ctx context.Context, hash string) (bool, error)
    ListPeers(ctx context.Context) ([]model.Peer, error)
    UpsertPeer(ctx context.Context, p model.Peer) error
    ListSnapshots(ctx context.Context) ([]model.Snapshot, error)
    AppendSnapshot(ctx context.Context, snap model.Snapshot) error
}

type StateReader interface {
    State(ctx context.Context) (model.StateView, error)
    UpsertState(ctx context.Context, st model.StateView) error
}

type Importer interface {
    ImportRecord(ctx context.Context, rec model.EventRecord) error
}

type ReplayProvider interface {
    Replay(ctx context.Context) (model.ReplayReport, []model.EventRecord, error)
}

type Manager struct {
    Store   EventStore
    State   StateReader
    Import  Importer
    Replay  ReplayProvider
    Mesh    *p2p.Mesh
    Me      model.Peer
    Limits  Limits
}

type Limits struct {
    PullBatch      int
    PushBatch      int
    Interval       time.Duration
    MaxConcurrent  int
}

func New(store EventStore, state StateReader, importer Importer, replay ReplayProvider, mesh *p2p.Mesh, me model.Peer, limits Limits) *Manager {
    return &Manager{
        Store: store, State: state, Import: importer, Replay: replay, Mesh: mesh, Me: me, Limits: limits,
    }
}

func (m *Manager) SyncOnce(ctx context.Context) (model.SyncSummary, error) {
    peers, err := m.Store.ListPeers(ctx)
    if err != nil {
        return model.SyncSummary{}, err
    }
    summary := model.SyncSummary{}
    for _, peer := range peers {
        summary.PeersVisited++
        if peer.BaseURL == "" {
            continue
        }
        events, err := m.Mesh.SyncPeer(ctx, peer, m.Limits.PullBatch)
        if err != nil {
            continue
        }
        summary.EventsFetched += len(events)
        for _, ev := range events {
            exists, _ := m.Store.ExistsEventHash(ctx, ev.EventHash)
            if exists {
                continue
            }
            if err := m.Import.ImportRecord(ctx, ev); err == nil {
                summary.EventsImported++
            }
        }
        snaps, err := m.Mesh.Transport.FetchSnapshots(ctx, peer.BaseURL)
        if err == nil {
            summary.SnapshotsPulled += len(snaps)
        }
    }
    report, _, err := m.Replay.Replay(ctx)
    if err == nil {
        summary.StateRoot = report.StateRoot
        summary.HistoryRoot = report.HistoryRoot
    }
    return summary, nil
}
