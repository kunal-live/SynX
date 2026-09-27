package protocol

import (
	"encoding/json"
	"time"
)

// Standard error codes per technical specification Section 18.
const (
	ErrInvalidRequest            = "INVALID_REQUEST"
	ErrUnauthorized              = "UNAUTHORIZED"
	ErrForbidden                 = "FORBIDDEN"
	ErrPeerNotFound              = "PEER_NOT_FOUND"
	ErrCapabilityNotFound        = "CAPABILITY_NOT_FOUND"
	ErrActionNotSupported        = "ACTION_NOT_SUPPORTED"
	ErrSessionNotFound           = "SESSION_NOT_FOUND"
	ErrTimeout                   = "TIMEOUT"
	ErrTransferFailed            = "TRANSFER_FAILED"
	ErrProtocolVersionUnsupported= "PROTOCOL_VERSION_UNSUPPORTED"
	ErrInternalError             = "INTERNAL_ERROR"
)

// ErrorDetail describes a structured failure returned in a response envelope.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response represents the result of a capability invocation.
type Response struct {
	MessageID  string          `json:"message_id"`
	Success    bool            `json:"success"`
	Capability string          `json:"capability"`
	Action     string          `json:"action"`
	Data       json.RawMessage `json:"data,omitempty"`
	Error      *ErrorDetail    `json:"error,omitempty"`
	Timestamp  int64           `json:"timestamp"`
}

// SuccessResponse creates a successful response payload.
func SuccessResponse(messageID, capability, action string, data any) *Response {
	var raw json.RawMessage
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	return &Response{
		MessageID:  messageID,
		Success:    true,
		Capability: capability,
		Action:     action,
		Data:       raw,
		Timestamp:  time.Now().Unix(),
	}
}

// ErrorResponse creates an error response payload with code and message.
func ErrorResponse(messageID, capability, action, code, errMsg string) *Response {
	return &Response{
		MessageID:  messageID,
		Success:    false,
		Capability: capability,
		Action:     action,
		Error: &ErrorDetail{
			Code:    code,
			Message: errMsg,
		},
		Timestamp: time.Now().Unix(),
	}
}

// ToMessage converts the Response into a protocol Message envelope.
func (r *Response) ToMessage() *Message {
	b, _ := json.Marshal(r)
	return &Message{
		Version:    EnvelopeVersion,
		MessageID:  r.MessageID,
		Type:       TypeResponse,
		Capability: r.Capability,
		Action:     r.Action,
		Payload:    b,
		Timestamp:  r.Timestamp,
	}
}
