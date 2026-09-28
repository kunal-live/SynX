package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"synx/internal/chunk"
	"synx/internal/observability"
	"synx/internal/peers"
	"synx/internal/protocol"
)

func TestPeerURL(t *testing.T) {
	p1 := peers.Peer{Address: "192.168.1.5:8787"}
	if u := peerURL(p1, "/test"); u != "http://192.168.1.5:8787/test" {
		t.Errorf("expected http://192.168.1.5:8787/test, got %s", u)
	}

	p2 := peers.Peer{Address: "https://secure.node.lan:9000/"}
	if u := peerURL(p2, "api/v1/transfers"); u != "https://secure.node.lan:9000/api/v1/transfers" {
		t.Errorf("expected https://secure.node.lan:9000/api/v1/transfers, got %s", u)
	}
}

func TestHTTPTransportFlow(t *testing.T) {
	observability.DefaultMetrics().Reset()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/transfers":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(protocol.TransferNegotiationResponse{
				TransferID: "tx_123",
				Accepted:   true,
				ChunkSize:  1024,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v1/transfers/tx_123/chunks/"):
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/v1/transfers/tx_123/commit":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(protocol.CommitResponse{
				Status:   "completed",
				Verified: true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	addr := strings.TrimPrefix(ts.URL, "http://")
	peer := peers.Peer{
		ID:      "test-peer",
		Address: addr,
	}

	tr := NewHTTPTransport()
	ctx := context.Background()

	// 1. Negotiate
	nego, err := tr.Negotiate(ctx, peer, protocol.TransferNegotiationRequest{
		FileName: "test.txt",
		Size:     1024,
	})
	if err != nil {
		t.Fatalf("negotiation failed: %v", err)
	}
	if nego.TransferID != "tx_123" || !nego.Accepted {
		t.Fatalf("unexpected negotiation response: %+v", nego)
	}

	// 2. SendChunk
	ch := &chunk.Chunk{
		TransferID: "tx_123",
		Index:      0,
		Offset:     0,
		Size:       512,
		Data:       make([]byte, 512),
		SHA256:     "hash0",
	}
	if err := tr.SendChunk(ctx, peer, "tx_123", ch); err != nil {
		t.Fatalf("SendChunk failed: %v", err)
	}

	if observability.DefaultMetrics().Snapshot()["bytes_sent"] != int64(512) {
		t.Errorf("expected bytes_sent to be 512, got %v", observability.DefaultMetrics().Snapshot()["bytes_sent"])
	}

	// 3. Commit
	commitRes, err := tr.Commit(ctx, peer, "tx_123", "dummyhash")
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	if !commitRes.Verified {
		t.Errorf("expected commit verified true, got false")
	}
}

func TestHTTPTransportErrors(t *testing.T) {
	tr := NewHTTPTransport()
	ctx := context.Background()

	// Negotiate to non-existent server should return error (not silent success!)
	peerDead := peers.Peer{Address: "127.0.0.1:59999"}
	_, err := tr.Negotiate(ctx, peerDead, protocol.TransferNegotiationRequest{})
	if err == nil {
		t.Fatalf("expected error negotiating to dead peer, got nil")
	}
}
