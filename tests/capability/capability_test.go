package capability_test

import (
	"context"
	"testing"

	"synx/internal/capability"
	"synx/internal/capability/clipboard"
	"synx/internal/capability/command"
	"synx/internal/capability/files"
	"synx/internal/capability/terminal"
	"synx/internal/protocol"
)

func TestCapabilityRegistryAndDispatch(t *testing.T) {
	reg := capability.NewRegistry()

	clip := clipboard.New(true)
	cmd := command.New(func(senderID string) bool { return true })
	term := terminal.New(nil)
	fileCap := files.New("/tmp/test")

	_ = reg.Register(clip)
	_ = reg.Register(cmd)
	_ = reg.Register(term)
	_ = reg.Register(fileCap)

	names := reg.ListNames()
	if len(names) != 4 {
		t.Errorf("expected 4 capabilities, got %d", len(names))
	}

	// Test clipboard set & get
	ctx := context.Background()

	setReq := &protocol.Request{
		MessageID:  protocol.GenerateMessageID(),
		Capability: "clipboard",
		Action:     "set",
		Payload:    []byte(`{"text":"hello developer mesh"}`),
	}
	resp := reg.Dispatch(ctx, setReq)
	if !resp.Success {
		t.Fatalf("clipboard set failed: %v", resp.Error)
	}

	getReq := &protocol.Request{
		MessageID:  protocol.GenerateMessageID(),
		Capability: "clipboard",
		Action:     "get",
	}
	resp = reg.Dispatch(ctx, getReq)
	if !resp.Success {
		t.Fatalf("clipboard get failed: %v", resp.Error)
	}

	// Test invalid capability
	badReq := &protocol.Request{
		MessageID:  protocol.GenerateMessageID(),
		Capability: "non_existent",
		Action:     "foo",
	}
	resp = reg.Dispatch(ctx, badReq)
	if resp.Success {
		t.Errorf("expected failure for non-existent capability")
	}
	if resp.Error.Code != protocol.ErrCapabilityNotFound {
		t.Errorf("expected CAPABILITY_NOT_FOUND, got %s", resp.Error.Code)
	}
}
