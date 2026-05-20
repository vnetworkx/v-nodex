package p2p

import "github.com/vnetworkx/v-nodex/internal/model"

type Node struct {
    Mesh *Mesh
    Self model.Peer
}
