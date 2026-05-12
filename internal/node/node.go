package node

import (
	"fmt"
	"net/http"

	"vnodex/internal/config"
	"vnodex/internal/model"
	"vnodex/internal/peer"
	"vnodex/internal/server"
	"vnodex/internal/store"
	"vnodex/internal/sync"
)

type Node struct {
	cfg    config.Config
	store  *store.Store
	peers  *peer.Registry
	sync   *sync.Manager
	server *server.Server
}

func New(cfg config.Config) (*Node, error) {
	st, err := store.New(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	reg := peer.NewRegistry(cfg.NodeID)
	mgr := sync.NewManager(st, reg)
	srv := server.New(st, reg, mgr, cfg.NodeID, false)
	for _, addr := range cfg.BootstrapPeers {
		reg.Add(model.PeerInfo{PeerID: addr, Address: addr, RegionScope: ""})
	}
	return &Node{cfg: cfg, store: st, peers: reg, sync: mgr, server: srv}, nil
}

func (n *Node) Router() http.Handler { return n.server.Router() }

func (n *Node) Run() error {
	addr := n.cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}
	return http.ListenAndServe(addr, n.Router())
}

func (n *Node) Store() *store.Store   { return n.store }
func (n *Node) Peers() *peer.Registry { return n.peers }
func (n *Node) Sync() *sync.Manager   { return n.sync }
func (n *Node) Config() config.Config { return n.cfg }

func (n *Node) String() string { return fmt.Sprintf("node<%s>", n.cfg.NodeID) }
