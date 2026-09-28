package peers

import (
	"path/filepath"
	"testing"
	"time"

	"synx/internal/events"
	"synx/internal/observability"
	"synx/internal/storage"
)

func TestPeersManagerLifecycle(t *testing.T) {
	observability.DefaultMetrics().Reset()

	tempDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tempDir, "peers_test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	bus := events.NewBus()
	mgr := NewManager(db, bus)
	defer mgr.Close()

	// 1. Add Peer
	mgr.AddOrUpdate(Peer{
		DeviceID:     "dev_alpha",
		DeviceName:   "Workstation Alpha",
		Address:      "192.168.1.100:8787",
		Platform:     "linux",
		Capabilities: []string{"terminal", "command", "files"},
		Status:       StatusReachable,
	})

	if observability.DefaultMetrics().GetPeerCount() != 1 {
		t.Errorf("expected 1 active peer metric, got %d", observability.DefaultMetrics().GetPeerCount())
	}

	p, ok := mgr.Get("dev_alpha")
	if !ok || p.ID != "dev_alpha" {
		t.Fatalf("failed to get peer dev_alpha: %+v", p)
	}
	if !p.HasCapability("terminal") || p.HasCapability("clipboard") {
		t.Errorf("capability check failed on peer")
	}
	if !p.IsOnline(10 * time.Second) {
		t.Errorf("expected peer to be online")
	}

	// 2. Add second peer
	mgr.AddOrUpdate(Peer{
		ID:       "dev_beta",
		Name:     "Laptop Beta",
		Address:  "192.168.1.101:8787",
		Platform: "darwin",
		Status:   StatusReachable,
	})

	if mgr.ActiveCount() != 2 || observability.DefaultMetrics().GetPeerCount() != 2 {
		t.Errorf("expected 2 active peers, got active=%d metric=%d", mgr.ActiveCount(), observability.DefaultMetrics().GetPeerCount())
	}

	// 3. Mark trusted
	mgr.SetTrusted("dev_alpha", true)
	pTrusted, _ := mgr.Get("dev_alpha")
	if !pTrusted.Trusted || pTrusted.Status != StatusTrusted {
		t.Errorf("expected dev_alpha to be trusted, got %+v", pTrusted)
	}

	// 4. Remove peer
	mgr.Remove("dev_beta")
	if _, found := mgr.Get("dev_beta"); found {
		t.Errorf("expected dev_beta to be removed")
	}
	if mgr.ActiveCount() != 1 || observability.DefaultMetrics().GetPeerCount() != 1 {
		t.Errorf("expected 1 active peer after removal, got active=%d metric=%d", mgr.ActiveCount(), observability.DefaultMetrics().GetPeerCount())
	}
}
