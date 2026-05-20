package p2p

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Transport interface {
    FetchEvents(ctx context.Context, baseURL string, limit, offset int) ([]model.EventRecord, error)
    PushEvent(ctx context.Context, baseURL string, rec model.EventRecord) error
    FetchState(ctx context.Context, baseURL string) (model.StateView, error)
    FetchSnapshots(ctx context.Context, baseURL string) ([]model.Snapshot, error)
    AnnouncePeer(ctx context.Context, baseURL string, peer model.Peer) error
}

type HTTPTransport struct {
    client *http.Client
}

func NewHTTP(timeoutSeconds int) *HTTPTransport {
    if timeoutSeconds <= 0 {
        timeoutSeconds = 10
    }
    return &HTTPTransport{
        client: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
    }
}

func (t *HTTPTransport) FetchEvents(ctx context.Context, baseURL string, limit, offset int) ([]model.EventRecord, error) {
    url := trim(baseURL) + fmt.Sprintf("/v1/events?limit=%d&offset=%d", limit, offset)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := t.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("fetch events: %s %s", resp.Status, string(b))
    }
    var out []model.EventRecord
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return nil, err
    }
    return out, nil
}

func (t *HTTPTransport) PushEvent(ctx context.Context, baseURL string, rec model.EventRecord) error {
    b, err := json.Marshal(rec)
    if err != nil {
        return err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, trim(baseURL)+"/v1/events/ingest", bytes.NewReader(b))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := t.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("push event: %s %s", resp.Status, string(b))
    }
    return nil
}

func (t *HTTPTransport) FetchState(ctx context.Context, baseURL string) (model.StateView, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, trim(baseURL)+"/v1/state", nil)
    if err != nil {
        return model.StateView{}, err
    }
    resp, err := t.client.Do(req)
    if err != nil {
        return model.StateView{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return model.StateView{}, fmt.Errorf("fetch state: %s %s", resp.Status, string(b))
    }
    var st model.StateView
    if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
        return model.StateView{}, err
    }
    return st, nil
}

func (t *HTTPTransport) FetchSnapshots(ctx context.Context, baseURL string) ([]model.Snapshot, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, trim(baseURL)+"/v1/snapshots", nil)
    if err != nil {
        return nil, err
    }
    resp, err := t.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("fetch snapshots: %s %s", resp.Status, string(b))
    }
    var snaps []model.Snapshot
    if err := json.NewDecoder(resp.Body).Decode(&snaps); err != nil {
        return nil, err
    }
    return snaps, nil
}

func (t *HTTPTransport) AnnouncePeer(ctx context.Context, baseURL string, peer model.Peer) error {
    b, err := json.Marshal(peer)
    if err != nil {
        return err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, trim(baseURL)+"/v1/peers", bytes.NewReader(b))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := t.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("announce peer: %s %s", resp.Status, string(b))
    }
    return nil
}

func trim(s string) string {
    for len(s) > 0 && s[len(s)-1] == '/' {
        s = s[:len(s)-1]
    }
    return s
}
