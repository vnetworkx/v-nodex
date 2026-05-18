package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NodeID string `yaml:"node_id"`

	// HTTP
	ListenAddr string `yaml:"listen_addr"`
	APIAddr    string `yaml:"api_addr"`

	// libp2p / QUIC
	P2PListenAddrs []string `yaml:"p2p_listen_addrs"`

	// bootstrap peers
	BootstrapPeers []string `yaml:"bootstrap_peers"`

	// storage
	DataDir string `yaml:"data_dir"`

	// consensus / snapshots
	RegionQuorum        int `yaml:"region_quorum"`
	SnapshotEveryEvents int `yaml:"snapshot_every_events"`

	// sync
	MaxPullBatch int `yaml:"max_pull_batch"`
	MaxPushBatch int `yaml:"max_push_batch"`

	// metrics
	EnableMetrics bool   `yaml:"enable_metrics"`
	MetricsAddr   string `yaml:"metrics_addr"`

	// debug
	EnableDebugRoutes bool `yaml:"enable_debug_routes"`

	// limits
	EventMaxSizeBytes int `yaml:"event_max_size_bytes"`

	// networking
	PeerConnectTimeoutSeconds int `yaml:"peer_connect_timeout_seconds"`
	SyncIntervalSeconds       int `yaml:"sync_interval_seconds"`

	MaxConcurrentSyncs int `yaml:"max_concurrent_syncs"`
}

func Load(path string) (Config, error) {
	cfg := Config{
		ListenAddr:                "0.0.0.0:18080",
		APIAddr:                   "0.0.0.0:18090",
		P2PListenAddrs:            []string{"/ip4/0.0.0.0/udp/19000/quic-v1"},
		DataDir:                   "./data",
		RegionQuorum:              2,
		SnapshotEveryEvents:       1024,
		MaxPullBatch:              512,
		MaxPushBatch:              512,
		EnableMetrics:             true,
		MetricsAddr:               "0.0.0.0:19100",
		EnableDebugRoutes:         false,
		EventMaxSizeBytes:         1048576,
		PeerConnectTimeoutSeconds: 10,
		SyncIntervalSeconds:       5,
		MaxConcurrentSyncs:        32,
	}

	if path == "" {
		return cfg, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}

	if cfg.NodeID == "" {
		return cfg, fmt.Errorf("node_id is required")
	}

	return cfg, nil
}
