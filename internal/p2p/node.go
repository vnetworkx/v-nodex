package p2p

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	ProtocolSync = "/vnodex/sync/1.0.0"
	DefaultTopic = "vnodex-events"
)

var (
	ErrNotStarted   = errors.New("p2p: not started")
	ErrBadMultiaddr = errors.New("p2p: invalid bootstrap multiaddr")
)

type Message struct {
	Topic      string          `json:"topic"`
	From       string          `json:"from"`
	Data       json.RawMessage `json:"data"`
	ReceivedAt time.Time       `json:"received_at"`
}

type Node struct {
	ctx       context.Context
	cancel    context.CancelFunc
	host      host.Host
	dht       *dht.IpfsDHT
	ps        *pubsub.PubSub
	mu        sync.RWMutex
	topics    map[string]*pubsub.Topic
	subs      map[string]*pubsub.Subscription
	inbound   chan Message
	started   bool
	keyPath   string
	selfID    string
	bootstrap []string
}

type Config struct {
	ListenAddrs    []string
	BootstrapPeers []string
	DataDir        string
	Topic          string
}

func New(ctx context.Context, cfg Config) (*Node, error) {
	if cfg.Topic == "" {
		cfg.Topic = DefaultTopic
	}
	if len(cfg.ListenAddrs) == 0 {
		cfg.ListenAddrs = []string{"/ip4/0.0.0.0/udp/0/quic-v1"}
	}

	nodeCtx, cancel := context.WithCancel(ctx)

	keyPath := filepath.Join(cfg.DataDir, "p2p.key")
	priv, err := loadOrCreateKey(keyPath)
	if err != nil {
		cancel()
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
		libp2p.DefaultTransports,
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, err
	}

	d, err := dht.New(nodeCtx, h)
	if err != nil {
		_ = h.Close()
		cancel()
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(nodeCtx, h)
	if err != nil {
		_ = h.Close()
		cancel()
		return nil, err
	}

	n := &Node{
		ctx:       nodeCtx,
		cancel:    cancel,
		host:      h,
		dht:       d,
		ps:        ps,
		topics:    make(map[string]*pubsub.Topic),
		subs:      make(map[string]*pubsub.Subscription),
		inbound:   make(chan Message, 1024),
		keyPath:   keyPath,
		selfID:    h.ID().String(),
		bootstrap: append([]string(nil), cfg.BootstrapPeers...),
	}

	if err := n.bootstrapPeers(nodeCtx, cfg.BootstrapPeers); err != nil {
		_ = h.Close()
		cancel()
		return nil, err
	}

	if err := d.Bootstrap(nodeCtx); err != nil {
		_ = h.Close()
		cancel()
		return nil, err
	}

	if err := n.ensureTopic(cfg.Topic); err != nil {
		_ = h.Close()
		cancel()
		return nil, err
	}

	h.SetStreamHandler(ProtocolSync, n.handleSyncStream)

	n.started = true
	go n.runSubscriptions()

	return n, nil
}

func (n *Node) Host() host.Host { return n.host }

func (n *Node) ID() string {
	if n == nil || n.host == nil {
		return ""
	}
	return n.host.ID().String()
}

func (n *Node) Inbound() <-chan Message { return n.inbound }

func (n *Node) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.cancel != nil {
		n.cancel()
	}
	if n.host != nil {
		_ = n.host.Close()
	}
	if n.inbound != nil {
		close(n.inbound)
	}
	n.started = false
	return nil
}

func (n *Node) ConnectBootstrapPeers(ctx context.Context) error {
	return n.bootstrapPeers(ctx, n.bootstrap)
}

func (n *Node) EnsureTopic(topicName string) error {
	return n.ensureTopic(topicName)
}

func (n *Node) Subscribe(topicName string) (<-chan Message, error) {
	if err := n.ensureTopic(topicName); err != nil {
		return nil, err
	}

	n.mu.RLock()
	sub := n.subs[topicName]
	n.mu.RUnlock()

	if sub == nil {
		return nil, ErrNotStarted
	}

	out := make(chan Message, 256)
	go func() {
		defer close(out)
		for {
			msg, err := sub.Next(n.ctx)
			if err != nil {
				return
			}
			if msg == nil {
				continue
			}
			out <- Message{
				Topic:      topicName,
				From:       msg.ReceivedFrom.String(),
				Data:       append([]byte(nil), msg.Data...),
				ReceivedAt: time.Now().UTC(),
			}
		}
	}()

	return out, nil
}

func (n *Node) Publish(topicName string, payload []byte) error {
	if err := n.ensureTopic(topicName); err != nil {
		return err
	}

	n.mu.RLock()
	t := n.topics[topicName]
	n.mu.RUnlock()

	if t == nil {
		return ErrNotStarted
	}

	return t.Publish(n.ctx, payload)
}

func (n *Node) SendSync(ctx context.Context, peerID string, payload []byte) error {
	if n.host == nil {
		return ErrNotStarted
	}

	id, err := peer.Decode(peerID)
	if err != nil {
		return err
	}

	stream, err := n.host.NewStream(ctx, id, ProtocolSync)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = stream.Write(append(payload, '\n'))
	return err
}

func (n *Node) AddPeerAddr(addr string) error {
	if n.host == nil {
		return ErrNotStarted
	}

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadMultiaddr, err)
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadMultiaddr, err)
	}

	n.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	return n.host.Connect(n.ctx, *info)
}

func (n *Node) ListPeers() []peer.AddrInfo {
	if n.host == nil {
		return nil
	}

	infos := make([]peer.AddrInfo, 0)
	for _, p := range n.host.Network().Peers() {
		infos = append(infos, peer.AddrInfo{
			ID:    p,
			Addrs: n.host.Peerstore().Addrs(p),
		})
	}
	return infos
}

func (n *Node) ensureTopic(topicName string) error {
	if n.ps == nil {
		return ErrNotStarted
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, ok := n.topics[topicName]; ok {
		return nil
	}

	t, err := n.ps.Join(topicName)
	if err != nil {
		return err
	}

	sub, err := t.Subscribe()
	if err != nil {
		return err
	}

	n.topics[topicName] = t
	n.subs[topicName] = sub
	return nil
}

func (n *Node) bootstrapPeers(ctx context.Context, addrs []string) error {
	for _, raw := range addrs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		maddr, err := ma.NewMultiaddr(raw)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrBadMultiaddr, raw)
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrBadMultiaddr, raw)
		}

		n.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		_ = n.host.Connect(ctx, *info)
	}
	return nil
}

func (n *Node) runSubscriptions() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}

		n.mu.RLock()
		subs := make([]*pubsub.Subscription, 0, len(n.subs))
		names := make([]string, 0, len(n.subs))
		for name, sub := range n.subs {
			names = append(names, name)
			subs = append(subs, sub)
		}
		n.mu.RUnlock()

		for i, sub := range subs {
			msg, err := sub.Next(n.ctx)
			if err != nil {
				continue
			}
			if msg == nil {
				continue
			}

			select {
			case n.inbound <- Message{
				Topic:      names[i],
				From:       msg.ReceivedFrom.String(),
				Data:       append([]byte(nil), msg.Data...),
				ReceivedAt: time.Now().UTC(),
			}:
			default:
			}
		}
	}
}

func (n *Node) handleSyncStream(s network.Stream) {
	defer s.Close()

	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)

		select {
		case n.inbound <- Message{
			Topic:      ProtocolSync,
			From:       s.Conn().RemotePeer().String(),
			Data:       line,
			ReceivedAt: time.Now().UTC(),
		}:
		default:
		}
	}
}

func loadOrCreateKey(path string) (crypto.PrivKey, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		priv, err := crypto.UnmarshalPrivateKey(data)
		if err == nil {
			return priv, nil
		}
	}

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	raw, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return nil, err
	}

	return priv, nil
}
