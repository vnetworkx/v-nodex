package main

import (
	"flag"
	"log"

	"vnodex/internal/config"
	"vnodex/internal/node"
)

func main() {
	configPath := flag.String(
		"config",
		"../../configs/node.example.yaml",
		"path to config file",
	)

	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	n, err := node.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting vnodex node: %s", cfg.NodeID)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
