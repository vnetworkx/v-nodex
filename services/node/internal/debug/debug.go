package debug

import (
	"encoding/json"
	"net/http"

	"vnodex/node/internal/peer"
	"vnodex/node/internal/store"
)

func StateDump(_ *store.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"note": "state dump not wired yet",
		})
	}
}

func PeersDump(reg *peer.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"peers": reg.List(),
		})
	}
}
