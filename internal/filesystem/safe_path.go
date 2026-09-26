package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathEscape       = errors.New("path traversal detected: target path escapes allowed directory")
	ErrEmptyPath        = errors.New("empty path provided")
	ErrInvalidCharacters = errors.New("path contains illegal path characters")
)

// SanitizeRelPath cleans and normalizes a relative path, rejecting any attempt at traversal
func SanitizeRelPath(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", ErrEmptyPath
	}

	// Reject absolute paths, UNC paths, and drive traversal
	if strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "\\") || strings.Contains(rel, ":") || strings.Contains(rel, "\x00") {
		return "", ErrPathEscape
	}

	// Normalize separators to forward slash
	rel = strings.ReplaceAll(rel, "\\", "/")

	// Clean path
	cleaned := filepath.ToSlash(filepath.Clean(rel))

	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/..") {
		return "", ErrPathEscape
	}

	return cleaned, nil
}

// ResolveSafePath safely resolves rel inside rootDir and guarantees it cannot escape rootDir
func ResolveSafePath(rootDir, rel string) (string, error) {
	cleanRel, err := SanitizeRelPath(rel)
	if err != nil {
		return "", err
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("invalid root directory: %w", err)
	}

	target := filepath.Join(absRoot, filepath.FromSlash(cleanRel))
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("unable to resolve absolute path: %w", err)
	}

	// Check containment: target must be inside absRoot
	relToRoot, err := filepath.Rel(absRoot, absTarget)
	if err != nil || strings.HasPrefix(relToRoot, "..") || relToRoot == ".." {
		return "", ErrPathEscape
	}

	return absTarget, nil
}

// PartFilePath returns the temporary .synx.part path for an incoming transfer (Section 21)
func PartFilePath(finalPath string) string {
	return finalPath + ".synx.part"
}

// CommitPartFile atomically renames the .synx.part file to the final destination
func CommitPartFile(finalPath string) error {
	part := PartFilePath(finalPath)
	if _, err := os.Stat(part); err != nil {
		return fmt.Errorf("temporary part file does not exist: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(part, finalPath); err != nil {
		// Windows fallback: if destination exists, remove first then rename
		_ = os.Remove(finalPath)
		return os.Rename(part, finalPath)
	}
	return nil
}
