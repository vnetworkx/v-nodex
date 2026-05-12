package main

import (
	"flag"
	"fmt"
	"log"

	"vnodex/internal/config"
	"vnodex/internal/node"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to YAML config")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	n, err := node.New(cfg)
	if err != nil {
		log.Fatalf("create node: %v", err)
	}

	fmt.Printf("v-nodex node %s listening on %s\n", cfg.NodeID, cfg.ListenAddr)
	if err := n.Run(); err != nil {
		log.Fatalf("run node: %v", err)
	}
}
