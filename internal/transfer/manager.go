package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"synx/internal/chunk"
	"synx/internal/events"
	"synx/internal/integrity"
	"synx/internal/observability"
	"synx/internal/peers"
	"synx/internal/protocol"
	"synx/internal/security"
	"synx/internal/storage"
	"synx/internal/transport"
)

type Direction string

const (
	DirectionSend    Direction = "send"
	DirectionReceive Direction = "receive"
)

type Status string

const (
	StatusCreated      Status = "created"
	StatusNegotiating  Status = "negotiating"
	StatusQueued       Status = "queued"
	StatusTransferring Status = "transferring"
	StatusVerifying    Status = "verifying"
	StatusCommitting   Status = "committing"
	StatusCompleted    Status = "completed"
	StatusFailed       Status = "failed"
	StatusCancelled    Status = "cancelled"
	StatusPaused       Status = "paused"
	StatusInterrupted  Status = "interrupted"
)

type Transfer struct {
	ID               string     `json:"id"`
	PeerID           string     `json:"peer_id"`
	PeerName         string     `json:"peer_name"`
	Direction        Direction  `json:"direction"`
	SourcePath       string     `json:"source_path"`
	DestinationPath  string     `json:"destination_path"`
	FileName         string     `json:"name"`
	Size             int64      `json:"size"`
	BytesTransferred int64      `json:"done"`
	Progress         float64    `json:"progress"`
	Speed            string     `json:"speed"`
	ETASeconds       int        `json:"eta"`
	ChunkSize        int64      `json:"chunk_size"`
	TotalChunks      int        `json:"total_chunks"`
	CompletedChunks  int        `json:"completed_chunks"`
	Status           Status     `json:"status"`
	SHA256           string     `json:"sha256"`
	CreatedAt        time.Time  `json:"created_at"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`

	cancelFunc context.CancelFunc `json:"-"`
	pauseCh    chan struct{}      `json:"-"`
}

type SpeedSample struct {
	Timestamp time.Time
	Bytes     int64
}

type Manager struct {
	mu           sync.RWMutex
	transfers    map[string]*Transfer
	queue        chan *Transfer
	db           *storage.DB
	bus          *events.Bus
	peerMgr      *peers.Manager
	transport    transport.Transport
	maxWorkers   int
	speedSamples map[string][]SpeedSample
	stopCh       chan struct{}
}

func NewManager(db *storage.DB, bus *events.Bus, peerMgr *peers.Manager, tr transport.Transport, maxWorkers int) *Manager {
	if maxWorkers <= 0 {
		maxWorkers = 2
	}
	m := &Manager{
		transfers:    make(map[string]*Transfer),
		queue:        make(chan *Transfer, 100),
		db:           db,
		bus:          bus,
		peerMgr:      peerMgr,
		transport:    tr,
		maxWorkers:   maxWorkers,
		speedSamples: make(map[string][]SpeedSample),
		stopCh:       make(chan struct{}),
	}
	m.restoreTransfers()
	for i := 0; i < maxWorkers; i++ {
		go m.worker(i)
	}
	return m
}

func (m *Manager) Close() {
	close(m.stopCh)
}

func (m *Manager) restoreTransfers() {
	if m.db == nil {
		return
	}
	_ = m.db.MarkInterruptedOnStartup()
	records, err := m.db.ListTransfers()
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range records {
		var comp *time.Time
		if r.CompletedAt != 0 {
			t := time.Unix(r.CompletedAt, 0)
			comp = &t
		}
		prog := float64(0)
		if r.Size > 0 {
			prog = (float64(r.BytesTransferred) / float64(r.Size)) * 100
		}
		m.transfers[r.ID] = &Transfer{
			ID:               r.ID,
			PeerID:           r.PeerID,
			Direction:        Direction(r.Direction),
			FileName:         r.FileName,
			SourcePath:       r.SourcePath,
			DestinationPath:  r.DestinationPath,
			Size:             r.Size,
			BytesTransferred: r.BytesTransferred,
			Progress:         prog,
			ChunkSize:        r.ChunkSize,
			Status:           Status(r.Status),
			SHA256:           r.SHA256,
			CreatedAt:        time.Unix(r.CreatedAt, 0),
			StartedAt:        time.Unix(r.StartedAt, 0),
			CompletedAt:      comp,
		}
	}
}

func (m *Manager) CreateTransfer(peerID, sourcePath, destRel string) (*Transfer, error) {
	fi, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("source file error: %w", err)
	}

	peer, ok := m.peerMgr.Get(peerID)
	if !ok {
		return nil, fmt.Errorf("peer %s not found", peerID)
	}

	fileName := filepath.Base(sourcePath)
	if destRel == "" {
		destRel = fileName
	}

	chunkSize := chunk.DefaultChunkSize
	totalChunks := int((fi.Size() + chunkSize - 1) / chunkSize)
	if totalChunks == 0 {
		totalChunks = 1
	}

	txID := "tx_" + security.GenerateToken(8)
	t := &Transfer{
		ID:              txID,
		PeerID:          peer.ID,
		PeerName:        peer.Name,
		Direction:       DirectionSend,
		SourcePath:      sourcePath,
		DestinationPath: destRel,
		FileName:        fileName,
		Size:            fi.Size(),
		ChunkSize:       chunkSize,
		TotalChunks:     totalChunks,
		Status:          StatusQueued,
		CreatedAt:       time.Now(),
		pauseCh:         make(chan struct{}),
	}

	m.mu.Lock()
	m.transfers[txID] = t
	m.mu.Unlock()

	if m.db != nil {
		_ = m.db.SaveTransfer(storage.TransferRecord{
			ID:               t.ID,
			PeerID:           t.PeerID,
			Direction:        string(t.Direction),
			FileName:         t.FileName,
			SourcePath:       t.SourcePath,
			DestinationPath:  t.DestinationPath,
			Size:             t.Size,
			BytesTransferred: 0,
			ChunkSize:        t.ChunkSize,
			Status:           string(t.Status),
			CreatedAt:        t.CreatedAt.Unix(),
		})
	}

	m.bus.Publish("transfer.created", map[string]any{
		"id":        t.ID,
		"name":      t.FileName,
		"size":      t.Size,
		"direction": t.Direction,
		"peer":      t.PeerName,
	})

	m.queue <- t
	return t, nil
}

func (m *Manager) worker(id int) {
	for {
		select {
		case <-m.stopCh:
			return
		case t := <-m.queue:
			m.executeTransfer(t)
		}
	}
}

func (m *Manager) executeTransfer(t *Transfer) {
	ctx, cancel := context.WithCancel(context.Background())
	t.cancelFunc = cancel
	t.StartedAt = time.Now()
	t.Status = StatusTransferring

	m.bus.Publish("transfer.started", map[string]any{"id": t.ID, "name": t.FileName})

	peer, ok := m.peerMgr.Get(t.PeerID)
	if !ok {
		m.failTransfer(t, "peer disconnected")
		return
	}

	// 1. Calculate streaming SHA-256
	hash, err := integrity.FileSHA256(t.SourcePath)
	if err != nil {
		m.failTransfer(t, "hash calculation failed: "+err.Error())
		return
	}
	t.SHA256 = hash

	// 2. Negotiate with remote peer
	nego, err := m.transport.Negotiate(ctx, *peer, protocol.TransferNegotiationRequest{
		SourcePath:  t.SourcePath,
		Destination: t.DestinationPath,
		FileName:    t.FileName,
		Size:        t.Size,
		SHA256:      t.SHA256,
		ChunkSize:   t.ChunkSize,
	})
	if err != nil {
		m.failTransfer(t, "negotiation failed: "+err.Error())
		return
	}

	remoteID := nego.TransferID
	if remoteID == "" {
		remoteID = t.ID
	}

	completedMap := make(map[int]bool)
	for _, idx := range nego.Completed {
		completedMap[idx] = true
	}

	// 3. Send Chunks
	for i := 0; i < t.TotalChunks; i++ {
		select {
		case <-ctx.Done():
			m.cancelTransfer(t)
			return
		case <-m.stopCh:
			return
		default:
		}

		if completedMap[i] {
			continue
		}

		ch, err := chunk.ReadChunk(t.SourcePath, i, t.ChunkSize)
		if err != nil {
			m.failTransfer(t, fmt.Sprintf("chunk read error: %v", err))
			return
		}

		if err := m.transport.SendChunk(ctx, *peer, remoteID, ch); err != nil {
			m.failTransfer(t, fmt.Sprintf("chunk upload error: %v", err))
			return
		}

		m.mu.Lock()
		t.BytesTransferred += ch.Size
		t.CompletedChunks++
		if t.Size > 0 {
			t.Progress = (float64(t.BytesTransferred) / float64(t.Size)) * 100
		}
		m.recordSpeedSample(t.ID, ch.Size)
		t.Speed, t.ETASeconds = m.calculateSpeedAndETA(t.ID, t.Size-t.BytesTransferred)
		m.mu.Unlock()

		if m.db != nil {
			_ = m.db.UpdateTransferProgress(t.ID, t.BytesTransferred, string(StatusTransferring))
		}

		m.bus.Publish("transfer.progress", map[string]any{
			"id":       t.ID,
			"bytes":    t.BytesTransferred,
			"total":    t.Size,
			"progress": t.Progress,
			"speed":    t.Speed,
			"eta":      t.ETASeconds,
		})
	}

	// 4. Commit transfer
	t.Status = StatusVerifying
	commitRes, err := m.transport.Commit(ctx, *peer, remoteID, t.SHA256)
	if err != nil || (commitRes != nil && !commitRes.Verified) {
		m.failTransfer(t, "commit or integrity verification failed")
		return
	}

	now := time.Now()
	t.CompletedAt = &now
	t.Status = StatusCompleted
	t.Progress = 100.0

	if m.db != nil {
		_ = m.db.UpdateTransferProgress(t.ID, t.Size, string(StatusCompleted))
		_ = m.db.AddHistory(storage.HistoryRecord{
			ID:          "hist_" + security.GenerateToken(8),
			TransferID:  t.ID,
			PeerID:      t.PeerID,
			FileName:    t.FileName,
			Size:        t.Size,
			Direction:   string(t.Direction),
			Status:      string(StatusCompleted),
			CompletedAt: now.Unix(),
		})
	}

	observability.Info("Transfer %s (%s) completed successfully", t.ID, t.FileName)
	m.bus.Publish("transfer.completed", map[string]any{"id": t.ID, "name": t.FileName})
}

func (m *Manager) failTransfer(t *Transfer, reason string) {
	m.mu.Lock()
	t.Status = StatusFailed
	now := time.Now()
	t.CompletedAt = &now
	m.mu.Unlock()

	if m.db != nil {
		_ = m.db.UpdateTransferProgress(t.ID, t.BytesTransferred, string(StatusFailed))
		_ = m.db.AddHistory(storage.HistoryRecord{
			ID:          "hist_" + security.GenerateToken(8),
			TransferID:  t.ID,
			PeerID:      t.PeerID,
			FileName:    t.FileName,
			Size:        t.Size,
			Direction:   string(t.Direction),
			Status:      string(StatusFailed),
			CompletedAt: now.Unix(),
		})
	}

	observability.Error("Transfer %s failed: %s", t.ID, reason)
	m.bus.Publish("transfer.failed", map[string]any{"id": t.ID, "reason": reason})
}

func (m *Manager) cancelTransfer(t *Transfer) {
	m.mu.Lock()
	t.Status = StatusCancelled
	m.mu.Unlock()
	if m.db != nil {
		_ = m.db.UpdateTransferProgress(t.ID, t.BytesTransferred, string(StatusCancelled))
	}
	m.bus.Publish("transfer.cancelled", map[string]any{"id": t.ID})
}

func (m *Manager) Cancel(id string) error {
	m.mu.RLock()
	t, ok := m.transfers[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("transfer %s not found", id)
	}
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
	return nil
}

func (m *Manager) List() []Transfer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Transfer, 0, len(m.transfers))
	for _, t := range m.transfers {
		out = append(out, *t)
	}
	return out
}

func (m *Manager) Get(id string) (*Transfer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.transfers[id]
	if !ok {
		return nil, false
	}
	cpy := *t
	return &cpy, true
}

func (m *Manager) recordSpeedSample(id string, bytes int64) {
	now := time.Now()
	samples := m.speedSamples[id]
	samples = append(samples, SpeedSample{Timestamp: now, Bytes: bytes})

	// Keep last 5 seconds of samples
	cutoff := now.Add(-5 * time.Second)
	idx := 0
	for idx < len(samples) && samples[idx].Timestamp.Before(cutoff) {
		idx++
	}
	m.speedSamples[id] = samples[idx:]
}

func (m *Manager) calculateSpeedAndETA(id string, remainingBytes int64) (string, int) {
	samples := m.speedSamples[id]
	if len(samples) < 2 {
		return "—", 0
	}

	var totalBytes int64
	duration := samples[len(samples)-1].Timestamp.Sub(samples[0].Timestamp).Seconds()
	if duration <= 0.1 {
		return "—", 0
	}

	for _, s := range samples {
		totalBytes += s.Bytes
	}

	bps := float64(totalBytes) / duration
	speedStr := formatBytes(int64(bps)) + "/s"

	eta := 0
	if bps > 1024 {
		eta = int(float64(remainingBytes) / bps)
	}

	return speedStr, eta
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
