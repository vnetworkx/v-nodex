package main

import (
    "context"
    "flag"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/vnetworkx/v-nodex/internal/config"
    "github.com/vnetworkx/v-nodex/internal/kernel"
    "github.com/vnetworkx/v-nodex/internal/node"
    "github.com/vnetworkx/v-nodex/internal/server"
    "github.com/vnetworkx/v-nodex/internal/store"
)

func main() {
    var configPath string
    flag.StringVar(&configPath, "config", "configs/node.yaml", "path to node config")
    flag.Parse()

    log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

    cfg, err := config.Load(configPath)
    if err != nil {
        log.Error("load config failed", "error", err)
        os.Exit(1)
    }

    st, err := store.Open(cfg.DataDir)
    if err != nil {
        log.Error("open store failed", "error", err)
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
    n := node.New(rt, log, kernelClient, st)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := n.Start(ctx); err != nil {
        log.Error("node start failed", "error", err)
        os.Exit(1)
    }
    defer n.Stop()

    srv := server.New(n, int64(cfg.EventMaxSizeBytes), cfg.EnableDebugRoutes)
    httpSrv := &http.Server{
        Addr:              cfg.APIAddr,
        Handler:           srv.Handler(),
        ReadHeaderTimeout: 5 * time.Second,
    }

    go func() {
        log.Info("api listening", "addr", cfg.APIAddr)
        if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Error("api server failed", "error", err)
            cancel()
        }
    }()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    select {
    case <-sigCh:
    case <-ctx.Done():
    }

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()
    _ = httpSrv.Shutdown(shutdownCtx)
}
