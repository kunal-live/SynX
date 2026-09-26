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

type Identity struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform"`
	Version    string `json:"version"`
	PublicKey  string `json:"public_key"`
	privateKey ed25519.PrivateKey
	mu         sync.RWMutex
}

type persistedIdentity struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	Platform   string `json:"platform"`
	Version    string `json:"version"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func LoadOrCreate(dataDir, preferredName string) (*Identity, error) {
	_ = os.MkdirAll(dataDir, 0755)
	idPath := filepath.Join(dataDir, "identity.json")

	if data, err := os.ReadFile(idPath); err == nil {
		var p persistedIdentity
		if err := json.Unmarshal(data, &p); err == nil && p.DeviceID != "" && p.PrivateKey != "" {
			privBytes, _ := hex.DecodeString(p.PrivateKey)
			if len(privBytes) == ed25519.PrivateKeySize {
				return &Identity{
					DeviceID:   p.DeviceID,
					DeviceName: p.DeviceName,
					Platform:   runtime.GOOS,
					Version:    "1.0.0",
					PublicKey:  p.PublicKey,
					privateKey: ed25519.PrivateKey(privBytes),
				}, nil
			}
		}
	}

	// Generate new Ed25519 keypair and persistent device ID
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
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
		DeviceID:   deviceID,
		DeviceName: name,
		Platform:   runtime.GOOS,
		Version:    "1.0.0",
		PublicKey:  hex.EncodeToString(pub),
		privateKey: priv,
	}

	p := persistedIdentity{
		DeviceID:   id.DeviceID,
		DeviceName: id.DeviceName,
		Platform:   id.Platform,
		Version:    id.Version,
		PublicKey:  id.PublicKey,
		PrivateKey: hex.EncodeToString(priv),
	}

	if encoded, err := json.MarshalIndent(p, "", "  "); err == nil {
		_ = os.WriteFile(idPath, encoded, 0600)
	}

	return id, nil
}

func (id *Identity) Sign(message []byte) []byte {
	id.mu.RLock()
	defer id.mu.RUnlock()
	return ed25519.Sign(id.privateKey, message)
}

func Verify(publicKeyHex string, message, signature []byte) bool {
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pubBytes), message, signature)
}

func (id *Identity) SetName(name string) {
	id.mu.Lock()
	defer id.mu.Unlock()
	id.DeviceName = name
}

func (id *Identity) Summary() map[string]string {
	id.mu.RLock()
	defer id.mu.RUnlock()
	return map[string]string{
		"device_id":   id.DeviceID,
		"device_name": id.DeviceName,
		"platform":    id.Platform,
		"version":     id.Version,
		"public_key":  id.PublicKey,
	}
}
