package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"synx/internal/storage"
)

type TrustedPeer struct {
	PeerID    string    `json:"peer_id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	Token     string    `json:"token"`
	TrustedAt time.Time `json:"trusted_at"`
	LastSeen  time.Time `json:"last_seen"`
}

type TrustStore struct {
	mu    sync.RWMutex
	db    *storage.DB
	peers map[string]*TrustedPeer
}

func NewTrustStore(db *storage.DB) *TrustStore {
	ts := &TrustStore{
		db:    db,
		peers: make(map[string]*TrustedPeer),
	}
	ts.loadFromDB()
	return ts
}

func (ts *TrustStore) loadFromDB() {
	if ts.db == nil {
		return
	}
	records, err := ts.db.ListDevices()
	if err != nil {
		return
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, r := range records {
		if r.Trusted {
			ts.peers[r.ID] = &TrustedPeer{
				PeerID:    r.ID,
				Name:      r.Name,
				PublicKey: r.PublicKey,
				TrustedAt: time.Unix(r.CreatedAt, 0),
				LastSeen:  time.Unix(r.LastSeen, 0),
			}
		}
	}
}

func (ts *TrustStore) TrustPeer(p TrustedPeer) error {
	if p.PeerID == "" {
		return fmt.Errorf("peer_id cannot be empty")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if p.TrustedAt.IsZero() {
		p.TrustedAt = time.Now()
	}
	p.LastSeen = time.Now()
	ts.peers[p.PeerID] = &p

	if ts.db != nil {
		_ = ts.db.UpsertDevice(storage.DeviceRecord{
			ID:        p.PeerID,
			Name:      p.Name,
			PublicKey: p.PublicKey,
			Trusted:   true,
			CreatedAt: p.TrustedAt.Unix(),
			LastSeen:  p.LastSeen.Unix(),
		})
		_ = ts.db.SetDeviceTrust(p.PeerID, true)
	}
	return nil
}

func (ts *TrustStore) Get(peerID string) (*TrustedPeer, bool) {
	if peerID == "" {
		return nil, false
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	p, exists := ts.peers[peerID]
	if !exists {
		return nil, false
	}
	cpy := *p
	return &cpy, true
}

func (ts *TrustStore) IsTrusted(peerID string) bool {
	if peerID == "" {
		return false
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	_, exists := ts.peers[peerID]
	return exists
}

func (ts *TrustStore) VerifyToken(peerID, token string) bool {
	if peerID == "" || token == "" {
		return false
	}
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	p, ok := ts.peers[peerID]
	if !ok {
		return false
	}
	return p.Token == token
}

func (ts *TrustStore) RevokePeer(peerID string) error {
	if peerID == "" {
		return nil
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()

	delete(ts.peers, peerID)
	if ts.db != nil {
		_ = ts.db.SetDeviceTrust(peerID, false)
	}
	return nil
}

func (ts *TrustStore) ListTrusted() []TrustedPeer {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	out := make([]TrustedPeer, 0, len(ts.peers))
	for _, p := range ts.peers {
		out = append(out, *p)
	}
	return out
}

func GenerateToken(byteCount int) string {
	if byteCount <= 0 {
		byteCount = 16
	}
	b := make([]byte, byteCount)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (i * 8))
		}
	}
	return hex.EncodeToString(b)
}

