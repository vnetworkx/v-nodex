package replay

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "sort"
    "strings"
    "time"

    "github.com/vnetworkx/v-nodex/internal/dag"
    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/store"
)

type Engine struct {
    store *store.Store
}

func New(store *store.Store) *Engine {
    return &Engine{store: store}
}

func (e *Engine) Replay(ctx context.Context) (model.ReplayReport, []model.EventRecord, error) {
    started := time.Now().UTC()
    events, err := e.store.LoadEvents(ctx)
    if err != nil {
        return model.ReplayReport{}, nil, err
    }
    if err := dag.Validate(events); err != nil {
        return model.ReplayReport{}, nil, err
    }
    ordered, err := dag.TopologicalOrder(events)
    if err != nil {
        return model.ReplayReport{}, nil, err
    }

    report := model.ReplayReport{
        EventCount:    len(ordered),
        StartedAt:     started,
        FinishedAt:    time.Now().UTC(),
        Verified:      true,
        OrderedHashes: make([]string, 0, len(ordered)),
        Heads:         dag.Heads(ordered),
    }

    var stateRoot string
    var latestClock int64
    h := sha256.New()
    for _, ev := range ordered {
        report.OrderedHashes = append(report.OrderedHashes, ev.EventHash)
        if ev.LogicalClock > latestClock {
            latestClock = ev.LogicalClock
        }
        if ev.StateRoot != "" {
            stateRoot = ev.StateRoot
        }
        writeHashPiece(h, ev.EventHash)
        writeHashPiece(h, "|")
        writeHashPiece(h, ev.PayloadHash)
        writeHashPiece(h, "|")
        writeHashPiece(h, fmt.Sprintf("%d", ev.LogicalClock))
        writeHashPiece(h, "|")
        writeHashPiece(h, strings.Join(ev.ParentHashes, ","))
        writeHashPiece(h, "\n")
    }
    report.LatestClock = latestClock
    report.StateRoot = stateRoot
    report.HistoryRoot = hex.EncodeToString(h.Sum(nil))

    ch := sha256.Sum256([]byte(strings.Join(report.OrderedHashes, "|") + "::" + report.HistoryRoot))
    report.CanonicalHash = hex.EncodeToString(ch[:])
    return report, ordered, nil
}

func (e *Engine) Verify(ctx context.Context) (model.ReplayReport, error) {
    report, _, err := e.Replay(ctx)
    return report, err
}

func writeHashPiece(h interface{ Write([]byte) (int, error) }, s string) {
    _, _ = h.Write([]byte(s))
}

func SortByHash(records []model.EventRecord) {
    sort.SliceStable(records, func(i, j int) bool {
        if records[i].LogicalClock != records[j].LogicalClock {
            return records[i].LogicalClock < records[j].LogicalClock
        }
        if records[i].EventHash != records[j].EventHash {
            return records[i].EventHash < records[j].EventHash
        }
        return records[i].EventID < records[j].EventID
    })
}
