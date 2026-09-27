package clipboard

import (
	"context"
	"sync"
	"time"

	"synx/internal/protocol"
)

// Capability manages clipboard text sharing across SynX peers.
type Capability struct {
	mu           sync.RWMutex
	lastText     string
	lastUpdated  time.Time
	syncAllowed  bool
}

// New creates a new Clipboard capability.
func New(syncAllowed bool) *Capability {
	return &Capability{
		syncAllowed: syncAllowed,
	}
}

func (c *Capability) Name() string {
	return "clipboard"
}

func (c *Capability) Version() string {
	return "1.0.0"
}

func (c *Capability) Actions() []string {
	return []string{
		"get",
		"set",
	}
}

type SetPayload struct {
	Text string `json:"text"`
}

func (c *Capability) Handle(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	switch req.Action {
	case "get":
		c.mu.RLock()
		defer c.mu.RUnlock()
		return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]any{
			"text":         c.lastText,
			"last_updated": c.lastUpdated.Unix(),
		}), nil

	case "set":
		var payload SetPayload
		if err := req.ParsePayload(&payload); err != nil {
			return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInvalidRequest, "invalid text payload"), nil
		}

		c.mu.Lock()
		c.lastText = payload.Text
		c.lastUpdated = time.Now()
		c.mu.Unlock()

		return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]bool{"updated": true}), nil

	default:
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrActionNotSupported, "action not supported"), nil
	}
}

// GetText returns the cached clipboard string.
func (c *Capability) GetText() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastText
}

// SetText pushes new text into the capability.
func (c *Capability) SetText(text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastText = text
	c.lastUpdated = time.Now()
}
