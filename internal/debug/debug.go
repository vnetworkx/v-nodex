package debug

import (
    "context"
    "encoding/json"
    "net/http"
    "runtime"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type StateProvider interface {
    State(ctx context.Context) (model.StateView, error)
}

func Handler(provider StateProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        st, err := provider.State(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        out := map[string]any{
            "state": st,
            "runtime": map[string]any{
                "go_version": runtime.Version(),
                "time": time.Now().UTC(),
            },
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(out)
    }
}
