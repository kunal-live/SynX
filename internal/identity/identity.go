package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Identity represents the persistent device identity for this SynX node.
type Identity struct {
	DeviceID     string   `json:"device_id"`
	DeviceName   string   `json:"device_name"`
	Platform     string   `json:"platform"`
	Architecture string   `json:"architecture"`
	Version      string   `json:"version"`
	PublicKey    string   `json:"public_key"`
	Capabilities []string `json:"capabilities"`
	privateKey   ed25519.PrivateKey
	mu           sync.RWMutex
}

type persistedIdentity struct {
	DeviceID     string   `json:"device_id"`
	DeviceName   string   `json:"device_name"`
	Platform     string   `json:"platform"`
	Architecture string   `json:"architecture"`
	Version      string   `json:"version"`
	PublicKey    string   `json:"public_key"`
	PrivateKey   string   `json:"private_key"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// DefaultCapabilities returns the core set of developer capabilities supported by this node.
func DefaultCapabilities() []string {
	return []string{
		"terminal",
		"command",
		"files",
		"clipboard",
		"events",
	}
}

// LoadOrCreate initializes the device identity from disk, or generates a fresh persistent identity.
func LoadOrCreate(dataDir, preferredName string) (*Identity, error) {
	_ = os.MkdirAll(dataDir, 0755)
	idPath := filepath.Join(dataDir, "identity.json")

	caps := DefaultCapabilities()

	if data, err := os.ReadFile(idPath); err == nil {
		var p persistedIdentity
		if err := json.Unmarshal(data, &p); err == nil && p.DeviceID != "" && p.PrivateKey != "" {
			priv, err := HexToPrivateKey(p.PrivateKey)
			if err == nil {
				if len(p.Capabilities) > 0 {
					caps = p.Capabilities
				}
				return &Identity{
					DeviceID:     p.DeviceID,
					DeviceName:   p.DeviceName,
					Platform:     runtime.GOOS,
					Architecture: runtime.GOARCH,
					Version:      "1.0.0",
					PublicKey:    p.PublicKey,
					Capabilities: caps,
					privateKey:   priv,
				}, nil
			}
		}
	}

	// Generate new Ed25519 keypair and persistent random device ID
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	randomID := make([]byte, 16)
	_, _ = rand.Read(randomID)
	deviceID := "synx_" + hex.EncodeToString(randomID)

	name := preferredName
	if name == "" {
		h, _ := os.Hostname()
		if h != "" {
			name = h
		} else {
			name = "SynX Node"
		}
	}

	id := &Identity{
		DeviceID:     deviceID,
		DeviceName:   name,
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
		Version:      "1.0.0",
		PublicKey:    PublicKeyToHex(pub),
		Capabilities: caps,
		privateKey:   priv,
	}

	p := persistedIdentity{
		DeviceID:     id.DeviceID,
		DeviceName:   id.DeviceName,
		Platform:     id.Platform,
		Architecture: id.Architecture,
		Version:      id.Version,
		PublicKey:    id.PublicKey,
		PrivateKey:   PrivateKeyToHex(priv),
		Capabilities: id.Capabilities,
	}

	if encoded, err := json.MarshalIndent(p, "", "  "); err == nil {
		_ = os.WriteFile(idPath, encoded, 0600)
	}

	return id, nil
}

// Sign signs an arbitrary byte payload using the node's private key.
func (id *Identity) Sign(message []byte) []byte {
	id.mu.RLock()
	defer id.mu.RUnlock()
	return Sign(id.privateKey, message)
}

// VerifySignature checks a signature against a given public key hex and message.
func VerifySignature(publicKeyHex string, message, signature []byte) bool {
	pub, err := HexToPublicKey(publicKeyHex)
	if err != nil {
		return false
	}
	return VerifyKey(pub, message, signature)
}

// Verify is a backward-compatible alias for VerifySignature.
func Verify(publicKeyHex string, message, signature []byte) bool {
	return VerifySignature(publicKeyHex, message, signature)
}

// SetName updates the device name in memory.
func (id *Identity) SetName(name string) {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.DeviceName = name
}

// Summary returns a key-value snapshot of the identity.
func (id *Identity) Summary() map[string]string {
	id.mu.RLock()
	defer id.mu.RUnlock()
	return map[string]string{
		"device_id":    id.DeviceID,
		"device_name":  id.DeviceName,
		"platform":     id.Platform,
		"architecture": id.Architecture,
		"version":      id.Version,
		"public_key":   id.PublicKey,
	}
}
