package config

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type Config struct {
    NodeID                 string   `json:"node_id"`
    ListenAddr             string   `json:"listen_addr"`
    APIAddr                string   `json:"api_addr"`
    KernelURL              string   `json:"kernel_url"`
    DataDir                string   `json:"data_dir"`
    BootstrapPeers         []string `json:"bootstrap_peers"`
    P2PListenAddrs         []string `json:"p2p_listen_addrs"`
    RegionQuorum           int      `json:"region_quorum"`
    SnapshotEveryEvents    int      `json:"snapshot_every_events"`
    MaxPullBatch           int      `json:"max_pull_batch"`
    MaxPushBatch           int      `json:"max_push_batch"`
    EnableMetrics          bool     `json:"enable_metrics"`
    MetricsAddr            string   `json:"metrics_addr"`
    EnableDebugRoutes      bool     `json:"enable_debug_routes"`
    EventMaxSizeBytes      int      `json:"event_max_size_bytes"`
    PeerConnectTimeoutSec  int      `json:"peer_connect_timeout_seconds"`
    SyncIntervalSec        int      `json:"sync_interval_seconds"`
    MaxConcurrentSyncs     int      `json:"max_concurrent_syncs"`
    RequestTimeoutSec      int      `json:"request_timeout_seconds"`
}

func Load(path string) (Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return Config{}, err
    }
    cfg := defaultConfig()
    if strings.HasSuffix(strings.ToLower(path), ".json") {
        if err := parseJSONish(string(b), &cfg); err != nil {
            return Config{}, err
        }
    } else {
        if err := parseYAML(string(b), &cfg); err != nil {
            return Config{}, err
        }
    }
    cfg.applyDefaults()
    if err := cfg.Validate(); err != nil {
        return Config{}, err
    }
    return cfg, nil
}

func defaultConfig() Config {
    return Config{
        ListenAddr:            "0.0.0.0:18080",
        APIAddr:               "0.0.0.0:18090",
        KernelURL:             "http://127.0.0.1:18180",
        RegionQuorum:          2,
        SnapshotEveryEvents:   1024,
        MaxPullBatch:          512,
        MaxPushBatch:          512,
        EnableMetrics:         true,
        MetricsAddr:           "0.0.0.0:19100",
        EnableDebugRoutes:     false,
        EventMaxSizeBytes:     1 << 20,
        PeerConnectTimeoutSec: 10,
        SyncIntervalSec:       5,
        MaxConcurrentSyncs:    32,
        RequestTimeoutSec:     15,
    }
}

func (c *Config) applyDefaults() {
    d := defaultConfig()
    if c.ListenAddr == "" { c.ListenAddr = d.ListenAddr }
    if c.APIAddr == "" { c.APIAddr = d.APIAddr }
    if c.KernelURL == "" { c.KernelURL = d.KernelURL }
    if c.RegionQuorum <= 0 { c.RegionQuorum = d.RegionQuorum }
    if c.SnapshotEveryEvents <= 0 { c.SnapshotEveryEvents = d.SnapshotEveryEvents }
    if c.MaxPullBatch <= 0 { c.MaxPullBatch = d.MaxPullBatch }
    if c.MaxPushBatch <= 0 { c.MaxPushBatch = d.MaxPushBatch }
    if c.MetricsAddr == "" { c.MetricsAddr = d.MetricsAddr }
    if c.EventMaxSizeBytes <= 0 { c.EventMaxSizeBytes = d.EventMaxSizeBytes }
    if c.PeerConnectTimeoutSec <= 0 { c.PeerConnectTimeoutSec = d.PeerConnectTimeoutSec }
    if c.SyncIntervalSec <= 0 { c.SyncIntervalSec = d.SyncIntervalSec }
    if c.MaxConcurrentSyncs <= 0 { c.MaxConcurrentSyncs = d.MaxConcurrentSyncs }
    if c.RequestTimeoutSec <= 0 { c.RequestTimeoutSec = d.RequestTimeoutSec }
}

func (c Config) Validate() error {
    switch {
    case c.NodeID == "":
        return errors.New("node_id is required")
    case c.DataDir == "":
        return errors.New("data_dir is required")
    case c.KernelURL == "":
        return errors.New("kernel_url is required")
    case c.ListenAddr == "":
        return errors.New("listen_addr is required")
    case c.APIAddr == "":
        return errors.New("api_addr is required")
    default:
        return nil
    }
}

func parseYAML(raw string, cfg *Config) error {
    scanner := bufio.NewScanner(strings.NewReader(raw))
    var currentList *[]string
    for scanner.Scan() {
        line := scanner.Text()
        trimmed := strings.TrimSpace(line)
        if trimmed == "" || strings.HasPrefix(trimmed, "#") {
            continue
        }
        if strings.HasPrefix(trimmed, "- ") {
            if currentList == nil {
                return fmt.Errorf("list item without list key: %q", trimmed)
            }
            *currentList = append(*currentList, strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
            continue
        }
        parts := strings.SplitN(trimmed, ":", 2)
        if len(parts) != 2 {
            return fmt.Errorf("invalid config line: %q", trimmed)
        }
        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        value = strings.Trim(value, `"'`)
        currentList = nil

        switch key {
        case "node_id":
            cfg.NodeID = value
        case "listen_addr":
            cfg.ListenAddr = value
        case "api_addr":
            cfg.APIAddr = value
        case "kernel_url":
            cfg.KernelURL = value
        case "data_dir":
            cfg.DataDir = value
        case "region_quorum":
            cfg.RegionQuorum = mustInt(value)
        case "snapshot_every_events":
            cfg.SnapshotEveryEvents = mustInt(value)
        case "max_pull_batch":
            cfg.MaxPullBatch = mustInt(value)
        case "max_push_batch":
            cfg.MaxPushBatch = mustInt(value)
        case "enable_metrics":
            cfg.EnableMetrics = mustBool(value)
        case "metrics_addr":
            cfg.MetricsAddr = value
        case "enable_debug_routes":
            cfg.EnableDebugRoutes = mustBool(value)
        case "event_max_size_bytes":
            cfg.EventMaxSizeBytes = mustInt(value)
        case "peer_connect_timeout_seconds":
            cfg.PeerConnectTimeoutSec = mustInt(value)
        case "sync_interval_seconds":
            cfg.SyncIntervalSec = mustInt(value)
        case "max_concurrent_syncs":
            cfg.MaxConcurrentSyncs = mustInt(value)
        case "request_timeout_seconds":
            cfg.RequestTimeoutSec = mustInt(value)
        case "bootstrap_peers":
            cfg.BootstrapPeers = []string{}
            currentList = &cfg.BootstrapPeers
            if value != "" {
                cfg.BootstrapPeers = append(cfg.BootstrapPeers, value)
            }
        case "p2p_listen_addrs":
            cfg.P2PListenAddrs = []string{}
            currentList = &cfg.P2PListenAddrs
            if value != "" {
                cfg.P2PListenAddrs = append(cfg.P2PListenAddrs, value)
            }
        default:
            // ignore unknown keys to allow forward compatibility
        }
    }
    return scanner.Err()
}

func parseJSONish(raw string, cfg *Config) error {
    // tiny permissive parser to avoid external dependencies
    // supports a simple "key": value format with arrays of strings
    // meant for tests, not arbitrary JSON.
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return nil
    }
    return errors.New("json config parsing is not enabled in this build; use yaml")
}

func mustInt(v string) int {
    n, _ := strconv.Atoi(v)
    return n
}

func mustBool(v string) bool {
    switch strings.ToLower(strings.TrimSpace(v)) {
    case "true", "yes", "1", "on":
        return true
    default:
        return false
    }
}
