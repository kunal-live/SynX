package peers

import (
	"sync"
	"time"

	"synx/internal/events"
	"synx/internal/storage"
)

type PeerStatus string

const (
	StatusDiscovered PeerStatus = "discovered"
	StatusReachable  PeerStatus = "reachable"
	StatusPaired     PeerStatus = "paired"
	StatusTrusted    PeerStatus = "trusted"
	StatusOffline    PeerStatus = "offline"
)

type Peer struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	Port         int        `json:"port"`
	Platform     string     `json:"platform"`
	Version      string     `json:"version"`
	Token        string     `json:"token"`
	PublicKey    string     `json:"public_key"`
	Capabilities []string   `json:"capabilities"`
	LastSeen     time.Time  `json:"last_seen"`
	Status       PeerStatus `json:"status"`
	Trusted      bool       `json:"trusted"`
}

type Manager struct {
	mu     sync.RWMutex
	peers  map[string]*Peer
	db     *storage.DB
	bus    *events.Bus
	stopCh chan struct{}
}

func NewManager(db *storage.DB, bus *events.Bus) *Manager {
	m := &Manager{
		peers:  make(map[string]*Peer),
		db:     db,
		bus:    bus,
		stopCh: make(chan struct{}),
	}
	m.loadPersistedDevices()
	go m.cleanupLoop()
	return m
}

func (m *Manager) Close() {
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
}

func (m *Manager) loadPersistedDevices() {
	if m.db == nil {
		return
	}
	devices, err := m.db.ListDevices()
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range devices {
		status := StatusOffline
		if d.Trusted {
			status = StatusTrusted
		}
		m.peers[d.ID] = &Peer{
			ID:        d.ID,
			Name:      d.Name,
			Platform:  d.Platform,
			Version:   d.Version,
			PublicKey: d.PublicKey,
			Trusted:   d.Trusted,
			Status:    status,
			LastSeen:  time.Unix(d.LastSeen, 0),
		}
	}
}

func (m *Manager) AddOrUpdate(p Peer) {
	m.mu.Lock()
	existing, found := m.peers[p.ID]
	if !found {
		p.LastSeen = time.Now()
		if p.Status == "" {
			p.Status = StatusDiscovered
		}
		m.peers[p.ID] = &p
		m.mu.Unlock()

		if m.bus != nil {
			m.bus.Publish("peer.discovered", map[string]any{
				"peer_id":  p.ID,
				"name":     p.Name,
				"address":  p.Address,
				"port":     p.Port,
				"platform": p.Platform,
			})
		}
		if m.db != nil {
			_ = m.db.UpsertDevice(storage.DeviceRecord{
				ID:        p.ID,
				Name:      p.Name,
				Platform:  p.Platform,
				Version:   p.Version,
				PublicKey: p.PublicKey,
				Trusted:   p.Trusted,
				CreatedAt: time.Now().Unix(),
				LastSeen:  time.Now().Unix(),
			})
		}
		return
	}

	// Update existing peer details
	existing.LastSeen = time.Now()
	if p.Name != "" {
		existing.Name = p.Name
	}
	if p.Address != "" {
		existing.Address = p.Address
	}
	if p.Port != 0 {
		existing.Port = p.Port
	}
	if p.Platform != "" {
		existing.Platform = p.Platform
	}
	if p.Token != "" {
		existing.Token = p.Token
	}
	if p.PublicKey != "" {
		existing.PublicKey = p.PublicKey
	}
	if len(p.Capabilities) > 0 {
		existing.Capabilities = p.Capabilities
	}
	if p.Trusted {
		existing.Trusted = true
		existing.Status = StatusTrusted
	} else if existing.Status == StatusOffline {
		existing.Status = StatusReachable
	}
	m.mu.Unlock()

	if m.db != nil {
		_ = m.db.UpsertDevice(storage.DeviceRecord{
			ID:        existing.ID,
			Name:      existing.Name,
			Platform:  existing.Platform,
			Version:   existing.Version,
			PublicKey: existing.PublicKey,
			Trusted:   existing.Trusted,
			LastSeen:  time.Now().Unix(),
		})
	}
}

func (m *Manager) Get(id string) (*Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[id]
	if !ok {
		return nil, false
	}
	cpy := *p
	return &cpy, true
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	delete(m.peers, id)
	m.mu.Unlock()

	if m.bus != nil {
		m.bus.Publish("peer.disconnected", map[string]any{"peer_id": id})
	}
}

func (m *Manager) List() []Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Peer, 0, len(m.peers))
	for _, p := range m.peers {
		out = append(out, *p)
	}
	return out
}

func (m *Manager) SetTrusted(id string, trusted bool) {
	m.mu.Lock()
	if p, ok := m.peers[id]; ok {
		p.Trusted = trusted
		if trusted {
			p.Status = StatusTrusted
		} else {
			p.Status = StatusReachable
		}
	}
	m.mu.Unlock()

	if m.db != nil {
		_ = m.db.SetDeviceTrust(id, trusted)
	}
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for _, p := range m.peers {
				if now.Sub(p.LastSeen) > 15*time.Second && p.Status != StatusOffline {
					p.Status = StatusOffline
					if m.bus != nil {
						m.bus.Publish("peer.disconnected", map[string]any{
							"peer_id": p.ID,
							"name":    p.Name,
						})
					}
				}
			}
			m.mu.Unlock()
		}
	}
}
