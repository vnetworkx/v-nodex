package node

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"vnodex/node/internal/config"
	"vnodex/node/internal/debug"
	"vnodex/node/internal/peer"
	"vnodex/node/internal/store"
	"vnodex/node/internal/sync"
)

type Node struct {
	cfg   config.Config
	store *store.Storage
	peers *peer.Registry
	sync  *sync.Manager
}

func New(cfg config.Config) (*Node, error) {
	st, err := store.New(cfg.DBDSN, cfg.RedisAddr)
	if err != nil {
		return nil, err
	}
	reg := peer.NewRegistry(cfg.NodeID)
	mgr := sync.NewManager(st, reg)
	return &Node{cfg: cfg, store: st, peers: reg, sync: mgr}, nil
}

func (n *Node) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/debug/state", debug.StateDump(n.store))
	r.Get("/debug/peers", debug.PeersDump(n.peers))
	r.Post("/sync/pull", n.sync.PullHandler)
	r.Post("/sync/push", n.sync.PushHandler)
	r.Post("/sync/ack", n.sync.AckHandler)
	return r
}
