package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	NodeID              string
	ListenAddr          string
	APIAddr             string
	DataDir             string
	BootstrapPeers      []string
	RegionQuorum        int
	SnapshotEveryEvents int
	MaxPullBatch        int
	MaxPushBatch        int
}

func Default() Config {
	return Config{
		NodeID:              "node-0001",
		ListenAddr:          ":8080",
		APIAddr:             ":8090",
		DataDir:             "./data",
		RegionQuorum:        2,
		SnapshotEveryEvents: 128,
		MaxPullBatch:        512,
		MaxPushBatch:        512,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	applyEnv(&cfg)
	if path != "" {
		if err := applyYAML(path, &cfg); err != nil {
			return cfg, err
		}
	}
	if cfg.NodeID == "" {
		return cfg, errors.New("node id is required")
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.APIAddr == "" {
		cfg.APIAddr = ":8090"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}
	return cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("VNODEX_NODE_ID"); v != "" {
		cfg.NodeID = v
	}
	if v := os.Getenv("VNODEX_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("VNODEX_API_ADDR"); v != "" {
		cfg.APIAddr = v
	}
	if v := os.Getenv("VNODEX_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("VNODEX_REGION_QUORUM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RegionQuorum = n
		}
	}
	if v := os.Getenv("VNODEX_SNAPSHOT_EVERY_EVENTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SnapshotEveryEvents = n
		}
	}
	if v := os.Getenv("VNODEX_MAX_PULL_BATCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxPullBatch = n
		}
	}
	if v := os.Getenv("VNODEX_MAX_PUSH_BATCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxPushBatch = n
		}
	}
}

func applyYAML(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var currentList string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") {
			currentList = strings.TrimSuffix(line, ":")
			continue
		}
		if strings.HasPrefix(line, "-") && currentList == "bootstrap_peers" {
			cfg.BootstrapPeers = append(cfg.BootstrapPeers, strings.TrimSpace(strings.TrimPrefix(line, "-")))
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		switch key {
		case "node_id":
			cfg.NodeID = value
		case "listen_addr":
			cfg.ListenAddr = value
		case "api_addr":
			cfg.APIAddr = value
		case "data_dir":
			cfg.DataDir = value
		case "region_quorum":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.RegionQuorum = n
			}
		case "snapshot_every_events":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.SnapshotEveryEvents = n
			}
		case "max_pull_batch":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.MaxPullBatch = n
			}
		case "max_push_batch":
			if n, err := strconv.Atoi(value); err == nil {
				cfg.MaxPushBatch = n
			}
		}
	}
	return s.Err()
}
