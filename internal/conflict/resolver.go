package conflict

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type ResolutionPolicy string

const (
	PolicyKeepLocal  ResolutionPolicy = "keep_local"
	PolicyKeepRemote ResolutionPolicy = "keep_remote"
	PolicyKeepBoth   ResolutionPolicy = "keep_both"
)

type Conflict struct {
	Path       string `json:"path"`
	LocalHash  string `json:"local_hash"`
	RemoteHash string `json:"remote_hash"`
	LocalMod   int64  `json:"local_mod"`
	RemoteMod  int64  `json:"remote_mod"`
}

// GenerateConflictPath creates a unique conflict filename e.g. "report (Kunal-PC conflict 2026-09-26).txt"
func GenerateConflictPath(originalPath, deviceName string) string {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(originalPath, ext)
	dateStr := time.Now().Format("2006-01-02_15-04")
	return fmt.Sprintf("%s (%s conflict %s)%s", base, deviceName, dateStr, ext)
}

func DetectConflict(localHash, remoteHash string, localMod, remoteMod int64) bool {
	if localHash == remoteHash {
		return false
	}
	// If both files were modified independently
	return localMod != 0 && remoteMod != 0 && localHash != "" && remoteHash != ""
}
