package node

import (
    "context"
    "encoding/json"
    "errors"
    "log/slog"
    "path/filepath"
    "sync"
    "time"

    vcrypto "github.com/vnetworkx/v-nodex/internal/crypto"
    "github.com/vnetworkx/v-nodex/internal/dag"
    "github.com/vnetworkx/v-nodex/internal/kernel"
    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/p2p"
    "github.com/vnetworkx/v-nodex/internal/replay"
    "github.com/vnetworkx/v-nodex/internal/snapshot"
    "github.com/vnetworkx/v-nodex/internal/store"
    nodesync "github.com/vnetworkx/v-nodex/internal/sync"
)

type RuntimeConfig struct {
    NodeID             string
    DataDir            string
    SnapshotEveryEvents int
    SyncInterval       time.Duration
    MaxPullBatch       int
    MaxConcurrentSyncs int
}

type Node struct {
    cfg    RuntimeConfig
    log    *slog.Logger
    kernel *kernel.Client
    store  *store.Store
    ident  *vcrypto.Identity

    replayEngine *replay.Engine
    snapshotMgr  *snapshot.Manager
    mesh         *p2p.Mesh
    syncManager  *nodesync.Manager

    stateMu sync.RWMutex
    state   model.StateView

    closing chan struct{}
    wg      sync.WaitGroup
}

func New(cfg RuntimeConfig, log *slog.Logger, kernelClient *kernel.Client, st *store.Store) *Node {
    rep := replay.New(st)
    mesh := p2p.NewMesh(p2p.NewHTTP(int(cfg.SyncInterval.Seconds())))
    identPath := filepath.Join(cfg.DataDir, "identity.json")
    ident, _ := vcrypto.LoadOrCreate(identPath)
    n := &Node{
        cfg:          cfg,
        log:          log,
        kernel:       kernelClient,
        store:        st,
        ident:        ident,
        replayEngine: rep,
        snapshotMgr:  snapshot.New(st, rep),
        mesh:         mesh,
        closing:      make(chan struct{}),
    }
    n.syncManager = nodesync.New(
        st,
        n,
        n,
        rep,
        mesh,
        model.Peer{
            PeerID:   n.KernelPeerID(),
            Healthy:  true,
            Score:    100,
            LastSeen: time.Now().UTC(),
        },
        nodesync.Limits{
            PullBatch:     cfg.MaxPullBatch,
            PushBatch:     cfg.MaxPullBatch,
            Interval:      cfg.SyncInterval,
            MaxConcurrent: cfg.MaxConcurrentSyncs,
        },
    )
    return n
}

func (n *Node) Start(ctx context.Context) error {
    if err := n.kernel.Health(ctx); err != nil {
        return err
    }
    if err := n.store.Init(ctx); err != nil {
        return err
    }
    report, err := n.replayEngine.Verify(ctx)
    if err != nil {
        report = model.ReplayReport{Verified: true, StartedAt: time.Now().UTC(), FinishedAt: time.Now().UTC()}
    }
    st := model.StateView{
        NodeID:      n.cfg.NodeID,
        StateRoot:   report.StateRoot,
        LatestClock:  report.LatestClock,
        EventCount:   int64(report.EventCount),
        Heads:        report.Heads,
        DerivedState: json.RawMessage(`{}`),
        UpdatedAt:    time.Now().UTC(),
    }
    if err := n.store.UpsertState(ctx, st); err != nil {
        return err
    }
    n.stateMu.Lock()
    n.state = st
    n.stateMu.Unlock()

    n.wg.Add(1)
    go n.syncLoop()
    return nil
}

func (n *Node) Stop() {
    close(n.closing)
    n.wg.Wait()
}

func (n *Node) syncLoop() {
    defer n.wg.Done()
    ticker := time.NewTicker(n.cfg.SyncInterval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            summary, err := n.syncManager.SyncOnce(ctx)
            cancel()
            if err == nil {
                n.log.Info("sync complete", "state_root", summary.StateRoot, "history_root", summary.HistoryRoot, "imported", summary.EventsImported)
            }
        case <-n.closing:
            return
        }
    }
}

func (n *Node) Submit(ctx context.Context, req model.EventRequest) (model.EventRecord, error) {
    if req.EventID == "" {
        return model.EventRecord{}, errors.New("event_id is required")
    }
    rec, err := n.kernel.SubmitEvent(ctx, req)
    if err != nil {
        return model.EventRecord{}, err
    }
    if err := n.ImportRecord(ctx, rec); err != nil {
        return model.EventRecord{}, err
    }
    return rec, nil
}

func (n *Node) ImportRecord(ctx context.Context, rec model.EventRecord) error {
    if rec.EventHash == "" {
        return errors.New("event_hash is required")
    }
    exists, err := n.store.ExistsEventHash(ctx, rec.EventHash)
    if err != nil {
        return err
    }
    if exists {
        return nil
    }
    if err := n.store.AppendEvent(ctx, rec); err != nil {
        return err
    }
    if err := n.ValidateDAG(ctx); err != nil {
        return err
    }
    report, _, err := n.replayEngine.Replay(ctx)
    if err != nil {
        return err
    }
    st := model.StateView{
        NodeID:      n.cfg.NodeID,
        StateRoot:   report.StateRoot,
        LatestClock:  report.LatestClock,
        EventCount:   int64(report.EventCount),
        Heads:        report.Heads,
        DerivedState: json.RawMessage(`{}`),
        UpdatedAt:    time.Now().UTC(),
    }
    if err := n.store.UpsertState(ctx, st); err != nil {
        return err
    }
    n.stateMu.Lock()
    n.state = st
    n.stateMu.Unlock()

    if n.cfg.SnapshotEveryEvents > 0 && report.EventCount > 0 && report.EventCount%n.cfg.SnapshotEveryEvents == 0 {
        _, _ = n.snapshotMgr.Create(ctx, st)
    }
    return nil
}

func (n *Node) State(ctx context.Context) (model.StateView, error) {
    n.stateMu.RLock()
    defer n.stateMu.RUnlock()
    return n.state, nil
}

func (n *Node) GetState(ctx context.Context) (model.StateView, error) {
    return n.State(ctx)
}


func (n *Node) UpsertState(ctx context.Context, st model.StateView) error {
    if err := n.store.UpsertState(ctx, st); err != nil {
        return err
    }
    n.stateMu.Lock()
    n.state = st
    n.stateMu.Unlock()
    return nil
}

func (n *Node) Replay(ctx context.Context) (model.ReplayReport, []model.EventRecord, error) {
    return n.replayEngine.Replay(ctx)
}

func (n *Node) CreateSnapshot(ctx context.Context) (model.Snapshot, error) {
    st, err := n.State(ctx)
    if err != nil {
        return model.Snapshot{}, err
    }
    return n.snapshotMgr.Create(ctx, st)
}

func (n *Node) ListEvents(ctx context.Context, limit, offset int) ([]model.EventRecord, error) {
    return n.store.ListEvents(ctx, limit, offset)
}

func (n *Node) ListPeers(ctx context.Context) ([]model.Peer, error) {
    return n.store.ListPeers(ctx)
}

func (n *Node) UpsertPeer(ctx context.Context, p model.Peer) error {
    if p.PeerID == "" {
        return errors.New("peer_id required")
    }
    if p.LastSeen.IsZero() {
        p.LastSeen = time.Now().UTC()
    }
    return n.store.UpsertPeer(ctx, p)
}

func (n *Node) ListSnapshots(ctx context.Context) ([]model.Snapshot, error) {
    return n.store.ListSnapshots(ctx)
}

func (n *Node) KernelPeerID() string {
    if n.ident == nil {
        return ""
    }
    return n.ident.PeerID()
}

func (n *Node) ValidateDAG(ctx context.Context) error {
    events, err := n.store.LoadEvents(ctx)
    if err != nil {
        return err
    }
    return dag.Validate(events)
}

func (n *Node) Metrics() map[string]any {
    st, _ := n.State(context.Background())
    events, _ := n.ListEvents(context.Background(), 1000000, 0)
    peers, _ := n.ListPeers(context.Background())
    return map[string]any{
        "event_count": len(events),
        "peer_count":  len(peers),
        "state_root":   st.StateRoot,
    }
}
