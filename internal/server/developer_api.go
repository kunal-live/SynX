package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"synx/internal/capability"
	"synx/internal/capability/clipboard"
	"synx/internal/capability/command"
	"synx/internal/capability/files"
	"synx/internal/capability/terminal"
	"synx/internal/protocol"
)

type terminalOutputBuffer struct {
	mu     sync.Mutex
	buffer []byte
}

var (
	devRegistry     *capability.Registry
	terminalCap     *terminal.Capability
	commandCap      *command.Capability
	clipboardCap    *clipboard.Capability
	filesCap        *files.Capability
	terminalBuffers = make(map[string]*terminalOutputBuffer)
	termBufMu       sync.RWMutex
)

func initDeveloperPlatform(cfg Config) {
	devRegistry = capability.NewRegistry()

	terminalCap = terminal.New(func(sessionID string, data []byte) {
		termBufMu.Lock()
		buf, exists := terminalBuffers[sessionID]
		if !exists {
			buf = &terminalOutputBuffer{}
			terminalBuffers[sessionID] = buf
		}
		termBufMu.Unlock()

		buf.mu.Lock()
		buf.buffer = append(buf.buffer, data...)
		// Keep ring buffer at max 64KB
		if len(buf.buffer) > 65536 {
			buf.buffer = buf.buffer[len(buf.buffer)-65536:]
		}
		buf.mu.Unlock()
	})

	commandCap = command.New(func(senderID string) bool {
		// Allow local node requests or cryptographically trusted peers
		if senderID == "" || senderID == "local" || (cfg.Identity != nil && senderID == cfg.Identity.DeviceID) {
			return true
		}
		if cfg.TrustStore != nil && cfg.TrustStore.IsTrusted(senderID) {
			return true
		}
		return false
	})

	clipboardCap = clipboard.New(true)
	filesCap = files.New(cfg.Dir)

	_ = devRegistry.Register(terminalCap)
	_ = devRegistry.Register(commandCap)
	_ = devRegistry.Register(clipboardCap)
	_ = devRegistry.Register(filesCap)
}

// registerDeveloperAPI attaches the Developer Platform routes per Section 10 & 15.
func (s *Server) registerDeveloperAPI(mux *http.ServeMux) {
	initDeveloperPlatform(s.cfg)

	mux.HandleFunc("/api/devices", s.apiDevices)
	mux.HandleFunc("/api/devices/", s.apiDeviceRouter)
	mux.HandleFunc("/api/sessions", s.apiSessions)
	mux.HandleFunc("/api/terminal/open", s.apiTerminalOpen)
	mux.HandleFunc("/api/terminal/input", s.apiTerminalInput)
	mux.HandleFunc("/api/terminal/output", s.apiTerminalOutput)
	mux.HandleFunc("/api/terminal/close", s.apiTerminalClose)
	mux.HandleFunc("/api/terminal/list", s.apiTerminalList)
	mux.HandleFunc("/api/command/execute", s.apiCommandExecute)
	mux.HandleFunc("/api/clipboard", s.apiClipboard)
}

// GET /api/devices
func (s *Server) apiDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peerList := s.cfg.PeerMgr.List()
	res := make([]map[string]any, 0, len(peerList))
	for _, p := range peerList {
		caps := p.Capabilities
		if len(caps) == 0 {
			caps = []string{"terminal", "command", "files", "clipboard"}
		}
		res = append(res, map[string]any{
			"device_id":    p.ID,
			"device_name":  p.Name,
			"address":      p.Address,
			"port":         p.Port,
			"platform":     p.Platform,
			"version":      p.Version,
			"capabilities": caps,
			"trusted":      p.Trusted,
			"status":       p.Status,
			"last_seen":    p.LastSeen.Unix(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"devices": res,
		"count":   len(res),
	})
}

// Routes /api/devices/:id and /api/devices/:id/capabilities and /api/devices/:id/actions
func (s *Server) apiDeviceRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	deviceID := parts[1]
	p, ok := s.cfg.PeerMgr.Get(deviceID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "device not found"})
		return
	}

	// GET /api/devices/:id/capabilities
	if len(parts) == 3 && parts[2] == "capabilities" {
		caps := p.Capabilities
		if len(caps) == 0 {
			caps = []string{"terminal", "command", "files", "clipboard"}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"device_id":    p.ID,
			"capabilities": caps,
		})
		return
	}

	// POST /api/devices/:id/actions
	if len(parts) == 3 && parts[2] == "actions" && r.Method == http.MethodPost {
		var req protocol.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req.SenderID = s.cfg.Identity.DeviceID

		// Dispatch via Capability Registry
		resp := devRegistry.Dispatch(r.Context(), &req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// GET /api/devices/:id
	caps := p.Capabilities
	if len(caps) == 0 {
		caps = []string{"terminal", "command", "files", "clipboard"}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"device_id":    p.ID,
		"device_name":  p.Name,
		"address":      p.Address,
		"port":         p.Port,
		"platform":     p.Platform,
		"version":      p.Version,
		"capabilities": caps,
		"trusted":      p.Trusted,
		"status":       p.Status,
		"last_seen":    p.LastSeen.Unix(),
	})
}

// GET /api/sessions
func (s *Server) apiSessions(w http.ResponseWriter, r *http.Request) {
	resp, _ := terminalCap.Handle(r.Context(), &protocol.Request{Action: "list"})
	w.Header().Set("Content-Type", "application/json")
	if resp != nil && resp.Success {
		w.Write(resp.Data)
	} else {
		w.Write([]byte("[]"))
	}
}

// POST /api/terminal/open
func (s *Server) apiTerminalOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Cols  int    `json:"cols"`
		Rows  int    `json:"rows"`
		Shell string `json:"shell"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	raw, _ := json.Marshal(payload)
	req := &protocol.Request{
		MessageID:  protocol.GenerateMessageID(),
		Capability: "terminal",
		Action:     "open",
		Payload:    raw,
	}

	resp, err := terminalCap.Handle(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /api/terminal/input
func (s *Server) apiTerminalInput(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID string `json:"session_id"`
		Data      string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := terminalCap.WriteStdin(payload.SessionID, []byte(payload.Data))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"written": true})
}

// GET /api/terminal/output?session_id=...&offset=...
func (s *Server) apiTerminalOutput(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	termBufMu.RLock()
	buf, exists := terminalBuffers[sessionID]
	termBufMu.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"output": "", "closed": true})
		return
	}

	buf.mu.Lock()
	output := string(buf.buffer)
	buf.buffer = buf.buffer[:0] // drain read buffer
	buf.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"output": output,
		"closed": false,
	})
}

// POST /api/terminal/close
func (s *Server) apiTerminalClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID string `json:"session_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	terminalCap.CloseSession(payload.SessionID)

	termBufMu.Lock()
	delete(terminalBuffers, payload.SessionID)
	termBufMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"closed": true})
}

// GET /api/terminal/list
func (s *Server) apiTerminalList(w http.ResponseWriter, r *http.Request) {
	resp, _ := terminalCap.Handle(r.Context(), &protocol.Request{Action: "list"})
	w.Header().Set("Content-Type", "application/json")
	if resp != nil {
		w.Write(resp.Data)
	} else {
		w.Write([]byte("[]"))
	}
}

// POST /api/command/execute
func (s *Server) apiCommandExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Command string `json:"command"`
		Dir     string `json:"dir"`
		Timeout int    `json:"timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	raw, _ := json.Marshal(payload)
	req := &protocol.Request{
		MessageID:  protocol.GenerateMessageID(),
		Capability: "command",
		Action:     "execute",
		Payload:    raw,
		SenderID:   s.cfg.Identity.DeviceID,
	}

	resp, err := commandCap.Handle(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GET/POST /api/clipboard
func (s *Server) apiClipboard(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		resp, _ := clipboardCap.Handle(r.Context(), &protocol.Request{Action: "get"})
		w.Header().Set("Content-Type", "application/json")
		if resp != nil {
			_ = json.NewEncoder(w).Encode(resp)
		}
		return
	}

	if r.Method == http.MethodPost {
		var payload struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		raw, _ := json.Marshal(payload)
		resp, _ := clipboardCap.Handle(r.Context(), &protocol.Request{
			Action:  "set",
			Payload: raw,
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
