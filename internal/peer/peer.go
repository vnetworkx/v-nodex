package peer

import (
    "sort"
    "time"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Registry struct {
    Peers []model.Peer `json:"peers"`
}

func ScorePeer(p model.Peer, invalidEvents int, syncFailures int, bytesDropped int) model.Peer {
    score := 100
    score -= invalidEvents * 10
    score -= syncFailures * 5
    if bytesDropped > 0 {
        score -= 1
    }
    if score < 0 {
        score = 0
    }
    p.Score = score
    p.LastSeen = time.Now().UTC()
    if p.BaseURL == "" && p.Multiaddr == "" {
        p.Healthy = false
    }
    return p
}

func RankPeers(peers []model.Peer) []model.Peer {
    out := append([]model.Peer(nil), peers...)
    sort.SliceStable(out, func(i, j int) bool {
        if out[i].Score != out[j].Score {
            return out[i].Score > out[j].Score
        }
        if !out[i].LastSeen.Equal(out[j].LastSeen) {
            return out[i].LastSeen.After(out[j].LastSeen)
        }
        return out[i].PeerID < out[j].PeerID
    })
    return out
}
