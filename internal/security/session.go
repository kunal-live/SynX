package security

import (
	"sync"
	"time"
)

type Session struct {
	SessionID    string    `json:"session_id"`
	PeerID       string    `json:"peer_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Capabilities []string  `json:"capabilities"`
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
}

func (sm *SessionManager) CreateSession(peerID string, capabilities []string) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s := &Session{
		SessionID:    "sess_" + GenerateToken(16),
		PeerID:       peerID,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(sm.ttl),
		Capabilities: capabilities,
	}
	sm.sessions[s.SessionID] = s
	return s
}

func (sm *SessionManager) Validate(sessionID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[sessionID]
	if !ok || time.Now().After(s.ExpiresAt) {
		return nil, false
	}
	return s, true
}

func (sm *SessionManager) Revoke(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}
