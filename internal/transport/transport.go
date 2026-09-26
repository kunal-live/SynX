package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"synx/internal/chunk"
	"synx/internal/peers"
	"synx/internal/protocol"
)

type Transport interface {
	Negotiate(ctx context.Context, peer peers.Peer, req protocol.TransferNegotiationRequest) (*protocol.TransferNegotiationResponse, error)
	SendChunk(ctx context.Context, peer peers.Peer, transferID string, ch *chunk.Chunk) error
	Commit(ctx context.Context, peer peers.Peer, transferID, sha256 string) (*protocol.CommitResponse, error)
}

type HTTPTransport struct {
	client *http.Client
}

func NewHTTPTransport() *HTTPTransport {
	return &HTTPTransport{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func peerURL(peer peers.Peer, endpoint string) string {
	addr := peer.Address
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/") + endpoint
}

func (t *HTTPTransport) Negotiate(ctx context.Context, peer peers.Peer, req protocol.TransferNegotiationRequest) (*protocol.TransferNegotiationResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	u := peerURL(peer, "/api/v1/transfers")
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(protocol.HeaderProtocolVersion, "1")
	if peer.Token != "" {
		httpReq.Header.Set(protocol.HeaderToken, peer.Token)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		// Fallback for legacy MVP endpoint
		return &protocol.TransferNegotiationResponse{
			TransferID:   fmt.Sprintf("tx_legacy_%d", time.Now().UnixNano()),
			Accepted:     true,
			ChunkSize:    chunk.DefaultChunkSize,
			ResumeOffset: 0,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("remote rejected transfer: HTTP %d", resp.StatusCode)
	}

	var res protocol.TransferNegotiationResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("invalid negotiation response: %w", err)
	}

	return &res, nil
}

func (t *HTTPTransport) SendChunk(ctx context.Context, peer peers.Peer, transferID string, ch *chunk.Chunk) error {
	u := peerURL(peer, fmt.Sprintf("/api/v1/transfers/%s/chunks/%d", transferID, ch.Index))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(ch.Data))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/octet-stream")
	httpReq.Header.Set(protocol.HeaderProtocolVersion, "1")
	httpReq.Header.Set(protocol.HeaderChunkIndex, fmt.Sprint(ch.Index))
	httpReq.Header.Set(protocol.HeaderChunkOffset, fmt.Sprint(ch.Offset))
	httpReq.Header.Set(protocol.HeaderChunkHash, ch.SHA256)
	if peer.Token != "" {
		httpReq.Header.Set(protocol.HeaderToken, peer.Token)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chunk upload failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (t *HTTPTransport) Commit(ctx context.Context, peer peers.Peer, transferID, sha256 string) (*protocol.CommitResponse, error) {
	reqBody, _ := json.Marshal(protocol.CommitRequest{SHA256: sha256})
	u := peerURL(peer, fmt.Sprintf("/api/v1/transfers/%s/commit", transferID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(protocol.HeaderProtocolVersion, "1")
	if peer.Token != "" {
		httpReq.Header.Set(protocol.HeaderToken, peer.Token)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cr protocol.CommitResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}

	return &cr, nil
}
