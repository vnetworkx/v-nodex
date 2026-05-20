package server

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "time"

    "github.com/vnetworkx/v-nodex/internal/debug"
    "github.com/vnetworkx/v-nodex/internal/model"
    "github.com/vnetworkx/v-nodex/internal/node"
)

type Server struct {
    node   *node.Node
    mux    *http.ServeMux
    maxReq int64
    debug  bool
}

func New(n *node.Node, maxRequestSize int64, enableDebug bool) *Server {
    s := &Server{
        node:   n,
        mux:    http.NewServeMux(),
        maxReq: maxRequestSize,
        debug:  enableDebug,
    }
    s.routes()
    return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
    s.mux.HandleFunc("/healthz", s.healthz)
    s.mux.HandleFunc("/readyz", s.readyz)
    s.mux.HandleFunc("/v1/state", s.state)
    s.mux.HandleFunc("/v1/replay", s.replay)
    s.mux.HandleFunc("/v1/events", s.events)
    s.mux.HandleFunc("/v1/events/submit", s.submit)
    s.mux.HandleFunc("/v1/events/ingest", s.ingest)
    s.mux.HandleFunc("/v1/peers", s.peers)
    s.mux.HandleFunc("/v1/snapshots", s.snapshots)
    s.mux.HandleFunc("/v1/node", s.nodeInfo)
    s.mux.HandleFunc("/metrics", s.metrics)
    if s.debug {
        s.mux.HandleFunc("/debug/state", debug.Handler(s.node))
    }
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
    st, err := s.node.State(r.Context())
    if err != nil {
        writeErr(w, http.StatusServiceUnavailable, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state_root": st.StateRoot})
}

func (s *Server) state(w http.ResponseWriter, r *http.Request) {
    st, err := s.node.State(r.Context())
    if err != nil {
        writeErr(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, http.StatusOK, st)
}

func (s *Server) replay(w http.ResponseWriter, r *http.Request) {
    report, _, err := s.node.Replay(r.Context())
    if err != nil {
        writeErr(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, http.StatusOK, report)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
    limit := parseInt(r.URL.Query().Get("limit"), 100)
    offset := parseInt(r.URL.Query().Get("offset"), 0)
    rows, err := s.node.ListEvents(r.Context(), limit, offset)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, http.StatusOK, rows)
}

func (s *Server) submit(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
        return
    }
    var req model.EventRequest
    if err := decodeJSON(w, r, &req, s.maxReq); err != nil {
        writeErr(w, http.StatusBadRequest, err)
        return
    }
    rec, err := s.node.Submit(r.Context(), req)
    if err != nil {
        writeErr(w, http.StatusBadRequest, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"accepted": true, "record": rec})
}

func (s *Server) ingest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
        return
    }
    var rec model.EventRecord
    if err := decodeJSON(w, r, &rec, s.maxReq); err != nil {
        writeErr(w, http.StatusBadRequest, err)
        return
    }
    if err := s.node.ImportRecord(r.Context(), rec); err != nil {
        writeErr(w, http.StatusBadRequest, err)
        return
    }
    writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (s *Server) peers(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        rows, err := s.node.ListPeers(r.Context())
        if err != nil {
            writeErr(w, http.StatusInternalServerError, err)
            return
        }
        writeJSON(w, http.StatusOK, rows)
    case http.MethodPost:
        var p model.Peer
        if err := decodeJSON(w, r, &p, s.maxReq); err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        if err := s.node.UpsertPeer(r.Context(), p); err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
    default:
        writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
    }
}

func (s *Server) snapshots(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        snaps, err := s.node.ListSnapshots(r.Context())
        if err != nil {
            writeErr(w, http.StatusInternalServerError, err)
            return
        }
        writeJSON(w, http.StatusOK, snaps)
    case http.MethodPost:
        snap, err := s.node.CreateSnapshot(r.Context())
        if err != nil {
            writeErr(w, http.StatusBadRequest, err)
            return
        }
        writeJSON(w, http.StatusCreated, snap)
    default:
        writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
    }
}

func (s *Server) nodeInfo(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]any{
        "peer_id": s.node.KernelPeerID(),
        "time": time.Now().UTC(),
    })
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
    st, err := s.node.State(r.Context())
    if err != nil {
        writeErr(w, http.StatusInternalServerError, err)
        return
    }
    events, _ := s.node.ListEvents(r.Context(), 100000, 0)
    peers, _ := s.node.ListPeers(r.Context())
    lines := []string{
        fmt.Sprintf("# HELP vnodex_event_count Total events"),
        fmt.Sprintf("# TYPE vnodex_event_count gauge"),
        fmt.Sprintf("vnodex_event_count %d", len(events)),
        fmt.Sprintf("# HELP vnodex_peer_count Total peers"),
        fmt.Sprintf("# TYPE vnodex_peer_count gauge"),
        fmt.Sprintf("vnodex_peer_count %d", len(peers)),
        fmt.Sprintf("# HELP vnodex_state_root_info Current state root"),
        fmt.Sprintf("# TYPE vnodex_state_root_info gauge"),
        fmt.Sprintf("vnodex_state_root_info{state_root=%q} 1", st.StateRoot),
    }
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    _, _ = io.WriteString(w, stringsJoin(lines, "\n")+"\n")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, max int64) error {
    r.Body = http.MaxBytesReader(w, r.Body, max)
    defer r.Body.Close()
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
    writeJSON(w, code, map[string]any{"ok": false, "error": err.Error()})
}

func parseInt(s string, fallback int) int {
    if s == "" {
        return fallback
    }
    n, err := strconv.Atoi(s)
    if err != nil || n < 0 {
        return fallback
    }
    return n
}

func stringsJoin(parts []string, sep string) string {
    if len(parts) == 0 {
        return ""
    }
    out := parts[0]
    for _, p := range parts[1:] {
        out += sep + p
    }
    return out
}
