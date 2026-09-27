package protocol

import (
	"encoding/json"
	"time"
)

// Event encapsulates an asynchronous streaming event from a peer.
type Event struct {
	MessageID  string          `json:"message_id"`
	Capability string          `json:"capability"`
	Action     string          `json:"action"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

// NewEvent creates a new Event.
func NewEvent(capability, action string, payload any) *Event {
	var raw json.RawMessage
	if payload != nil {
		raw, _ = json.Marshal(payload)
	}
	return &Event{
		MessageID:  GenerateMessageID(),
		Capability: capability,
		Action:     action,
		Payload:    raw,
		Timestamp:  time.Now().Unix(),
	}
}

// ToMessage converts the Event into a protocol Message envelope.
func (e *Event) ToMessage() *Message {
	return &Message{
		Version:    EnvelopeVersion,
		MessageID:  e.MessageID,
		Type:       TypeEvent,
		Capability: e.Capability,
		Action:     e.Action,
		Payload:    e.Payload,
		Timestamp:  e.Timestamp,
	}
}
