package main

import (
	"flag"
	"log"
	"net/http"

	"vnodex/internal/config"
	"vnodex/internal/peer"
	"vnodex/internal/server"
	"vnodex/internal/store"
	"vnodex/internal/sync"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	st, err := store.New(cfg.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	reg := peer.NewRegistry(cfg.NodeID)
	mgr := sync.NewManager(st, reg)
	srv := server.New(st, reg, mgr, cfg.NodeID, true)

	log.Printf("v-nodex api listening on %s", cfg.APIAddr)
	if err := http.ListenAndServe(cfg.APIAddr, srv.Router()); err != nil {
		log.Fatalf("serve api: %v", err)
	}
}
