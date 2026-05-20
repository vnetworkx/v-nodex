package spatial

import (
    "sort"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type RegionIndex struct {
    ByRegion map[string][]model.EventRecord
    ByEntity map[string][]model.EventRecord
}

func Build(records []model.EventRecord) RegionIndex {
    idx := RegionIndex{
        ByRegion: map[string][]model.EventRecord{},
        ByEntity: map[string][]model.EventRecord{},
    }
    for _, r := range records {
        idx.ByRegion[r.RegionID] = append(idx.ByRegion[r.RegionID], r)
        idx.ByEntity[r.EntityID] = append(idx.ByEntity[r.EntityID], r)
    }
    for k := range idx.ByRegion {
        sort.SliceStable(idx.ByRegion[k], func(i, j int) bool {
            return idx.ByRegion[k][i].LogicalClock < idx.ByRegion[k][j].LogicalClock
        })
    }
    for k := range idx.ByEntity {
        sort.SliceStable(idx.ByEntity[k], func(i, j int) bool {
            return idx.ByEntity[k][i].LogicalClock < idx.ByEntity[k][j].LogicalClock
        })
    }
    return idx
}
