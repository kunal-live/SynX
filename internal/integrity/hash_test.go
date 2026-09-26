package integrity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSHA256(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")
	content := []byte("SynX Integrity Test Content")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash, err := FileSHA256(filePath)
	if err != nil {
		t.Fatalf("FileSHA256 error: %v", err)
	}

	expected := BytesSHA256(content)
	if hash != expected {
		t.Errorf("hash mismatch: expected %s, got %s", expected, hash)
	}

	ok, actual, err := VerifyFile(filePath, expected)
	if err != nil || !ok || actual != expected {
		t.Errorf("VerifyFile failed: ok=%v, actual=%s, err=%v", ok, actual, err)
	}

	okBad, _, _ := VerifyFile(filePath, "invalidhash")
	if okBad {
		t.Errorf("expected mismatch with wrong hash")
	}
}
