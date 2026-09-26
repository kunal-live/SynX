package chunk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunkReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.bin")
	partPath := filepath.Join(tempDir, "dest.bin.synx.part")

	data := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if err := os.WriteFile(srcPath, data, 0644); err != nil {
		t.Fatalf("failed to create source: %v", err)
	}

	chunkSize := int64(16)
	ch0, err := ReadChunk(srcPath, 0, chunkSize)
	if err != nil {
		t.Fatalf("ReadChunk 0 failed: %v", err)
	}
	if ch0.Size != 16 {
		t.Errorf("expected chunk size 16, got %d", ch0.Size)
	}

	if err := WriteChunk(partPath, ch0, ch0.SHA256); err != nil {
		t.Fatalf("WriteChunk 0 failed: %v", err)
	}

	ch1, err := ReadChunk(srcPath, 1, chunkSize)
	if err != nil {
		t.Fatalf("ReadChunk 1 failed: %v", err)
	}
	if err := WriteChunk(partPath, ch1, ch1.SHA256); err != nil {
		t.Fatalf("WriteChunk 1 failed: %v", err)
	}

	// Verify part file content for first 32 bytes
	partData, err := os.ReadFile(partPath)
	if err != nil || len(partData) != 32 {
		t.Fatalf("expected 32 bytes in part file, got %d", len(partData))
	}
	if string(partData) != string(data[:32]) {
		t.Errorf("part data mismatch")
	}

	// Test Manifest
	manifest := NewManifest("tx_1", "dest.bin", int64(len(data)), chunkSize)
	manifest.MarkCompleted(0, 16)
	manifest.MarkCompleted(1, 16)

	if !manifest.IsCompleted(0) || !manifest.IsCompleted(1) || manifest.IsCompleted(2) {
		t.Errorf("manifest completion state incorrect")
	}
	if manifest.CompletedBytes() != 32 {
		t.Errorf("expected 32 completed bytes, got %d", manifest.CompletedBytes())
	}
}
