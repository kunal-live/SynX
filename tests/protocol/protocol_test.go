package protocol_test

import (
	"testing"

	"synx/internal/protocol"
)

func TestMessageEnvelope(t *testing.T) {
	msg, err := protocol.NewMessage(protocol.TypeRequest, "terminal", "open", map[string]int{
		"cols": 80,
		"rows": 24,
	})
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if msg.Version != protocol.EnvelopeVersion {
		t.Errorf("expected version %d, got %d", protocol.EnvelopeVersion, msg.Version)
	}
	if msg.MessageID == "" {
		t.Errorf("expected non-empty message ID")
	}

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	decoded, err := protocol.DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}

	if decoded.Capability != "terminal" || decoded.Action != "open" {
		t.Errorf("mismatch decoded capability or action: %s.%s", decoded.Capability, decoded.Action)
	}

	var payload struct {
		Cols int `json:"cols"`
		Rows int `json:"rows"`
	}
	if err := decoded.ParsePayload(&payload); err != nil {
		t.Fatalf("failed to parse payload: %v", err)
	}
	if payload.Cols != 80 || payload.Rows != 24 {
		t.Errorf("expected 80x24, got %dx%d", payload.Cols, payload.Rows)
	}
}

func TestResponseErrors(t *testing.T) {
	errResp := protocol.ErrorResponse("msg-123", "command", "execute", protocol.ErrForbidden, "unauthorized command")
	if errResp.Success {
		t.Errorf("expected Success=false")
	}
	if errResp.Error == nil || errResp.Error.Code != protocol.ErrForbidden {
		t.Errorf("expected code %s, got %v", protocol.ErrForbidden, errResp.Error)
	}
}
