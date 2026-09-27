package capability

import (
	"context"

	"synx/internal/protocol"
)

// Capability defines the interface that all modular SynX functionalities implement.
type Capability interface {
	// Name returns the unique canonical capability identifier (e.g. "terminal", "command", "files", "clipboard").
	Name() string

	// Version returns the semver version string of this capability implementation.
	Version() string

	// Actions returns the list of action identifiers this capability can process.
	Actions() []string

	// Handle processes an incoming request and returns a structured response.
	Handle(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
}
