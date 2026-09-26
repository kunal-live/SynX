package transfer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"synx/internal/chunk"
	"synx/internal/events"
	"synx/internal/peers"
	"synx/internal/protocol"
	"synx/internal/storage"
)

type mockTransport struct {
	negotiateCalled bool
	chunksSent      int
	commitCalled    bool
}

func (m *mockTransport) Negotiate(ctx context.Context, peer peers.Peer, req protocol.TransferNegotiationRequest) (*protocol.TransferNegotiationResponse, error) {
	m.negotiateCalled = true
	return &protocol.TransferNegotiationResponse{
		TransferID: "mock_tx_1",
		Accepted:   true,
		ChunkSize:  req.ChunkSize,
	}, nil
}

func (m *mockTransport) SendChunk(ctx context.Context, peer peers.Peer, transferID string, ch *chunk.Chunk) error {
	m.chunksSent++
	return nil
}

func (m *mockTransport) Commit(ctx context.Context, peer peers.Peer, transferID, sha256 string) (*protocol.CommitResponse, error) {
	m.commitCalled = true
	return &protocol.CommitResponse{
		Status:   "completed",
		Verified: true,
	}, nil
}

func TestTransferManagerFlow(t *testing.T) {
	tempDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("storage.Open failed: %v", err)
	}
	defer db.Close()

	bus := events.NewBus()
	peerMgr := peers.NewManager(db, bus)
	peerMgr.AddOrUpdate(peers.Peer{
		ID:      "peer_target",
		Name:    "Target Device",
		Address: "127.0.0.1:8787",
		Status:  peers.StatusTrusted,
		Trusted: true,
	})

	mockTr := &mockTransport{}
	mgr := NewManager(db, bus, peerMgr, mockTr, 1)
	defer mgr.Close()

	// Create test file
	srcFile := filepath.Join(tempDir, "transfer_test.dat")
	data := make([]byte, 1024*1024) // 1MB
	_ = os.WriteFile(srcFile, data, 0644)

	tx, err := mgr.CreateTransfer("peer_target", srcFile, "transfer_test.dat")
	if err != nil {
		t.Fatalf("CreateTransfer failed: %v", err)
	}

	// Wait up to 3 seconds for transfer completion
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		cur, ok := mgr.Get(tx.ID)
		if ok && cur.Status == StatusCompleted {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	completedTx, ok := mgr.Get(tx.ID)
	if !ok || completedTx.Status != StatusCompleted {
		t.Fatalf("expected transfer to complete, status is: %s", completedTx.Status)
	}

	if !mockTr.negotiateCalled || mockTr.chunksSent == 0 || !mockTr.commitCalled {
		t.Errorf("mock transport expectation failed: nego=%v, chunks=%d, commit=%v", mockTr.negotiateCalled, mockTr.chunksSent, mockTr.commitCalled)
	}

	// Check history record in database
	hist, err := db.ListHistory(10)
	if err != nil || len(hist) == 0 {
		t.Fatalf("expected history record to be written in DB, got len=%d, err=%v", len(hist), err)
	}
}
