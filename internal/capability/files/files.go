package files

import (
	"context"

	"synx/internal/protocol"
)

// Capability wraps chunked file streaming operations into the SynX capability model.
type Capability struct {
	sharedDir string
}

// New creates a new Files capability.
func New(sharedDir string) *Capability {
	return &Capability{
		sharedDir: sharedDir,
	}
}

func (c *Capability) Name() string {
	return "files"
}

func (c *Capability) Version() string {
	return "1.0.0"
}

func (c *Capability) Actions() []string {
	return []string{
		"info",
		"list",
	}
}

func (c *Capability) Handle(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	switch req.Action {
	case "info":
		return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]any{
			"chunk_size":     8 * 1024 * 1024,
			"resumable":      true,
			"checksum":       "sha256",
			"shared_dir_set": c.sharedDir != "",
		}), nil
	default:
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrActionNotSupported, "action not supported"), nil
	}
}
