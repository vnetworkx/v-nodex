package kernel

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

type Client struct {
    baseURL string
    http    *http.Client
}

func New(baseURL string, timeoutSeconds int) *Client {
    if timeoutSeconds <= 0 {
        timeoutSeconds = 15
    }
    return &Client{
        baseURL: stringsTrimRightSlash(baseURL),
        http: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
    }
}

func (c *Client) Health(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/health", nil)
    if err != nil {
        return err
    }
    resp, err := c.http.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("kernel health failed: %s %s", resp.Status, string(b))
    }
    return nil
}

func (c *Client) SubmitEvent(ctx context.Context, reqBody model.EventRequest) (model.EventRecord, error) {
    b, err := json.Marshal(reqBody)
    if err != nil {
        return model.EventRecord{}, err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/events/submit", bytes.NewReader(b))
    if err != nil {
        return model.EventRecord{}, err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.http.Do(req)
    if err != nil {
        return model.EventRecord{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return model.EventRecord{}, fmt.Errorf("kernel submit failed: %s %s", resp.Status, string(body))
    }
    var out struct {
        Accepted bool              `json:"accepted"`
        Record   model.EventRecord `json:"record"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return model.EventRecord{}, err
    }
    if !out.Accepted {
        return model.EventRecord{}, fmt.Errorf("kernel rejected event: %s", out.Record.Reason)
    }
    return out.Record, nil
}

func (c *Client) Replay(ctx context.Context) (model.ReplayReport, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/replay", nil)
    if err != nil {
        return model.ReplayReport{}, err
    }
    resp, err := c.http.Do(req)
    if err != nil {
        return model.ReplayReport{}, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return model.ReplayReport{}, fmt.Errorf("kernel replay failed: %s %s", resp.Status, string(body))
    }
    var report model.ReplayReport
    if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
        return model.ReplayReport{}, err
    }
    return report, nil
}

func (c *Client) ComputeStateRoot(ctx context.Context) (string, error) {
    report, err := c.Replay(ctx)
    if err != nil {
        return "", err
    }
    return report.StateRoot, nil
}

func (c *Client) VerifyRecord(ctx context.Context, rec model.EventRecord) (bool, error) {
    b, err := json.Marshal(rec)
    if err != nil {
        return false, err
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/records/verify", bytes.NewReader(b))
    if err != nil {
        return false, err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.http.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return false, fmt.Errorf("kernel verify failed: %s %s", resp.Status, string(body))
    }
    var out struct{ OK bool `json:"ok"` }
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        return false, err
    }
    return out.OK, nil
}

func stringsTrimRightSlash(s string) string {
    for len(s) > 0 && s[len(s)-1] == '/' {
        s = s[:len(s)-1]
    }
    return s
}
