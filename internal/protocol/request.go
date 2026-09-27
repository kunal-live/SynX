package protocol

import (
	"encoding/json"
)

// Request encapsulates a client invocation targeting a peer's capability.
type Request struct {
	MessageID  string          `json:"message_id"`
	SenderID   string          `json:"sender_id,omitempty"`
	Capability string          `json:"capability"`
	Action     string          `json:"action"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

// NewRequest creates a Request from a decoded Message.
func NewRequest(m *Message, senderID string) *Request {
	return &Request{
		MessageID:  m.MessageID,
		SenderID:   senderID,
		Capability: m.Capability,
		Action:     m.Action,
		Payload:    m.Payload,
		Timestamp:  m.Timestamp,
	}
}

// ToMessage converts the Request back into a protocol envelope.
func (r *Request) ToMessage() *Message {
	return &Message{
		Version:    EnvelopeVersion,
		MessageID:  r.MessageID,
		Type:       TypeRequest,
		Capability: r.Capability,
		Action:     r.Action,
		Payload:    r.Payload,
		Timestamp:  r.Timestamp,
	}
}

// ParsePayload unmarshals the request payload into target.
func (r *Request) ParsePayload(target any) error {
	if len(r.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(r.Payload, target)
}
