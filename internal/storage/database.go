package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

func Open(path string) (*DB, error) {
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize for single-desktop client concurrency
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &DB{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return s, nil
}

func (s *DB) Close() error {
	return s.db.Close()
}

func (s *DB) migrate() error {
	schema := `
	PRAGMA journal_mode = WAL;
	PRAGMA synchronous = NORMAL;

	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		platform TEXT,
		version TEXT,
		public_key TEXT,
		trusted INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		last_seen INTEGER
	);

	CREATE TABLE IF NOT EXISTS transfers (
		id TEXT PRIMARY KEY,
		peer_id TEXT NOT NULL,
		direction TEXT NOT NULL,
		file_name TEXT NOT NULL,
		source_path TEXT,
		destination_path TEXT,
		size INTEGER NOT NULL,
		bytes_transferred INTEGER NOT NULL DEFAULT 0,
		sha256 TEXT,
		chunk_size INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		started_at INTEGER,
		completed_at INTEGER
	);

	CREATE TABLE IF NOT EXISTS chunks (
		transfer_id TEXT NOT NULL,
		chunk_index INTEGER NOT NULL,
		offset INTEGER NOT NULL,
		size INTEGER NOT NULL,
		sha256 TEXT,
		status TEXT NOT NULL,
		PRIMARY KEY (transfer_id, chunk_index)
	);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transfer_history (
		id TEXT PRIMARY KEY,
		transfer_id TEXT,
		peer_id TEXT,
		file_name TEXT,
		size INTEGER,
		direction TEXT,
		status TEXT,
		completed_at INTEGER
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Device methods
type DeviceRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	Version   string `json:"version"`
	PublicKey string `json:"public_key"`
	Trusted   bool   `json:"trusted"`
	CreatedAt int64  `json:"created_at"`
	LastSeen  int64  `json:"last_seen"`
}

func (s *DB) UpsertDevice(d DeviceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO devices (id, name, platform, version, public_key, trusted, created_at, last_seen)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		platform = excluded.platform,
		version = excluded.version,
		public_key = excluded.public_key,
		last_seen = excluded.last_seen;
	`
	trustedInt := 0
	if d.Trusted {
		trustedInt = 1
	}
	_, err := s.db.Exec(query, d.ID, d.Name, d.Platform, d.Version, d.PublicKey, trustedInt, d.CreatedAt, d.LastSeen)
	return err
}

func (s *DB) SetDeviceTrust(id string, trusted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val := 0
	if trusted {
		val = 1
	}
	_, err := s.db.Exec("UPDATE devices SET trusted = ? WHERE id = ?", val, id)
	return err
}

func (s *DB) ListDevices() ([]DeviceRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, name, platform, version, public_key, trusted, created_at, last_seen FROM devices ORDER BY last_seen DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DeviceRecord
	for rows.Next() {
		var d DeviceRecord
		var tr int
		if err := rows.Scan(&d.ID, &d.Name, &d.Platform, &d.Version, &d.PublicKey, &tr, &d.CreatedAt, &d.LastSeen); err != nil {
			return nil, err
		}
		d.Trusted = tr == 1
		out = append(out, d)
	}
	return out, nil
}

// Transfer records
type TransferRecord struct {
	ID               string `json:"id"`
	PeerID           string `json:"peer_id"`
	Direction        string `json:"direction"`
	FileName         string `json:"file_name"`
	SourcePath       string `json:"source_path"`
	DestinationPath  string `json:"destination_path"`
	Size             int64  `json:"size"`
	BytesTransferred int64  `json:"bytes_transferred"`
	SHA256           string `json:"sha256"`
	ChunkSize        int64  `json:"chunk_size"`
	Status           string `json:"status"`
	CreatedAt        int64  `json:"created_at"`
	StartedAt        int64  `json:"started_at"`
	CompletedAt      int64  `json:"completed_at"`
}

func (s *DB) SaveTransfer(t TransferRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO transfers (id, peer_id, direction, file_name, source_path, destination_path, size, bytes_transferred, sha256, chunk_size, status, created_at, started_at, completed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		bytes_transferred = excluded.bytes_transferred,
		status = excluded.status,
		completed_at = excluded.completed_at;
	`
	_, err := s.db.Exec(query, t.ID, t.PeerID, t.Direction, t.FileName, t.SourcePath, t.DestinationPath, t.Size, t.BytesTransferred, t.SHA256, t.ChunkSize, t.Status, t.CreatedAt, t.StartedAt, t.CompletedAt)
	return err
}

func (s *DB) UpdateTransferProgress(id string, bytesTransferred int64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var completedAt *int64
	if status == "completed" || status == "failed" || status == "cancelled" {
		now := time.Now().Unix()
		completedAt = &now
	}

	if completedAt != nil {
		_, err := s.db.Exec("UPDATE transfers SET bytes_transferred = ?, status = ?, completed_at = ? WHERE id = ?", bytesTransferred, status, *completedAt, id)
		return err
	}
	_, err := s.db.Exec("UPDATE transfers SET bytes_transferred = ?, status = ? WHERE id = ?", bytesTransferred, status, id)
	return err
}

func (s *DB) GetTransfer(id string) (*TransferRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var t TransferRecord
	var completedAt sql.NullInt64
	var startedAt sql.NullInt64

	row := s.db.QueryRow("SELECT id, peer_id, direction, file_name, source_path, destination_path, size, bytes_transferred, sha256, chunk_size, status, created_at, started_at, completed_at FROM transfers WHERE id = ?", id)
	err := row.Scan(&t.ID, &t.PeerID, &t.Direction, &t.FileName, &t.SourcePath, &t.DestinationPath, &t.Size, &t.BytesTransferred, &t.SHA256, &t.ChunkSize, &t.Status, &t.CreatedAt, &startedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t.StartedAt = startedAt.Int64
	}
	if completedAt.Valid {
		t.CompletedAt = completedAt.Int64
	}
	return &t, nil
}

func (s *DB) ListTransfers() ([]TransferRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT id, peer_id, direction, file_name, source_path, destination_path, size, bytes_transferred, sha256, chunk_size, status, created_at, started_at, completed_at FROM transfers ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TransferRecord
	for rows.Next() {
		var t TransferRecord
		var startedAt sql.NullInt64
		var completedAt sql.NullInt64
		if err := rows.Scan(&t.ID, &t.PeerID, &t.Direction, &t.FileName, &t.SourcePath, &t.DestinationPath, &t.Size, &t.BytesTransferred, &t.SHA256, &t.ChunkSize, &t.Status, &t.CreatedAt, &startedAt, &completedAt); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t.StartedAt = startedAt.Int64
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.Int64
		}
		out = append(out, t)
	}
	return out, nil
}

// MarkInterruptedOnStartup recovers crashed/uncleanly terminated transfers (Section 50)
func (s *DB) MarkInterruptedOnStartup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("UPDATE transfers SET status = 'interrupted' WHERE status IN ('transferring', 'queued', 'verifying', 'committing')")
	return err
}

// Chunks
type ChunkRecord struct {
	TransferID string `json:"transfer_id"`
	ChunkIndex int    `json:"chunk_index"`
	Offset     int64  `json:"offset"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	Status     string `json:"status"`
}

func (s *DB) SaveChunk(c ChunkRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO chunks (transfer_id, chunk_index, offset, size, sha256, status)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(transfer_id, chunk_index) DO UPDATE SET
		sha256 = excluded.sha256,
		status = excluded.status;
	`
	_, err := s.db.Exec(query, c.TransferID, c.ChunkIndex, c.Offset, c.Size, c.SHA256, c.Status)
	return err
}

func (s *DB) GetCompletedChunkIndices(transferID string) ([]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT chunk_index FROM chunks WHERE transfer_id = ? AND status = 'completed' ORDER BY chunk_index ASC", transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indices []int
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err == nil {
			indices = append(indices, idx)
		}
	}
	return indices, nil
}

// History
type HistoryRecord struct {
	ID          string `json:"id"`
	TransferID  string `json:"transfer_id"`
	PeerID      string `json:"peer_id"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	Direction   string `json:"direction"`
	Status      string `json:"status"`
	CompletedAt int64  `json:"completed_at"`
}

func (s *DB) AddHistory(h HistoryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO transfer_history (id, transfer_id, peer_id, file_name, size, direction, status, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, h.ID, h.TransferID, h.PeerID, h.FileName, h.Size, h.Direction, h.Status, h.CompletedAt)
	return err
}

func (s *DB) ListHistory(limit int) ([]HistoryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query("SELECT id, transfer_id, peer_id, file_name, size, direction, status, completed_at FROM transfer_history ORDER BY completed_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []HistoryRecord
	for rows.Next() {
		var h HistoryRecord
		if err := rows.Scan(&h.ID, &h.TransferID, &h.PeerID, &h.FileName, &h.Size, &h.Direction, &h.Status, &h.CompletedAt); err == nil {
			list = append(list, h)
		}
	}
	return list, nil
}

// Settings KV
func (s *DB) SetSetting(key, val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value", key, val)
	return err
}

func (s *DB) GetSetting(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	return val, err
}
