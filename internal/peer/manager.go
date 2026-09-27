package peer

import (
	"sync"
	"time"

	"synx/internal/observability"
)

// EventListener is invoked on peer state mutations.
type EventListener func(event string, p Peer)

// Manager maintains the live peer table and enforces peer lifecycle state transitions.
type Manager struct {
	mu        sync.RWMutex
	peers     map[string]*Peer
	listeners []EventListener
	timeout   time.Duration
	stopCh    chan struct{}
}

// NewManager creates a new Peer Manager with an expiration timeout.
func NewManager(timeout time.Duration) *Manager {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	m := &Manager{
		peers:   make(map[string]*Peer),
		timeout: timeout,
		stopCh:  make(chan struct{}),
	}
	go m.reaperLoop()
	return m
}

// Subscribe registers an event listener for peer lifecycle events.
func (m *Manager) Subscribe(l EventListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, l)
}

// AddOrUpdate registers a newly discovered peer or updates an existing peer's last seen time.
func (m *Manager) AddOrUpdate(incoming Peer) {
	id := incoming.DeviceID
	if id == "" {
		id = incoming.ID
	}
	if id == "" {
		return
	}

	m.mu.Lock()
	existing, found := m.peers[id]

	if !found {
		incoming.DeviceID = id
		incoming.ID = id
		if incoming.DeviceName == "" && incoming.Name != "" {
			incoming.DeviceName = incoming.Name
		}
		if incoming.Name == "" {
			incoming.Name = incoming.DeviceName
		}
		incoming.LastSeen = time.Now()
		if incoming.Status == "" {
			incoming.Status = StateAvailable
		}
		m.peers[id] = &incoming
		m.mu.Unlock()

		observability.Info("New peer discovered: %s (%s)", incoming.DeviceName, id)
		m.emit("peer.discovered", incoming)
		return
	}

	// Update existing record
	existing.LastSeen = time.Now()
	if incoming.DeviceName != "" {
		existing.DeviceName = incoming.DeviceName
		existing.Name = incoming.DeviceName
	}
	if incoming.Port != 0 {
		existing.Port = incoming.Port
	}
	if incoming.Platform != "" {
		existing.Platform = incoming.Platform
	}
	if len(incoming.IPAddresses) > 0 {
		existing.IPAddresses = incoming.IPAddresses
	}
	if incoming.Address != "" {
		existing.Address = incoming.Address
	}
	if incoming.PublicKey != "" {
		existing.PublicKey = incoming.PublicKey
	}
	if len(incoming.Capabilities) > 0 {
		existing.Capabilities = incoming.Capabilities
	}
	if incoming.Token != "" {
		existing.Token = incoming.Token
	}

	if existing.Status == StateDisconnected {
		existing.Status = StateAvailable
	}

	cpy := *existing
	m.mu.Unlock()

	m.emit("peer.updated", cpy)
}

// Get returns a thread-safe snapshot of a peer by its device ID.
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

// List returns a snapshot slice of all known peers.
func (m *Manager) List() []Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]Peer, 0, len(m.peers))
	for _, p := range m.peers {
		res = append(res, *p)
	}
	return res
}

// SetStatus updates a peer's state.
func (m *Manager) SetStatus(id string, status State) {
	m.mu.Lock()
	p, ok := m.peers[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	p.Status = status
	cpy := *p
	m.mu.Unlock()

	m.emit("peer.status_changed", cpy)
}

// SetTrusted updates whether a peer is cryptographically trusted.
func (m *Manager) SetTrusted(id string, trusted bool) {
	m.mu.Lock()
	p, ok := m.peers[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	p.Trusted = trusted
	cpy := *p
	m.mu.Unlock()

	m.emit("peer.trusted_changed", cpy)
}

// Remove drops a peer from the table.
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	p, ok := m.peers[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	delete(m.peers, id)
	cpy := *p
	m.mu.Unlock()

	m.emit("peer.removed", cpy)
}

// Close stops the reaper loop.
func (m *Manager) Close() {
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
}

func (m *Manager) emit(event string, p Peer) {
	m.mu.RLock()
	listeners := make([]EventListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.RUnlock()

	for _, l := range listeners {
		l(event, p)
	}
}

func (m *Manager) reaperLoop() {
	ticker := time.NewTicker(m.timeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			m.mu.Lock()
			for _, p := range m.peers {
				if p.Status != StateDisconnected && now.Sub(p.LastSeen) > m.timeout {
					p.Status = StateDisconnected
					cpy := *p
					go m.emit("peer.disconnected", cpy)
				}
			}
			m.mu.Unlock()
		}
	}
}
