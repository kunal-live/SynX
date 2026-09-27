package protocol

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	EnvelopeVersion = 1

	TypeRequest  = "request"
	TypeResponse = "response"
	TypeEvent    = "event"
)

// Message is the standard communication envelope for all SynX peer-to-peer traffic.
type Message struct {
	Version    int             `json:"version"`
	MessageID  string          `json:"message_id"`
	Type       string          `json:"type"` // "request", "response", "event"
	Capability string          `json:"capability"`
	Action     string          `json:"action"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

// GenerateMessageID generates a cryptographically random UUID/hex string.
func GenerateMessageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewMessage creates a fresh envelope with default version and timestamp.
func NewMessage(msgType, capability, action string, payload any) (*Message, error) {
	var rawPayload json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to encode payload: %w", err)
		}
		rawPayload = b
	}

	return &Message{
		Version:    EnvelopeVersion,
		MessageID:  GenerateMessageID(),
		Type:       msgType,
		Capability: capability,
		Action:     action,
		Payload:    rawPayload,
		Timestamp:  time.Now().Unix(),
	}, nil
}

// Encode serializes the message envelope into JSON.
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// DecodeMessage parses bytes into a validated Message envelope.
func DecodeMessage(data []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("malformed protocol envelope: %w", err)
	}
	if m.Version <= 0 {
		return nil, fmt.Errorf("invalid envelope version: %d", m.Version)
	}
	if m.MessageID == "" {
		return nil, fmt.Errorf("missing message_id")
	}
	if m.Type == "" {
		return nil, fmt.Errorf("missing message type")
	}
	return &m, nil
}

// ParsePayload unmarshals the raw payload into a destination target struct.
func (m *Message) ParsePayload(target any) error {
	if len(m.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(m.Payload, target)
}
