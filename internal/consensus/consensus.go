package consensus

import (
    "sort"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Selection struct {
    HeadHashes []string `json:"head_hashes"`
    WinnerHash  string   `json:"winner_hash"`
}

func SelectCanonical(records []model.EventRecord) Selection {
    if len(records) == 0 {
        return Selection{}
    }
    ordered := append([]model.EventRecord(nil), records...)
    sort.SliceStable(ordered, func(i, j int) bool {
        if ordered[i].LogicalClock != ordered[j].LogicalClock {
            return ordered[i].LogicalClock > ordered[j].LogicalClock
        }
        if ordered[i].StateRoot != ordered[j].StateRoot {
            return ordered[i].StateRoot < ordered[j].StateRoot
        }
        if ordered[i].EventHash != ordered[j].EventHash {
            return ordered[i].EventHash < ordered[j].EventHash
        }
        return ordered[i].EventID < ordered[j].EventID
    })
    heads := make([]string, 0, 1)
    heads = append(heads, ordered[0].EventHash)
    return Selection{
        HeadHashes: heads,
        WinnerHash:  ordered[0].EventHash,
    }
}

func SortByCanonicalPriority(records []model.EventRecord) []model.EventRecord {
    ordered := append([]model.EventRecord(nil), records...)
    sort.SliceStable(ordered, func(i, j int) bool {
        if ordered[i].LogicalClock != ordered[j].LogicalClock {
            return ordered[i].LogicalClock < ordered[j].LogicalClock
        }
        if ordered[i].StateRoot != ordered[j].StateRoot {
            return ordered[i].StateRoot < ordered[j].StateRoot
        }
        if ordered[i].EventHash != ordered[j].EventHash {
            return ordered[i].EventHash < ordered[j].EventHash
        }
        return ordered[i].EventID < ordered[j].EventID
    })
    return ordered
}
