package p2p

import (
    "context"

    "github.com/vnetworkx/v-nodex/internal/model"
)

type Discovery struct {
    Transport Transport
}

func (d *Discovery) Announce(ctx context.Context, peerURL string, me model.Peer) error {
    return d.Transport.AnnouncePeer(ctx, peerURL, me)
}
