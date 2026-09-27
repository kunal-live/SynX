package capability

import (
	"context"
	"fmt"
	"sync"

	"synx/internal/observability"
	"synx/internal/protocol"
)

// Registry manages available capabilities and routes incoming protocol requests.
type Registry struct {
	mu           sync.RWMutex
	capabilities map[string]Capability
}

// NewRegistry initializes an empty capability registry.
func NewRegistry() *Registry {
	return &Registry{
		capabilities: make(map[string]Capability),
	}
}

// Register adds a capability to the registry.
func (r *Registry) Register(cap Capability) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cap.Name()
	if name == "" {
		return fmt.Errorf("capability name cannot be empty")
	}

	r.capabilities[name] = cap
	observability.Info("Registered capability: %s (v%s) with actions %v", name, cap.Version(), cap.Actions())
	return nil
}

// Get returns a capability by name.
func (r *Registry) Get(name string) (Capability, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.capabilities[name]
	return c, ok
}

// ListNames returns the names of all registered capabilities.
func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.capabilities))
	for name := range r.capabilities {
		names = append(names, name)
	}
	return names
}

// Summary returns a mapping of capability name to supported actions and version.
func (r *Registry) Summary() map[string]map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make(map[string]map[string]any)
	for name, cap := range r.capabilities {
		res[name] = map[string]any{
			"version": cap.Version(),
			"actions": cap.Actions(),
		}
	}
	return res
}

// Dispatch executes the incoming protocol request against the matching registered capability.
func (r *Registry) Dispatch(ctx context.Context, req *protocol.Request) *protocol.Response {
	if req == nil {
		return protocol.ErrorResponse("", "", "", protocol.ErrInvalidRequest, "empty request")
	}

	r.mu.RLock()
	cap, exists := r.capabilities[req.Capability]
	r.mu.RUnlock()

	if !exists {
		return protocol.ErrorResponse(
			req.MessageID,
			req.Capability,
			req.Action,
			protocol.ErrCapabilityNotFound,
			fmt.Sprintf("capability %q not found on this node", req.Capability),
		)
	}

	// Verify action support
	actionSupported := false
	for _, act := range cap.Actions() {
		if act == req.Action {
			actionSupported = true
			break
		}
	}

	if !actionSupported {
		return protocol.ErrorResponse(
			req.MessageID,
			req.Capability,
			req.Action,
			protocol.ErrActionNotSupported,
			fmt.Sprintf("action %q not supported by capability %q", req.Action, req.Capability),
		)
	}

	resp, err := cap.Handle(ctx, req)
	if err != nil {
		return protocol.ErrorResponse(
			req.MessageID,
			req.Capability,
			req.Action,
			protocol.ErrInternalError,
			err.Error(),
		)
	}

	if resp == nil {
		return protocol.SuccessResponse(req.MessageID, req.Capability, req.Action, nil)
	}

	return resp
}
