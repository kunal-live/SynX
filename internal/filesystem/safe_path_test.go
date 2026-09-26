package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathTraversalProtection(t *testing.T) {
	tempDir := t.TempDir()

	// Legitimate path
	safe, err := ResolveSafePath(tempDir, "subfolder/hello.txt")
	if err != nil {
		t.Fatalf("expected legitimate path to succeed, got %v", err)
	}
	expected := filepath.Join(tempDir, "subfolder", "hello.txt")
	if safe != expected {
		t.Fatalf("expected %s, got %s", expected, safe)
	}

	// Traversal attacks
	malicious := []string{
		"../secret.txt",
		"..\\secret.txt",
		"sub/../../secret.txt",
		"/etc/passwd",
		"C:\\Windows\\System32\\calc.exe",
		"..",
		".",
	}

	for _, mal := range malicious {
		_, err := ResolveSafePath(tempDir, mal)
		if err == nil {
			t.Errorf("expected attack %q to be blocked, but it was accepted", mal)
		}
	}
}

func TestPartFileCommit(t *testing.T) {
	tempDir := t.TempDir()
	finalPath := filepath.Join(tempDir, "sample.txt")
	partPath := PartFilePath(finalPath)

	if err := os.WriteFile(partPath, []byte("sample content"), 0644); err != nil {
		t.Fatalf("failed to write part file: %v", err)
	}

	if err := CommitPartFile(finalPath); err != nil {
		t.Fatalf("CommitPartFile failed: %v", err)
	}

	if _, err := os.Stat(partPath); !os.IsNotExist(err) {
		t.Errorf("expected part file to be removed after commit")
	}

	data, err := os.ReadFile(finalPath)
	if err != nil || string(data) != "sample content" {
		t.Errorf("committed content mismatch: got %s, err %v", string(data), err)
	}
}
