package session

import (
	"sync"
	"time"

	"synx/internal/network"
	"synx/internal/observability"
)

// PermissionSet tracks authorized capabilities for a specific peer.
type PermissionSet map[string]bool

// Session represents an authenticated active connection session with a peer.
type Session struct {
	ID          string
	PeerID      string
	PeerName    string
	Conn        network.Connection
	CreatedAt   time.Time
	LastActive  time.Time
	Permissions PermissionSet
	mu          sync.RWMutex
}

// HasPermission checks if the peer is authorized for a capability.
func (s *Session) HasPermission(capability string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.Permissions == nil {
		return false
	}
	return s.Permissions[capability]
}

// SetPermission toggles permission for a capability.
func (s *Session) SetPermission(capability string, allowed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Permissions == nil {
		s.Permissions = make(PermissionSet)
	}
	s.Permissions[capability] = allowed
}

// Manager maintains all live peer sessions.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewManager creates a session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// Create establishes and registers a new active session.
func (m *Manager) Create(sessionID, peerID, peerName string, conn network.Connection, initialPerms PermissionSet) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if initialPerms == nil {
		initialPerms = PermissionSet{
			"terminal":  true,
			"files":     true,
			"clipboard": true,
			"command":   true,
		}
	}

	s := &Session{
		ID:          sessionID,
		PeerID:      peerID,
		PeerName:    peerName,
		Conn:        conn,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		Permissions: initialPerms,
	}

	m.sessions[sessionID] = s
	observability.Info("Session created: %s for peer %s (%s)", sessionID, peerName, peerID)
	return s
}

// Get finds an active session by ID.
func (m *Manager) Get(sessionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

// GetByPeer finds an active session by peer device ID.
func (m *Manager) GetByPeer(peerID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		if s.PeerID == peerID {
			return s, true
		}
	}
	return nil, false
}

// Close terminates and removes a session.
func (m *Manager) Close(sessionID string) {
	m.mu.Lock()
	s, ok := m.sessions[sessionID]
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	if ok && s.Conn != nil {
		_ = s.Conn.Close()
		observability.Info("Session closed: %s (peer %s)", sessionID, s.PeerName)
	}
}

// List returns a list of summary records for all active sessions.
func (m *Manager) List() []map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]map[string]any, 0, len(m.sessions))
	for _, s := range m.sessions {
		s.mu.RLock()
		perms := make(map[string]bool)
		for k, v := range s.Permissions {
			perms[k] = v
		}
		s.mu.RUnlock()

		out = append(out, map[string]any{
			"session_id":  s.ID,
			"peer_id":     s.PeerID,
			"peer_name":   s.PeerName,
			"created_at":  s.CreatedAt.Unix(),
			"last_active": s.LastActive.Unix(),
			"permissions": perms,
		})
	}
	return out
}
