package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"vnodex/api/internal/store"
)

type Handler struct {
	store *store.Storage
}

func New(store *store.Storage) *Handler { return &Handler{store: store} }

type envelope struct {
	RequestID  string      `json:"request_id"`
	ServerTime time.Time   `json:"server_time"`
	NodeID     string      `json:"node_id"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, envelope{RequestID: "healthz", ServerTime: time.Now().UTC(), NodeID: "local", Data: map[string]string{"status": "ok"}})
}

func (h *Handler) GetWalletState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{RequestID: "state", ServerTime: time.Now().UTC(), NodeID: "local", Error: "state lookup not wired yet"})
}

func (h *Handler) GetWalletRecords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{RequestID: "records", ServerTime: time.Now().UTC(), NodeID: "local", Error: "record lookup not wired yet"})
}

func (h *Handler) GetHead(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{RequestID: "head", ServerTime: time.Now().UTC(), NodeID: "local", Error: "head lookup not wired yet"})
}

func (h *Handler) PostOperation(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{RequestID: "operation", ServerTime: time.Now().UTC(), NodeID: "local", Error: "mutation path not wired yet"})
}

func (h *Handler) PushRecords(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, envelope{RequestID: "sync", ServerTime: time.Now().UTC(), NodeID: "local", Error: "sync path not wired yet"})
}
