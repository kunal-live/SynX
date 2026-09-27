package discovery

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	ProtocolName    = "synx"
	ProtocolVersion = 1
	MessageTypeHello = "hello"
)

// Message represents the standardized LAN announcement datagram.
type Message struct {
	Type            string   `json:"type"`
	Protocol        string   `json:"protocol"`
	ProtocolVersion int      `json:"protocol_version"`
	DeviceID        string   `json:"device_id"`
	DeviceName      string   `json:"device_name"`
	Platform        string   `json:"platform"`
	Port            int      `json:"port"`
	Capabilities    []string `json:"capabilities"`
	PublicKey       string   `json:"public_key"`
	Timestamp       int64    `json:"timestamp"`

	// Optional legacy token for pairing handshake fallback
	Token string `json:"token,omitempty"`
}

// NewHelloMessage constructs a compliant discovery broadcast packet.
func NewHelloMessage(deviceID, deviceName, platform string, port int, capabilities []string, pubKey, token string) *Message {
	return &Message{
		Type:            MessageTypeHello,
		Protocol:        ProtocolName,
		ProtocolVersion: ProtocolVersion,
		DeviceID:        deviceID,
		DeviceName:      deviceName,
		Platform:        platform,
		Port:            port,
		Capabilities:    capabilities,
		PublicKey:       pubKey,
		Timestamp:       time.Now().Unix(),
		Token:           token,
	}
}

// Encode serializes the discovery message to JSON bytes.
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// DecodeMessage parses bytes into a validated discovery Message.
func DecodeMessage(data []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("malformed discovery json: %w", err)
	}

	if m.Protocol != ProtocolName {
		return nil, fmt.Errorf("unsupported protocol %q", m.Protocol)
	}
	if m.ProtocolVersion <= 0 {
		return nil, fmt.Errorf("invalid protocol version %d", m.ProtocolVersion)
	}
	if m.DeviceID == "" {
		return nil, fmt.Errorf("missing device_id")
	}

	return &m, nil
}
