package node

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"vnodex/internal/config"
	"vnodex/internal/model"
	"vnodex/internal/p2p"
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
	http   *http.Server
	p2p    *p2p.Node
}

func New(cfg config.Config) (*Node, error) {
	st, err := store.New(cfg.DataDir)
	if err != nil {
		return nil, err
	}

	reg := peer.NewRegistry(cfg.NodeID)
	mgr := sync.NewManager(st, reg)

	p2pNode, err := p2p.New(context.Background(), p2p.Config{
		ListenAddrs:    []string{"/ip4/0.0.0.0/udp/0/quic-v1"},
		BootstrapPeers: cfg.BootstrapPeers,
		DataDir:        cfg.DataDir,
		Topic:          "vnodex-events",
	})
	if err != nil {
		return nil, err
	}

	for _, addr := range cfg.BootstrapPeers {
		reg.Add(model.PeerInfo{
			PeerID:      addr,
			Address:     addr,
			RegionScope: "",
		})
		_ = p2pNode.AddPeerAddr(addr)
	}

	srv := server.New(st, reg, mgr, cfg.NodeID, false)

	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Node{
		cfg:    cfg,
		store:  st,
		peers:  reg,
		sync:   mgr,
		server: srv,
		http:   httpSrv,
		p2p:    p2pNode,
	}, nil
}

func (n *Node) Router() http.Handler { return n.server.Router() }

func (n *Node) Run() error {
	if n.http == nil {
		return fmt.Errorf("node: http server not initialized")
	}
	err := n.http.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (n *Node) Shutdown(ctx context.Context) error {
	if n.p2p != nil {
		_ = n.p2p.Close()
	}
	if n.http != nil {
		return n.http.Shutdown(ctx)
	}
	return nil
}

func (n *Node) Store() *store.Store   { return n.store }
func (n *Node) Peers() *peer.Registry { return n.peers }
func (n *Node) Sync() *sync.Manager   { return n.sync }
func (n *Node) Config() config.Config { return n.cfg }
func (n *Node) P2P() *p2p.Node        { return n.p2p }

func (n *Node) String() string { return fmt.Sprintf("node<%s>", n.cfg.NodeID) }
