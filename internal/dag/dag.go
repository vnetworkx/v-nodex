package dag

import (
    "errors"
    "fmt"
    "sort"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Graph struct {
    ByHash map[string]model.EventRecord
    Parents map[string][]string
    Children map[string][]string
    Heads []string
}

func Build(records []model.EventRecord) *Graph {
    g := &Graph{
        ByHash:   make(map[string]model.EventRecord, len(records)),
        Parents:  make(map[string][]string, len(records)),
        Children: make(map[string][]string, len(records)),
    }
    for _, r := range records {
        g.ByHash[r.EventHash] = r
        g.Parents[r.EventHash] = append([]string(nil), r.ParentHashes...)
        if _, ok := g.Children[r.EventHash]; !ok {
            g.Children[r.EventHash] = nil
        }
        for _, p := range r.ParentHashes {
            g.Children[p] = append(g.Children[p], r.EventHash)
        }
    }
    for h := range g.ByHash {
        if len(g.Children[h]) == 0 {
            g.Heads = append(g.Heads, h)
        }
    }
    sort.Strings(g.Heads)
    return g
}

func Validate(records []model.EventRecord) error {
    g := Build(records)
    for _, r := range records {
        for _, p := range r.ParentHashes {
            if p == r.EventHash {
                return fmt.Errorf("self-parent detected: %s", r.EventHash)
            }
            if _, ok := g.ByHash[p]; !ok {
                return fmt.Errorf("missing parent %s for %s", p, r.EventHash)
            }
        }
    }

    visited := map[string]int{}
    var visit func(string) error
    visit = func(h string) error {
        switch visited[h] {
        case 1:
            return fmt.Errorf("cycle detected at %s", h)
        case 2:
            return nil
        }
        visited[h] = 1
        for _, c := range g.Children[h] {
            if err := visit(c); err != nil {
                return err
            }
        }
        visited[h] = 2
        return nil
    }
    for h := range g.ByHash {
        if visited[h] == 0 {
            if err := visit(h); err != nil {
                return err
            }
        }
    }
    return nil
}

func TopologicalOrder(records []model.EventRecord) ([]model.EventRecord, error) {
    if err := Validate(records); err != nil {
        return nil, err
    }
    out := append([]model.EventRecord(nil), records...)
    sort.SliceStable(out, func(i, j int) bool {
        if out[i].LogicalClock != out[j].LogicalClock {
            return out[i].LogicalClock < out[j].LogicalClock
        }
        if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
            return out[i].CreatedAt.Before(out[j].CreatedAt)
        }
        if out[i].EventHash != out[j].EventHash {
            return out[i].EventHash < out[j].EventHash
        }
        return out[i].EventID < out[j].EventID
    })
    return out, nil
}

func Heads(records []model.EventRecord) []string {
    g := Build(records)
    return g.Heads
}

func MissingParents(records []model.EventRecord) []string {
    seen := map[string]struct{}{}
    for _, r := range records {
        seen[r.EventHash] = struct{}{}
    }
    missing := map[string]struct{}{}
    for _, r := range records {
        for _, p := range r.ParentHashes {
            if _, ok := seen[p]; !ok {
                missing[p] = struct{}{}
            }
        }
    }
    out := make([]string, 0, len(missing))
    for m := range missing {
        out = append(out, m)
    }
    sort.Strings(out)
    return out
}

func EnsureNoDuplicates(records []model.EventRecord) error {
    seen := map[string]struct{}{}
    for _, r := range records {
        if _, ok := seen[r.EventHash]; ok {
            return errors.New("duplicate event hash detected: " + r.EventHash)
        }
        seen[r.EventHash] = struct{}{}
    }
    return nil
}
