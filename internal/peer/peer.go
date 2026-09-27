package peer

import (
	"time"
)

// State represents the lifecycle state of a discovered peer.
type State string

const (
	StateDiscovered   State = "DISCOVERED"
	StateAvailable    State = "AVAILABLE"
	StateConnecting   State = "CONNECTING"
	StateConnected    State = "CONNECTED"
	StateDisconnected State = "DISCONNECTED"
)

// Peer represents a remote SynX node discovered on the local area network.
type Peer struct {
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	IPAddresses  []string  `json:"ip_addresses"`
	Port         int       `json:"port"`
	Platform     string    `json:"platform"`
	Architecture string    `json:"architecture,omitempty"`
	Version      string    `json:"version"`
	Capabilities []string  `json:"capabilities"`
	PublicKey    string    `json:"public_key"`
	LastSeen     time.Time `json:"last_seen"`
	Status       State     `json:"status"`
	Trusted      bool      `json:"trusted"`

	// Legacy compatibility fields
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
	Token   string `json:"token,omitempty"`
}

// HasCapability checks whether the peer advertises a specific capability.
func (p *Peer) HasCapability(name string) bool {
	for _, c := range p.Capabilities {
		if c == name {
			return true
		}
	}
	return false
}

// IsOnline returns true if the peer has been seen recently and is not disconnected.
func (p *Peer) IsOnline(timeout time.Duration) bool {
	if p.Status == StateDisconnected {
		return false
	}
	return time.Since(p.LastSeen) <= timeout
}
