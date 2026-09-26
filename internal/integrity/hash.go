package integrity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FileSHA256 streams the entire file and computes the SHA-256 hex digest without loading the whole file into RAM
func FileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to open file for hashing: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	buf := make([]byte, 1024*1024) // 1MB streaming buffer
	if _, err := io.CopyBuffer(hasher, f, buf); err != nil {
		return "", fmt.Errorf("error during streaming file hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// BytesSHA256 computes the SHA-256 hex digest of a byte slice
func BytesSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// VerifyFile compares the calculated SHA-256 of the file against the expected hash
func VerifyFile(filePath, expectedHash string) (bool, string, error) {
	actual, err := FileSHA256(filePath)
	if err != nil {
		return false, "", err
	}
	return actual == expectedHash, actual, nil
}
