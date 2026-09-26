package sync

import (
	"synx/internal/filesystem"
)

type FileMetadata struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	SHA256   string `json:"sha256"`
}

type FolderManifest struct {
	Version  int            `json:"version"`
	DeviceID string         `json:"device_id"`
	Files    []FileMetadata `json:"files"`
}

type SyncActionType string

const (
	ActionUpload   SyncActionType = "upload"
	ActionDownload SyncActionType = "download"
	ActionUnchanged SyncActionType = "unchanged"
	ActionConflict SyncActionType = "conflict"
)

type SyncPlanItem struct {
	Path   string         `json:"path"`
	Action SyncActionType `json:"action"`
}

func GenerateManifest(rootDir string, deviceID string) (*FolderManifest, error) {
	items, err := filesystem.ScanDirectory(rootDir, "", filesystem.ScanOptions{
		Recursive:    true,
		ComputeHash:  true,
		ExcludeParts: true,
	})
	if err != nil {
		return nil, err
	}

	files := make([]FileMetadata, 0, len(items))
	for _, it := range items {
		if !it.IsDir {
			files = append(files, FileMetadata{
				Path:     it.Path,
				Size:     it.Size,
				Modified: it.Modified.Unix(),
				SHA256:   it.SHA256,
			})
		}
	}

	return &FolderManifest{
		Version:  1,
		DeviceID: deviceID,
		Files:    files,
	}, nil
}

func DiffManifests(local, remote *FolderManifest) []SyncPlanItem {
	localMap := make(map[string]FileMetadata)
	remoteMap := make(map[string]FileMetadata)

	for _, f := range local.Files {
		localMap[f.Path] = f
	}
	for _, f := range remote.Files {
		remoteMap[f.Path] = f
	}

	var plan []SyncPlanItem

	for path, l := range localMap {
		r, exists := remoteMap[path]
		if !exists {
			plan = append(plan, SyncPlanItem{Path: path, Action: ActionUpload})
		} else if l.SHA256 == r.SHA256 {
			plan = append(plan, SyncPlanItem{Path: path, Action: ActionUnchanged})
		} else {
			plan = append(plan, SyncPlanItem{Path: path, Action: ActionConflict})
		}
	}

	for path := range remoteMap {
		if _, exists := localMap[path]; !exists {
			plan = append(plan, SyncPlanItem{Path: path, Action: ActionDownload})
		}
	}

	return plan
}
