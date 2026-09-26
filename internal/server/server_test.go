package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"synx/internal/chunk"
	"synx/internal/identity"
	"synx/internal/integrity"
	"synx/internal/protocol"
)

func TestServerTransferFlow(t *testing.T) {
	tempDir := t.TempDir()
	id, err := identity.LoadOrCreate(tempDir, "Test-Server")
	if err != nil {
		t.Fatalf("identity.LoadOrCreate failed: %v", err)
	}

	token := "test-secret-token"
	srv := New(Config{
		Dir:      tempDir,
		Token:    token,
		Identity: id,
	})

	mux := http.NewServeMux()
	srv.Register(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 1. Check Device Info
	resp, err := http.Get(ts.URL + "/api/v1/device/info")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("v1DeviceInfo failed: status=%v, err=%v", resp.StatusCode, err)
	}
	var devInfo map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&devInfo)
	if devInfo["device_id"] != id.DeviceID {
		t.Errorf("device ID mismatch: got %v, expected %s", devInfo["device_id"], id.DeviceID)
	}

	// 2. Negotiate Transfer
	testContent := []byte("SynX Protocol V1 Chunked Transfer Data")
	expectedHash := integrity.BytesSHA256(testContent)

	negoReq := protocol.TransferNegotiationRequest{
		FileName:  "received.txt",
		Size:      int64(len(testContent)),
		SHA256:    expectedHash,
		ChunkSize: chunk.DefaultChunkSize,
	}
	negoBody, _ := json.Marshal(negoReq)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/transfers", bytes.NewReader(negoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(protocol.HeaderToken, token)
	negoResp, err := http.DefaultClient.Do(req)
	if err != nil || negoResp.StatusCode != http.StatusOK {
		t.Fatalf("Negotiate failed: status=%v, err=%v", negoResp.StatusCode, err)
	}

	var negoRes protocol.TransferNegotiationResponse
	_ = json.NewDecoder(negoResp.Body).Decode(&negoRes)
	if !negoRes.Accepted || negoRes.TransferID == "" {
		t.Fatalf("transfer not accepted: %+v", negoRes)
	}
	txID := negoRes.TransferID

	// 3. Upload Chunk 0
	chunkReq, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/transfers/"+txID+"/chunks/0", bytes.NewReader(testContent))
	chunkReq.Header.Set("Content-Type", "application/octet-stream")
	chunkReq.Header.Set(protocol.HeaderToken, token)
	chunkReq.Header.Set(protocol.HeaderChunkOffset, "0")
	chunkReq.Header.Set(protocol.HeaderChunkHash, expectedHash)

	chunkResp, err := http.DefaultClient.Do(chunkReq)
	if err != nil || chunkResp.StatusCode != http.StatusOK {
		t.Fatalf("chunk upload failed: status=%v, err=%v", chunkResp.StatusCode, err)
	}

	// 4. Commit Transfer
	commitBody, _ := json.Marshal(protocol.CommitRequest{SHA256: expectedHash})
	commitReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/transfers/"+txID+"/commit", bytes.NewReader(commitBody))
	commitReq.Header.Set("Content-Type", "application/json")
	commitReq.Header.Set(protocol.HeaderToken, token)

	commitResp, err := http.DefaultClient.Do(commitReq)
	if err != nil || commitResp.StatusCode != http.StatusOK {
		t.Fatalf("commit request failed: status=%v, err=%v", commitResp.StatusCode, err)
	}

	var commitRes protocol.CommitResponse
	_ = json.NewDecoder(commitResp.Body).Decode(&commitRes)
	if !commitRes.Verified || commitRes.Status != "completed" {
		t.Fatalf("commit verification failed: %+v", commitRes)
	}

	// 5. Verify the file exists on disk with exact content
	committedFile := filepath.Join(tempDir, "received.txt")
	diskData, err := os.ReadFile(committedFile)
	if err != nil || string(diskData) != string(testContent) {
		t.Fatalf("disk file mismatch: err=%v, content=%s", err, string(diskData))
	}
}
