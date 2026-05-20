package p2p

import (
    "context"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Mesh struct {
    Transport Transport
    Gossip    *Gossip
}

func NewMesh(transport Transport) *Mesh {
    return &Mesh{Transport: transport, Gossip: NewGossip()}
}

func (m *Mesh) SyncPeer(ctx context.Context, peer model.Peer, limit int) ([]model.EventRecord, error) {
    if peer.BaseURL == "" {
        return nil, nil
    }
    events, err := m.Transport.FetchEvents(ctx, peer.BaseURL, limit, 0)
    if err != nil {
        return nil, err
    }
    return events, nil
}
