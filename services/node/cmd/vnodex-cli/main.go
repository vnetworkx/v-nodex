package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log/slog"
    "os"
    "strings"
    "time"

    "github.com/vnetworkx/v-nodex/internal/config"
    "github.com/vnetworkx/v-nodex/internal/kernel"
    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/node"
    "github.com/vnetworkx/v-nodex/internal/store"
)

func main() {
    var configPath string
    flag.StringVar(&configPath, "config", "configs/node.yaml", "path to node config")
    flag.Parse()

    if flag.NArg() == 0 {
        usage()
        os.Exit(1)
    }

    cfg, err := config.Load(configPath)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    st, err := store.Open(cfg.DataDir)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    kernelClient := kernel.New(cfg.KernelURL, cfg.RequestTimeoutSec)
    rt := node.RuntimeConfig{
        NodeID:              cfg.NodeID,
        DataDir:             cfg.DataDir,
        SnapshotEveryEvents:  cfg.SnapshotEveryEvents,
        SyncInterval:        time.Duration(cfg.SyncIntervalSec) * time.Second,
        MaxPullBatch:        cfg.MaxPullBatch,
        MaxConcurrentSyncs:  cfg.MaxConcurrentSyncs,
    }
    n := node.New(rt, slog.New(slog.NewTextHandler(io.Discard, nil)), kernelClient, st)
    ctx := context.Background()

    cmd := strings.ToLower(flag.Arg(0))
    switch cmd {
    case "status":
        stv, _ := n.State(ctx)
        out, _ := json.MarshalIndent(stv, "", "  ")
        fmt.Println(string(out))
    case "events":
        rows, _ := n.ListEvents(ctx, 100, 0)
        out, _ := json.MarshalIndent(rows, "", "  ")
        fmt.Println(string(out))
    case "peers":
        rows, _ := n.ListPeers(ctx)
        out, _ := json.MarshalIndent(rows, "", "  ")
        fmt.Println(string(out))
    case "snapshots":
        rows, _ := n.ListSnapshots(ctx)
        out, _ := json.MarshalIndent(rows, "", "  ")
        fmt.Println(string(out))
    case "replay":
        rep, _, err := n.Replay(ctx)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        out, _ := json.MarshalIndent(rep, "", "  ")
        fmt.Println(string(out))
    case "submit":
        if flag.NArg() < 2 {
            fmt.Fprintln(os.Stderr, "submit expects a JSON event request string as the second argument")
            os.Exit(1)
        }
        var req model.EventRequest
        if err := json.Unmarshal([]byte(flag.Arg(1)), &req); err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        rec, err := n.Submit(ctx, req)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        out, _ := json.MarshalIndent(rec, "", "  ")
        fmt.Println(string(out))
    default:
        usage()
        os.Exit(1)
    }
}

func usage() {
    fmt.Println("vnodex-cli <status|events|peers|snapshots|replay|submit>")
}
