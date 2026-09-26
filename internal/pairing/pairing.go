package pairing

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"synx/internal/peers"
	"synx/internal/security"
)

type PairingRequest struct {
	DeviceID  string `json:"device_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	PIN       string `json:"pin"`
}

type PairingResponse struct {
	Success   bool   `json:"success"`
	DeviceID  string `json:"device_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	Token     string `json:"token"`
	Error     string `json:"error,omitempty"`
}

type Service struct {
	mu         sync.Mutex
	currentPIN string
	pinExpiry  time.Time
	trustStore *security.TrustStore
	peerMgr    *peers.Manager
	localToken string
}

func NewService(trustStore *security.TrustStore, peerMgr *peers.Manager, localToken string) *Service {
	return &Service{
		trustStore: trustStore,
		peerMgr:    peerMgr,
		localToken: localToken,
	}
}

// GeneratePIN produces a fresh 6-digit pairing code valid for 5 minutes (Section 11)
func (s *Service) GeneratePIN() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	pin := fmt.Sprintf("%06d", n.Int64()+100000)
	s.currentPIN = pin
	s.pinExpiry = time.Now().Add(5 * time.Minute)
	return pin
}

func (s *Service) CurrentPIN() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPIN == "" || time.Now().After(s.pinExpiry) {
		return "", false
	}
	return s.currentPIN, true
}

func (s *Service) VerifyPIN(pin string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentPIN == "" || time.Now().After(s.pinExpiry) {
		return false
	}
	match := strings.TrimSpace(pin) == s.currentPIN
	if match {
		s.currentPIN = "" // consume on success
	}
	return match
}

// PairDirect connects to a remote SynX node at address using token, verifies identity, and trusts peer
func (s *Service) PairDirect(address, token string) (*peers.Peer, error) {
	address = strings.TrimSpace(address)
	token = strings.TrimSpace(token)
	if address == "" || token == "" {
		return nil, errors.New("address and token are required")
	}

	url := address
	if !strings.Contains(url, "://") {
		url = "http://" + url
	}
	url = strings.TrimRight(url, "/")

	req, err := http.NewRequest(http.MethodGet, url+"/api/v1/device/info", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-SynX-Token", token)
	req.Header.Set("X-SynX-Protocol-Version", "1")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to legacy info endpoint
		req2, _ := http.NewRequest(http.MethodGet, url+"/api/info", nil)
		req2.Header.Set("X-SynX-Token", token)
		resp, err = client.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("pairing connection failed: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pairing rejected by remote peer: HTTP %d", resp.StatusCode)
	}

	var info map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid response from peer: %w", err)
	}

	hostPort := strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
	deviceID := fmt.Sprint(info["device_id"])
	if deviceID == "" || deviceID == "<nil>" {
		deviceID = hostPort
	}
	name := fmt.Sprint(info["device_name"])
	if name == "" || name == "<nil>" {
		name = fmt.Sprint(info["name"])
	}
	if name == "" || name == "<nil>" {
		name = "SynX Peer"
	}
	platform := fmt.Sprint(info["platform"])
	pubKey := fmt.Sprint(info["public_key"])

	peer := peers.Peer{
		ID:        deviceID,
		Name:      name,
		Address:   hostPort,
		Token:     token,
		Platform:  platform,
		PublicKey: pubKey,
		Status:    peers.StatusTrusted,
		Trusted:   true,
		LastSeen:  time.Now(),
	}

	s.peerMgr.AddOrUpdate(peer)
	_ = s.trustStore.TrustPeer(security.TrustedPeer{
		PeerID:    deviceID,
		Name:      name,
		PublicKey: pubKey,
		Token:     token,
		TrustedAt: time.Now(),
		LastSeen:  time.Now(),
	})

	return &peer, nil
}
