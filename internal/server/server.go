package server

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"synx/internal/chunk"
	"synx/internal/config"
	"synx/internal/events"
	"synx/internal/filesystem"
	"synx/internal/history"
	"synx/internal/identity"
	"synx/internal/integrity"
	"synx/internal/observability"
	"synx/internal/pairing"
	"synx/internal/peers"
	"synx/internal/protocol"
	"synx/internal/security"
	"synx/internal/storage"
)

type Config struct {
	Dir        string
	Token      string
	Port       int
	Web        embed.FS
	Identity   *identity.Identity
	TrustStore *security.TrustStore
	Pairing    *pairing.Service
	PeerMgr    *peers.Manager
	History    *history.Service
	DB         *storage.DB
	Bus        *events.Bus
	AppConfig  *config.Config
}

type incomingTransferState struct {
	ID              string
	FileName        string
	DestinationPath string
	Size            int64
	ExpectedSHA     string
	ChunkSize       int64
	PartPath        string
	Manifest        *chunk.Manifest
	CreatedAt       time.Time
}

type Server struct {
	cfg       Config
	httpSrv   *http.Server
	sessions  *security.SessionManager
	incoming  map[string]*incomingTransferState
	incomMu   sync.RWMutex
}

func New(cfg Config) *Server {
	return &Server{
		cfg:      cfg,
		sessions: security.NewSessionManager(24 * time.Hour),
		incoming: make(map[string]*incomingTransferState),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	s.Register(mux)

	addr := fmt.Sprintf("0.0.0.0:%d", s.cfg.Port)
	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Error("HTTP Server error: %v", err)
		}
	}()

	observability.Info("SynX HTTP service listening on %s", addr)
	return nil
}

func (s *Server) Stop() error {
	if s.httpSrv != nil {
		return s.httpSrv.Close()
	}
	return nil
}

func (s *Server) Register(mux *http.ServeMux) {
	// --- Production v1 API (Section 15) ---
	mux.HandleFunc("/api/v1/device/info", s.v1DeviceInfo)
	mux.HandleFunc("/api/v1/device/capabilities", s.v1DeviceCapabilities)
	mux.HandleFunc("/api/v1/pair/start", s.v1PairStart)
	mux.HandleFunc("/api/v1/pair/confirm", s.v1PairConfirm)

	mux.HandleFunc("/api/v1/files/list", s.auth(s.v1FilesList))
	mux.HandleFunc("/api/v1/files/stat", s.auth(s.v1FilesStat))
	mux.HandleFunc("/api/v1/files/hash", s.auth(s.v1FilesHash))

	mux.HandleFunc("/api/v1/transfers", s.auth(s.v1TransfersNegotiate))
	mux.HandleFunc("/api/v1/transfers/", s.auth(s.v1TransfersRouter))
	mux.HandleFunc("/api/v1/history", s.auth(s.v1History))

	// --- Legacy & Dashboard API (Backwards compatibility) ---
	mux.HandleFunc("/api/info", s.legacyInfo)
	mux.HandleFunc("/api/state", s.legacyState)
	mux.HandleFunc("/api/files", s.auth(s.legacyFiles))
	mux.HandleFunc("/api/download", s.auth(s.legacyDownload))
	mux.HandleFunc("/api/upload", s.auth(s.legacyUpload))
	mux.HandleFunc("/api/delete", s.auth(s.legacyDelete))
	mux.HandleFunc("/api/set-dir", s.auth(s.legacySetDir))
	mux.HandleFunc("/api/pair-peer", s.auth(s.legacyPairPeer))

	// Developer Platform API (Section 10 & 15)
	s.registerDeveloperAPI(mux)

	// Frontend Static UI
	mux.HandleFunc("/", s.index)
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(protocol.HeaderToken)
		sessionID := r.Header.Get(protocol.HeaderSession)

		// Check local loopback exemption for dashboard requests from same machine
		host := r.RemoteAddr
		isLoopback := strings.HasPrefix(host, "127.0.0.1:") || strings.HasPrefix(host, "[::1]:") || strings.HasPrefix(host, "localhost:")

		if token == s.cfg.Token {
			next(w, r)
			return
		}

		if sessionID != "" {
			if _, ok := s.sessions.Validate(sessionID); ok {
				next(w, r)
				return
			}
		}

		if isLoopback && (token == "" || token == "local") {
			next(w, r)
			return
		}

		http.Error(w, "unauthorized: valid X-SynX-Token or X-SynX-Session required", http.StatusUnauthorized)
	}
}

// /api/v1/device/info
func (s *Server) v1DeviceInfo(w http.ResponseWriter, r *http.Request) {
	summary := map[string]any{
		"name":        "SynX",
		"version":     "1.0.0",
		"device_id":   "",
		"device_name": "SynX Node",
		"platform":    "windows",
		"shared_dir":  s.cfg.Dir,
	}
	if s.cfg.Identity != nil {
		idSum := s.cfg.Identity.Summary()
		summary["device_id"] = idSum["device_id"]
		summary["device_name"] = idSum["device_name"]
		summary["platform"] = idSum["platform"]
		summary["public_key"] = idSum["public_key"]
	}
	writeJSON(w, summary)
}

// /api/v1/device/capabilities
func (s *Server) v1DeviceCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, protocol.DefaultCapabilities())
}

// /api/v1/pair/start
func (s *Server) v1PairStart(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Pairing == nil {
		http.Error(w, "pairing service unavailable", 503)
		return
	}
	pin := s.cfg.Pairing.GeneratePIN()
	writeJSON(w, map[string]string{
		"status": "pending",
		"pin":    pin,
	})
}

// /api/v1/pair/confirm
func (s *Server) v1PairConfirm(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Pairing == nil {
		http.Error(w, "pairing service unavailable", 503)
		return
	}
	var req pairing.PairingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if !s.cfg.Pairing.VerifyPIN(req.PIN) {
		http.Error(w, "invalid or expired pairing PIN", http.StatusForbidden)
		return
	}

	peer := peers.Peer{
		ID:        req.DeviceID,
		Name:      req.Name,
		PublicKey: req.PublicKey,
		Status:    peers.StatusTrusted,
		Trusted:   true,
		LastSeen:  time.Now(),
	}
	if s.cfg.PeerMgr != nil {
		s.cfg.PeerMgr.AddOrUpdate(peer)
	}
	if s.cfg.TrustStore != nil {
		_ = s.cfg.TrustStore.TrustPeer(security.TrustedPeer{
			PeerID:    req.DeviceID,
			Name:      req.Name,
			PublicKey: req.PublicKey,
			TrustedAt: time.Now(),
		})
	}

	sess := s.sessions.CreateSession(req.DeviceID, []string{"transfer", "sync"})
	writeJSON(w, pairing.PairingResponse{
		Success:   true,
		DeviceID:  s.cfg.Identity.DeviceID,
		Name:      s.cfg.Identity.DeviceName,
		PublicKey: s.cfg.Identity.PublicKey,
		Token:     sess.SessionID,
	})
}

// /api/v1/files/list
func (s *Server) v1FilesList(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	items, err := filesystem.ScanDirectory(s.cfg.Dir, rel, filesystem.ScanOptions{
		Recursive:    false,
		ComputeHash:  false,
		ExcludeParts: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, items)
}

// /api/v1/files/stat
func (s *Server) v1FilesStat(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	target, err := filesystem.ResolveSafePath(s.cfg.Dir, rel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fi, err := os.Stat(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, filesystem.FileItem{
		Name:     fi.Name(),
		Path:     rel,
		Size:     fi.Size(),
		IsDir:    fi.IsDir(),
		Modified: fi.ModTime(),
	})
}

// /api/v1/files/hash
func (s *Server) v1FilesHash(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	target, err := filesystem.ResolveSafePath(s.cfg.Dir, rel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hash, err := integrity.FileSHA256(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"sha256": hash})
}

// /api/v1/transfers (POST negotiation)
func (s *Server) v1TransfersNegotiate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.TransferNegotiationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	destPath := req.Destination
	if destPath == "" {
		destPath = req.FileName
	}
	if destPath == "" {
		destPath = filepath.Base(req.SourcePath)
	}

	resolvedFinal, err := filesystem.ResolveSafePath(s.cfg.Dir, destPath)
	if err != nil {
		writeJSON(w, protocol.TransferNegotiationResponse{
			Accepted: false,
			Reason:   "invalid target path: " + err.Error(),
		})
		return
	}

	txID := "tx_" + security.GenerateToken(8)
	partPath := filesystem.PartFilePath(resolvedFinal)

	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = chunk.DefaultChunkSize
	}

	manifest := chunk.NewManifest(txID, req.FileName, req.Size, chunkSize)
	state := &incomingTransferState{
		ID:              txID,
		FileName:        req.FileName,
		DestinationPath: resolvedFinal,
		Size:            req.Size,
		ExpectedSHA:     req.SHA256,
		ChunkSize:       chunkSize,
		PartPath:        partPath,
		Manifest:        manifest,
		CreatedAt:       time.Now(),
	}

	s.incomMu.Lock()
	s.incoming[txID] = state
	s.incomMu.Unlock()

	resumeOffset := int64(0)
	if fi, err := os.Stat(partPath); err == nil && fi.Size() > 0 {
		resumeOffset = fi.Size()
	}

	writeJSON(w, protocol.TransferNegotiationResponse{
		TransferID:   txID,
		Accepted:     true,
		ChunkSize:    chunkSize,
		ResumeOffset: resumeOffset,
	})
}

// /api/v1/transfers/{id}/* router
func (s *Server) v1TransfersRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/transfers/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "transfer ID required", 400)
		return
	}

	txID := parts[0]
	if len(parts) == 1 {
		// GET transfer status
		s.incomMu.RLock()
		st, ok := s.incoming[txID]
		s.incomMu.RUnlock()
		if !ok {
			http.Error(w, "transfer not found", 404)
			return
		}
		writeJSON(w, map[string]any{
			"id":        st.ID,
			"name":      st.FileName,
			"size":      st.Size,
			"done":      st.Manifest.CompletedBytes(),
			"completed": st.Manifest.CompletedIndices(),
		})
		return
	}

	action := parts[1]
	switch action {
	case "chunks":
		// PUT /api/v1/transfers/{id}/chunks/{chunk_index}
		if len(parts) < 3 {
			http.Error(w, "chunk index required", 400)
			return
		}
		idx, _ := strconv.Atoi(parts[2])
		s.handleChunkUpload(w, r, txID, idx)

	case "commit":
		// POST /api/v1/transfers/{id}/commit
		s.handleCommit(w, r, txID)

	default:
		http.Error(w, "unknown transfer endpoint", 404)
	}
}

func (s *Server) handleChunkUpload(w http.ResponseWriter, r *http.Request, txID string, chunkIndex int) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}

	s.incomMu.RLock()
	st, ok := s.incoming[txID]
	s.incomMu.RUnlock()
	if !ok {
		http.Error(w, "transfer session not found", 404)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read chunk payload", 500)
		return
	}
	observability.DefaultMetrics().AddBytesReceived(int64(len(data)))

	offset := int64(chunkIndex) * st.ChunkSize
	ch := &chunk.Chunk{
		TransferID: txID,
		Index:      chunkIndex,
		Offset:     offset,
		Size:       int64(len(data)),
		SHA256:     integrity.BytesSHA256(data),
		Data:       data,
	}

	claimedHash := r.Header.Get(protocol.HeaderChunkHash)
	if err := chunk.WriteChunk(st.PartPath, ch, claimedHash); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	st.Manifest.MarkCompleted(chunkIndex, ch.Size)
	writeJSON(w, map[string]any{
		"chunk":    chunkIndex,
		"accepted": true,
		"written":  len(data),
	})
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request, txID string) {
	s.incomMu.Lock()
	st, ok := s.incoming[txID]
	s.incomMu.Unlock()
	if !ok {
		http.Error(w, "transfer not found", 404)
		return
	}

	var req protocol.CommitRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	expected := req.SHA256
	if expected == "" {
		expected = st.ExpectedSHA
	}

	// Verify SHA-256 of .synx.part
	actualHash, err := integrity.FileSHA256(st.PartPath)
	if err != nil {
		http.Error(w, "integrity calculation failed: "+err.Error(), 500)
		return
	}

	if expected != "" && actualHash != expected {
		writeJSON(w, protocol.CommitResponse{
			Status:   "failed",
			Verified: false,
			Error:    fmt.Sprintf("hash mismatch: expected %s, got %s", expected, actualHash),
		})
		return
	}

	// Commit part file to final filename
	if err := filesystem.CommitPartFile(st.DestinationPath); err != nil {
		http.Error(w, "commit file failed: "+err.Error(), 500)
		return
	}

	if s.cfg.History != nil && s.cfg.DB != nil {
		_ = s.cfg.DB.AddHistory(storage.HistoryRecord{
			ID:          "hist_" + security.GenerateToken(8),
			TransferID:  st.ID,
			PeerID:      "remote",
			FileName:    st.FileName,
			Size:        st.Size,
			Direction:   "receive",
			Status:      "completed",
			CompletedAt: time.Now().Unix(),
		})
	}

	observability.DefaultMetrics().IncCompletedSuccess()
	writeJSON(w, protocol.CommitResponse{
		Status:   "completed",
		Verified: true,
	})
}

func txIDFromURL(p string) string {
	parts := strings.Split(strings.TrimPrefix(p, "/api/v1/transfers/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// /api/v1/history
func (s *Server) v1History(w http.ResponseWriter, r *http.Request) {
	if s.cfg.History == nil {
		writeJSON(w, []any{})
		return
	}
	entries, _ := s.cfg.History.List(100)
	writeJSON(w, entries)
}

// --- Legacy Handlers ---
func (s *Server) legacyInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"name":       "SynX",
		"version":    "0.2.0-desktop",
		"platform":   "windows",
		"shared_dir": s.cfg.Dir,
	})
}

func (s *Server) legacyState(w http.ResponseWriter, r *http.Request) {
	peersList := []peers.Peer{}
	if s.cfg.PeerMgr != nil {
		peersList = s.cfg.PeerMgr.List()
	}
	deviceID := ""
	if s.cfg.Identity != nil {
		deviceID = s.cfg.Identity.DeviceID
	}
	writeJSON(w, map[string]any{
		"name":      "SynX",
		"version":   "1.0.0",
		"device_id": deviceID,
		"platform":  runtime.GOOS,
		"sharedDir": s.cfg.Dir,
		"token":     s.cfg.Token,
		"address":   r.Host,
		"peers":     peersList,
		"transfers": []any{},
		"metrics":   observability.DefaultMetrics().Snapshot(),
	})
}

func (s *Server) legacyFiles(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	items, err := filesystem.ScanDirectory(s.cfg.Dir, rel, filesystem.ScanOptions{ExcludeParts: true})
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	type legacyItem struct {
		Name   string `json:"name"`
		Size   int64  `json:"size"`
		Dir    bool   `json:"dir"`
		SHA256 string `json:"sha256,omitempty"`
	}
	var out []legacyItem
	for _, it := range items {
		out = append(out, legacyItem{
			Name: it.Name,
			Size: it.Size,
			Dir:  it.IsDir,
		})
	}
	writeJSON(w, out)
}

func (s *Server) legacyDownload(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	target, err := filesystem.ResolveSafePath(s.cfg.Dir, rel)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	f, err := os.Open(target)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	defer f.Close()

	st, _ := f.Stat()
	observability.DefaultMetrics().AddBytesSent(st.Size())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", st.Name()))
	http.ServeContent(w, r, st.Name(), st.ModTime(), f)
}

func (s *Server) legacyUpload(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	target, err := filesystem.ResolveSafePath(s.cfg.Dir, rel)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	offset := int64(0)
	fmt.Sscan(r.Header.Get("X-SynX-Offset"), &offset)

	flags := os.O_CREATE | os.O_WRONLY
	if offset == 0 {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(target, flags, 0644)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()

	if _, err = f.Seek(offset, io.SeekStart); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	n, err := io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	observability.DefaultMetrics().AddBytesReceived(n)

	h := sha256.New()
	_, _ = f.Seek(0, io.SeekStart)
	_, _ = io.Copy(h, f)

	writeJSON(w, map[string]any{
		"written": n,
		"offset":  offset + n,
		"sha256":  hex.EncodeToString(h.Sum(nil)),
	})
}

func (s *Server) legacyDelete(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	target, err := filesystem.ResolveSafePath(s.cfg.Dir, rel)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if err := os.RemoveAll(target); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) legacySetDir(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Dir string `json:"dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if body.Dir == "" {
		http.Error(w, "dir required", 400)
		return
	}
	_ = os.MkdirAll(body.Dir, 0755)
	s.cfg.Dir = body.Dir
	if s.cfg.AppConfig != nil {
		_ = s.cfg.AppConfig.SetSharedDir(body.Dir)
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) legacyPairPeer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Address string `json:"address"`
		Token   string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if s.cfg.Pairing == nil {
		http.Error(w, "pairing service unavailable", 503)
		return
	}
	peer, err := s.cfg.Pairing.PairDirect(body.Address, body.Token)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, peer)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	var b []byte
	var err error

	// Try reading from embedded webFS
	b, err = s.cfg.Web.ReadFile("desktop/frontend/index.html")
	if err != nil {
		b, err = s.cfg.Web.ReadFile("web/index.html")
	}
	if err != nil {
		// Fallback to disk read
		b, err = os.ReadFile("desktop/frontend/index.html")
	}
	if err != nil {
		http.Error(w, "frontend index.html not found", 404)
		return
	}

	html := strings.Replace(string(b), "</head>", fmt.Sprintf("<script>window.__SYNX_LOCAL__={token:%q,dir:%q};</script></head>", s.cfg.Token, s.cfg.Dir), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
