package discovery_test

import (
	"testing"
	"time"

	"synx/internal/discovery"
	"synx/internal/peer"
)

func TestDiscoveryMessage(t *testing.T) {
	msg := discovery.NewHelloMessage("synx_12345", "Test-Device", "windows", 8787, []string{"terminal", "files"}, "pubkey_hex", "token123")

	payload, err := msg.Encode()
	if err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	decoded, err := discovery.DecodeMessage(payload)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if decoded.DeviceID != "synx_12345" || decoded.DeviceName != "Test-Device" {
		t.Errorf("decoded fields mismatch: %s / %s", decoded.DeviceID, decoded.DeviceName)
	}
	if len(decoded.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(decoded.Capabilities))
	}
}

func TestPeerManagerLifecycle(t *testing.T) {
	mgr := peer.NewManager(100 * time.Millisecond)
	defer mgr.Close()

	mgr.AddOrUpdate(peer.Peer{
		DeviceID:     "synx_node_b",
		DeviceName:   "Machine-B",
		Port:         8787,
		Platform:     "linux",
		Capabilities: []string{"terminal", "command", "files"},
	})

	p, ok := mgr.Get("synx_node_b")
	if !ok {
		t.Fatalf("expected peer to exist")
	}
	if p.Status != peer.StateAvailable {
		t.Errorf("expected state %s, got %s", peer.StateAvailable, p.Status)
	}
	if !p.HasCapability("terminal") {
		t.Errorf("expected peer to have terminal capability")
	}

	// Wait for reaper to transition peer to disconnected
	time.Sleep(250 * time.Millisecond)

	pAfter, ok := mgr.Get("synx_node_b")
	if !ok {
		t.Fatalf("expected peer to remain in table")
	}
	if pAfter.Status != peer.StateDisconnected {
		t.Errorf("expected peer to transition to DISCONNECTED after timeout, got %s", pAfter.Status)
	}
}
