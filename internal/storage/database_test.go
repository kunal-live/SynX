package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_synx.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// 1. Device upsert & trust
	dev := DeviceRecord{
		ID:        "dev_123",
		Name:      "MacBook Pro",
		Platform:  "darwin",
		Version:   "1.0.0",
		PublicKey: "abcdef",
		Trusted:   false,
		CreatedAt: time.Now().Unix(),
		LastSeen:  time.Now().Unix(),
	}
	if err := db.UpsertDevice(dev); err != nil {
		t.Fatalf("UpsertDevice failed: %v", err)
	}

	devices, err := db.ListDevices()
	if err != nil || len(devices) != 1 {
		t.Fatalf("ListDevices failed: len=%d, err=%v", len(devices), err)
	}
	if devices[0].Name != "MacBook Pro" || devices[0].Trusted {
		t.Errorf("device attributes mismatch")
	}

	if err := db.SetDeviceTrust("dev_123", true); err != nil {
		t.Fatalf("SetDeviceTrust failed: %v", err)
	}

	devices, _ = db.ListDevices()
	if !devices[0].Trusted {
		t.Errorf("expected device to be trusted")
	}

	// 2. Transfer persistence & recovery
	tx := TransferRecord{
		ID:               "tx_999",
		PeerID:           "dev_123",
		Direction:        "send",
		FileName:         "test.zip",
		Size:             1000,
		BytesTransferred: 500,
		ChunkSize:        100,
		Status:           "transferring",
		CreatedAt:        time.Now().Unix(),
	}
	if err := db.SaveTransfer(tx); err != nil {
		t.Fatalf("SaveTransfer failed: %v", err)
	}

	// Test crash recovery (Section 50)
	if err := db.MarkInterruptedOnStartup(); err != nil {
		t.Fatalf("MarkInterruptedOnStartup failed: %v", err)
	}

	loaded, err := db.GetTransfer("tx_999")
	if err != nil || loaded.Status != "interrupted" {
		t.Errorf("expected transfer to be recovered as 'interrupted', got status=%s, err=%v", loaded.Status, err)
	}

	// 3. Settings KV
	if err := db.SetSetting("test_key", "hello_synx"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	val, err := db.GetSetting("test_key")
	if err != nil || val != "hello_synx" {
		t.Errorf("GetSetting mismatch: val=%s, err=%v", val, err)
	}
}
