package p2p

import (
	"context"
	"encoding/json"
	"time"

	"vnodex/internal/record"
)

type Mesh struct {
	SelfID    string
	Transport Transport
	Registry  *Registry
	Gossip    *GossipManager

	OnRecord func(record.Record)
	OnPeer   func(Peer)
}

func NewMesh(selfID string, transport Transport, fanout int, ttl time.Duration) *Mesh {
	registry := NewRegistry()
	gossip := NewGossipManager(selfID, transport, registry, fanout, ttl)

	m := &Mesh{
		SelfID:    selfID,
		Transport: transport,
		Registry:  registry,
		Gossip:    gossip,
	}
	gossip.Handler = m.handleEnvelope
	return m
}

func (m *Mesh) Start(ctx context.Context) error {
	return m.Gossip.Run(ctx)
}

func (m *Mesh) AddPeer(peer Peer) {
	m.Registry.Upsert(peer)
}

func (m *Mesh) AnnouncePeer(ctx context.Context, peer Peer) error {
	payload, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	_, err = m.Gossip.Publish(ctx, "peer_announce", payload, peer.Region, nil)
	return err
}

func (m *Mesh) BroadcastRecord(ctx context.Context, r record.Record) error {
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = m.Gossip.Publish(ctx, "record", payload, r.RegionID, r.Parents)
	return err
}

func (m *Mesh) handleEnvelope(ctx context.Context, env Envelope) error {
	switch env.Type {
	case "peer_announce":
		var peer Peer
		if err := json.Unmarshal(env.Payload, &peer); err != nil {
			return err
		}
		m.Registry.Upsert(peer)
		if m.OnPeer != nil {
			m.OnPeer(peer)
		}
	case "record":
		var rec record.Record
		if err := json.Unmarshal(env.Payload, &rec); err != nil {
			return err
		}
		if m.OnRecord != nil {
			m.OnRecord(rec)
		}
	default:
		// no-op
	}
	return nil
}
