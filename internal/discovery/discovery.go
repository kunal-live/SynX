package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"

	"synx/internal/identity"
	"synx/internal/observability"
	"synx/internal/peers"
)

// AnnouncePacket maintains compatibility with legacy formats while supporting the new schema.
type AnnouncePacket struct {
	Protocol     string   `json:"protocol"`
	Version      int      `json:"version"`
	DeviceID     string   `json:"device_id"`
	Name         string   `json:"name,omitempty"`
	DeviceName   string   `json:"device_name,omitempty"`
	Platform     string   `json:"platform"`
	Port         int      `json:"port"`
	Token        string   `json:"token,omitempty"`
	PublicKey    string   `json:"public_key"`
	Capabilities []string `json:"capabilities,omitempty"`
	Timestamp    int64    `json:"timestamp,omitempty"`
}

type Discovery struct {
	port        int
	httpPort    int
	id          *identity.Identity
	token       string
	peerMgr     *peers.Manager
	broadcaster *Broadcaster
	listener    *Listener
	mu          sync.Mutex
	running     bool
}

func New(discoveryPort, httpPort int, id *identity.Identity, token string, peerMgr *peers.Manager) *Discovery {
	d := &Discovery{
		port:     discoveryPort,
		httpPort: httpPort,
		id:       id,
		token:    token,
		peerMgr:  peerMgr,
	}

	d.broadcaster = NewBroadcaster(discoveryPort, 2500*time.Millisecond, func() *Message {
		summary := id.Summary()
		return NewHelloMessage(
			summary["device_id"],
			summary["device_name"],
			runtime.GOOS,
			httpPort,
			id.Capabilities,
			summary["public_key"],
			token,
		)
	})

	d.listener = NewListener(discoveryPort, id.DeviceID, func(msg *Message, remoteIP net.IP) {
		name := msg.DeviceName
		if name == "" {
			name = "SynX Node"
		}
		port := msg.Port
		if port == 0 {
			port = httpPort
		}

		peerHost := remoteIP.String()
		peerMgr.AddOrUpdate(peers.Peer{
			ID:           msg.DeviceID,
			Name:         name,
			Address:      fmt.Sprintf("%s:%d", peerHost, port),
			Port:         port,
			Platform:     msg.Platform,
			Version:      fmt.Sprintf("v%d", msg.ProtocolVersion),
			Token:        msg.Token,
			PublicKey:    msg.PublicKey,
			Capabilities: msg.Capabilities,
			Status:       peers.StatusReachable,
		})
	})

	return d
}

func (d *Discovery) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return nil
	}
	d.running = true

	if err := d.listener.Start(); err != nil {
		observability.Warn("Discovery listener failed to bind UDP port %d: %v. Running in broadcast-only mode.", d.port, err)
	}

	if err := d.broadcaster.Start(); err != nil {
		observability.Warn("Discovery broadcaster failed to start: %v", err)
	}

	observability.Info("Peer discovery started on UDP port %d", d.port)
	return nil
}

func (d *Discovery) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}
	d.running = false

	if d.broadcaster != nil {
		d.broadcaster.Stop()
	}
	if d.listener != nil {
		d.listener.Stop()
	}
}

// DecodeAnnouncePacket parses either a legacy or new discovery payload.
func DecodeAnnouncePacket(data []byte) (*AnnouncePacket, error) {
	var pkt AnnouncePacket
	if err := json.Unmarshal(data, &pkt); err != nil {
		return nil, err
	}
	if pkt.Name == "" && pkt.DeviceName != "" {
		pkt.Name = pkt.DeviceName
	}
	return &pkt, nil
}
