package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"synx/internal/integrity"
)

type FileItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"` // relative path
	Size     int64     `json:"size"`
	IsDir    bool      `json:"dir"`
	Modified time.Time `json:"modified"`
	SHA256   string    `json:"sha256,omitempty"`
}

type ScanOptions struct {
	Recursive    bool
	ComputeHash  bool
	MaxDepth     int
	ExcludeParts bool
}

// ScanDirectory lists files and folders inside rootDir (or subpath relPath)
func ScanDirectory(rootDir, relPath string, opts ScanOptions) ([]FileItem, error) {
	safeRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	targetDir := safeRoot
	if relPath != "" {
		resolved, err := ResolveSafePath(safeRoot, relPath)
		if err != nil {
			return nil, err
		}
		targetDir = resolved
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	var results []FileItem
	for _, entry := range entries {
		name := entry.Name()
		if opts.ExcludeParts && strings.HasSuffix(name, ".synx.part") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(targetDir, name)
		rel, _ := filepath.Rel(safeRoot, fullPath)
		cleanRel := filepath.ToSlash(rel)

		item := FileItem{
			Name:     name,
			Path:     cleanRel,
			Size:     info.Size(),
			IsDir:    entry.IsDir(),
			Modified: info.ModTime(),
		}

		if !entry.IsDir() && opts.ComputeHash {
			item.SHA256, _ = integrity.FileSHA256(fullPath)
		}

		results = append(results, item)

		if entry.IsDir() && opts.Recursive {
			subItems, err := ScanDirectory(safeRoot, cleanRel, opts)
			if err == nil {
				results = append(results, subItems...)
			}
		}
	}

	return results, nil
}
