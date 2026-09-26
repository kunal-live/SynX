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

type AnnouncePacket struct {
	Protocol  string `json:"protocol"`
	Version   int    `json:"version"`
	DeviceID  string `json:"device_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	Port      int    `json:"port"`
	Token     string `json:"token"`
	PublicKey string `json:"public_key"`
}

type Discovery struct {
	port       int
	httpPort   int
	id         *identity.Identity
	token      string
	peerMgr    *peers.Manager
	listener   *net.UDPConn
	stopCh     chan struct{}
	running    bool
	mu         sync.Mutex
}

func New(discoveryPort, httpPort int, id *identity.Identity, token string, peerMgr *peers.Manager) *Discovery {
	return &Discovery{
		port:     discoveryPort,
		httpPort: httpPort,
		id:       id,
		token:    token,
		peerMgr:  peerMgr,
		stopCh:   make(chan struct{}),
	}
}

func (d *Discovery) Start() error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return nil
	}
	d.running = true
	d.mu.Unlock()

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", d.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("failed to bind UDP discovery port: %w", err)
	}
	d.listener = conn

	go d.listenLoop()
	go d.announceLoop()

	observability.Info("Peer discovery started on UDP port %d", d.port)
	return nil
}

func (d *Discovery) Stop() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}
	d.running = false
	close(d.stopCh)
	if d.listener != nil {
		_ = d.listener.Close()
	}
	d.mu.Unlock()
}

func (d *Discovery) announceLoop() {
	dst, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", d.port))
	if err != nil {
		return
	}

	outConn, err := net.DialUDP("udp4", nil, dst)
	if err != nil {
		return
	}
	defer outConn.Close()

	ticker := time.NewTicker(2500 * time.Millisecond)
	defer ticker.Stop()

	// Initial immediate ping
	d.broadcast(outConn)

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.broadcast(outConn)
		}
	}
}

func (d *Discovery) broadcast(conn *net.UDPConn) {
	summary := d.id.Summary()
	pkt := AnnouncePacket{
		Protocol:  "synx",
		Version:   1,
		DeviceID:  summary["device_id"],
		Name:      summary["device_name"],
		Platform:  runtime.GOOS,
		Port:      d.httpPort,
		Token:     d.token,
		PublicKey: summary["public_key"],
	}

	data, err := json.Marshal(pkt)
	if err != nil {
		return
	}

	_, _ = conn.Write(data)
}

func (d *Discovery) listenLoop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-d.stopCh:
			return
		default:
		}

		n, remoteAddr, err := d.listener.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-d.stopCh:
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		var pkt AnnouncePacket
		if err := json.Unmarshal(buf[:n], &pkt); err != nil {
			continue
		}

		// Ignore packets from self
		if pkt.Protocol != "synx" || pkt.DeviceID == d.id.DeviceID {
			continue
		}

		peerHost := remoteAddr.IP.String()
		port := pkt.Port
		if port == 0 {
			port = 8787
		}

		d.peerMgr.AddOrUpdate(peers.Peer{
			ID:        pkt.DeviceID,
			Name:      pkt.Name,
			Address:   fmt.Sprintf("%s:%d", peerHost, port),
			Port:      port,
			Platform:  pkt.Platform,
			Version:   fmt.Sprintf("v%d", pkt.Version),
			Token:     pkt.Token,
			PublicKey: pkt.PublicKey,
			Status:    peers.StatusReachable,
		})
	}
}
