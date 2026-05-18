package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"vnodex/internal/config"
	"vnodex/internal/model"
	"vnodex/internal/store"
)

func main() {
	var (
		dataDir    string
		configPath string
	)

	flag.StringVar(&dataDir, "data", "./data", "node data directory")

	flag.StringVar(
		&configPath,
		"config",
		"",
		"path to node config file",
	)

	flag.Parse()

	// optional config load
	if configPath != "" {
		cfg, err := config.Load(configPath)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}

		// config data_dir overrides CLI default
		if cfg.DataDir != "" {
			dataDir = cfg.DataDir
		}
	}

	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	s, err := store.New(dataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	switch args[0] {

	case "status":
		printJSON(s.State())

	case "events":
		limit := 50

		if len(args) > 1 {
			if n, err := parseInt(args[1]); err == nil {
				limit = n
			}
		}

		printJSON(s.ListEvents(limit))

	case "state":
		if len(args) < 2 {
			log.Fatal("state requires entity id")
		}

		entity, ok := s.Entity(args[1])
		if !ok {
			log.Fatalf("entity %s not found", args[1])
		}

		printJSON(entity)

	case "snapshot":
		scope := "global"

		if len(args) > 1 {
			scope = args[1]
		}

		snap, err := s.Snapshot(scope)
		if err != nil {
			log.Fatalf("snapshot: %v", err)
		}

		printJSON(snap)

	case "submit":
		ev := sampleEvent(args[1:])

		appended, err := s.Append(ev)
		if err != nil {
			log.Fatalf("submit: %v", err)
		}

		printJSON(appended)

	case "keygen":
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatalf("keygen: %v", err)
		}

		fmt.Println("public_key_hex:", hex.EncodeToString(pub))
		fmt.Println("private_key_hex:", hex.EncodeToString(priv))

	default:
		usage()
	}
}

func usage() {
	fmt.Println("vnodex-cli commands:")
	fmt.Println("  status")
	fmt.Println("  events [limit]")
	fmt.Println("  state <entity>")
	fmt.Println("  snapshot [scope]")
	fmt.Println("  submit [entity]")
	fmt.Println("  keygen")
}

func sampleEvent(args []string) model.Event {
	now := time.Now().UTC()

	entity := "entity-0001"

	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		entity = args[0]
	}

	core := model.EventCore{
		EventID:                entity + "-" + now.Format("20060102150405.000000000"),
		SpaceID:                "space-main",
		RegionID:               "region-0001",
		EntityID:               entity,
		VectorType:             model.VectorTypeFree,
		Operation:              model.OpCreate,
		InputVector:            model.Vector{1, 2, 3},
		OutputVector:           model.Vector{1, 2, 3},
		MagnitudeBefore:        0,
		MagnitudeAfter:         6,
		DirectionBefore:        model.Vector{0, 0, 0},
		DirectionAfter:         model.Vector{0.1666666667, 0.3333333333, 0.5},
		SpaceCoordinatesBefore: model.Vector{0, 0, 0},
		SpaceCoordinatesAfter:  model.Vector{1, 1, 1},
		TimeCreated:            now,
		LogicalOrder:           uint64(now.UnixNano()),
		SignerID:               entity,
		SignerPublicKey:        "",
		ValidationProof:        "local-node",
	}

	hash, _, _ := model.CanonicalEventHash(core)

	return model.Event{
		EventCore: core,
		EventHash: hash,
		Signature: "",
		Certified: true,
		Accepted:  true,
	}
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		log.Fatal(err)
	}
}

func parseInt(v string) (int, error) {
	var n int

	_, err := fmt.Sscanf(v, "%d", &n)

	return n, err
}
