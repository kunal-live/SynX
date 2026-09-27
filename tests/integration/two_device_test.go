package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"synx/internal/identity"
	"synx/internal/peer"
)

func TestTwoDeviceSimulatedDiscovery(t *testing.T) {
	tempDirA, err := os.MkdirTemp("", "synx_dev_a_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDirA)

	tempDirB, err := os.MkdirTemp("", "synx_dev_b_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDirB)

	// Machine A Identity
	idA, err := identity.LoadOrCreate(filepath.Join(tempDirA, "config"), "Workstation-A")
	if err != nil {
		t.Fatalf("failed to create Machine A identity: %v", err)
	}

	// Machine B Identity
	idB, err := identity.LoadOrCreate(filepath.Join(tempDirB, "config"), "Laptop-B")
	if err != nil {
		t.Fatalf("failed to create Machine B identity: %v", err)
	}

	if idA.DeviceID == idB.DeviceID {
		t.Errorf("Device IDs must be unique")
	}

	// Machine A Peer Manager
	peerMgrA := peer.NewManager(5 * time.Second)
	defer peerMgrA.Close()

	// Machine B Peer Manager
	peerMgrB := peer.NewManager(5 * time.Second)
	defer peerMgrB.Close()

	// Machine A receives B's announcement
	peerMgrA.AddOrUpdate(peer.Peer{
		DeviceID:     idB.DeviceID,
		DeviceName:   idB.DeviceName,
		Platform:     idB.Platform,
		Version:      idB.Version,
		Port:         8787,
		PublicKey:    idB.PublicKey,
		Capabilities: idB.Capabilities,
	})

	// Machine B receives A's announcement
	peerMgrB.AddOrUpdate(peer.Peer{
		DeviceID:     idA.DeviceID,
		DeviceName:   idA.DeviceName,
		Platform:     idA.Platform,
		Version:      idA.Version,
		Port:         8787,
		PublicKey:    idA.PublicKey,
		Capabilities: idA.Capabilities,
	})

	// Verify Machine A sees Machine B with capabilities
	bSeenByA, ok := peerMgrA.Get(idB.DeviceID)
	if !ok {
		t.Fatalf("Machine A failed to discover Machine B")
	}
	if bSeenByA.DeviceName != "Laptop-B" {
		t.Errorf("expected DeviceName Laptop-B, got %s", bSeenByA.DeviceName)
	}
	if !bSeenByA.HasCapability("terminal") || !bSeenByA.HasCapability("command") {
		t.Errorf("expected Machine B to advertise terminal and command capabilities")
	}

	// Verify Machine B sees Machine A
	aSeenByB, ok := peerMgrB.Get(idA.DeviceID)
	if !ok {
		t.Fatalf("Machine B failed to discover Machine A")
	}
	if aSeenByB.DeviceName != "Workstation-A" {
		t.Errorf("expected DeviceName Workstation-A, got %s", aSeenByB.DeviceName)
	}
}
