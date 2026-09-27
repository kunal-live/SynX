package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"synx/internal/observability"
	"synx/internal/protocol"
)

// Session represents an active terminal shell session.
type Session struct {
	ID        string
	Cmd       *exec.Cmd
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	Stderr    io.ReadCloser
	CreatedAt time.Time
	Cols      int
	Rows      int
	onOutput  func(data []byte)
	mu        sync.Mutex
	closed    bool
}

// Capability implements the interactive terminal capability.
type Capability struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	onOutput func(sessionID string, data []byte)
}

// New creates a new Terminal capability.
func New(onOutput func(sessionID string, data []byte)) *Capability {
	return &Capability{
		sessions: make(map[string]*Session),
		onOutput: onOutput,
	}
}

func (c *Capability) Name() string {
	return "terminal"
}

func (c *Capability) Version() string {
	return "1.0.0"
}

func (c *Capability) Actions() []string {
	return []string{
		"open",
		"input",
		"resize",
		"close",
		"list",
	}
}

func (c *Capability) Handle(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	switch req.Action {
	case "open":
		return c.handleOpen(req)
	case "input":
		return c.handleInput(req)
	case "resize":
		return c.handleResize(req)
	case "close":
		return c.handleClose(req)
	case "list":
		return c.handleList(req)
	default:
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrActionNotSupported, "unsupported terminal action"), nil
	}
}

type OpenPayload struct {
	Cols  int    `json:"cols,omitempty"`
	Rows  int    `json:"rows,omitempty"`
	Shell string `json:"shell,omitempty"`
}

func (c *Capability) handleOpen(req *protocol.Request) (*protocol.Response, error) {
	var payload OpenPayload
	_ = req.ParsePayload(&payload)

	cols := payload.Cols
	if cols <= 0 {
		cols = 80
	}
	rows := payload.Rows
	if rows <= 0 {
		rows = 24
	}

	shellPath := payload.Shell
	if shellPath == "" {
		if runtime.GOOS == "windows" {
			shellPath = os.Getenv("COMSPEC")
			if shellPath == "" {
				shellPath = "cmd.exe"
			}
		} else {
			shellPath = os.Getenv("SHELL")
			if shellPath == "" {
				shellPath = "/bin/sh"
			}
		}
	}

	cmd := exec.Command(shellPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInternalError, fmt.Sprintf("failed to create stdin pipe: %v", err)), nil
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInternalError, fmt.Sprintf("failed to create stdout pipe: %v", err)), nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInternalError, fmt.Sprintf("failed to create stderr pipe: %v", err)), nil
	}

	if err := cmd.Start(); err != nil {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInternalError, fmt.Sprintf("failed to launch shell %q: %v", shellPath, err)), nil
	}

	sessionID := protocol.GenerateMessageID()
	sess := &Session{
		ID:        sessionID,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		CreatedAt: time.Now(),
		Cols:      cols,
		Rows:      rows,
	}

	c.mu.Lock()
	c.sessions[sessionID] = sess
	c.mu.Unlock()

	// Stream stdout & stderr in background
	go c.pipeOutput(sessionID, stdout)
	go c.pipeOutput(sessionID, stderr)

	// Wait for process exit in background
	go func() {
		_ = cmd.Wait()
		c.mu.Lock()
		delete(c.sessions, sessionID)
		c.mu.Unlock()
		observability.Info("Terminal session %s exited", sessionID)
	}()

	observability.Info("Opened terminal session %s (shell: %s)", sessionID, shellPath)

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]any{
		"session_id": sessionID,
		"shell":      shellPath,
		"cols":       cols,
		"rows":       rows,
		"created_at": sess.CreatedAt.Unix(),
	}), nil
}

type InputPayload struct {
	SessionID string `json:"session_id"`
	Data      string `json:"data"`
}

func (c *Capability) handleInput(req *protocol.Request) (*protocol.Response, error) {
	var payload InputPayload
	if err := req.ParsePayload(&payload); err != nil || payload.SessionID == "" {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInvalidRequest, "missing session_id or data"), nil
	}

	c.mu.RLock()
	sess, exists := c.sessions[payload.SessionID]
	c.mu.RUnlock()

	if !exists {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrSessionNotFound, fmt.Sprintf("session %q not found", payload.SessionID)), nil
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.closed {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrSessionNotFound, "session closed"), nil
	}

	if _, err := io.WriteString(sess.Stdin, payload.Data); err != nil {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInternalError, fmt.Sprintf("write to stdin failed: %v", err)), nil
	}

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]bool{"written": true}), nil
}

type ResizePayload struct {
	SessionID string `json:"session_id"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

func (c *Capability) handleResize(req *protocol.Request) (*protocol.Response, error) {
	var payload ResizePayload
	if err := req.ParsePayload(&payload); err != nil || payload.SessionID == "" {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInvalidRequest, "invalid resize payload"), nil
	}

	c.mu.Lock()
	sess, exists := c.sessions[payload.SessionID]
	if exists {
		sess.Cols = payload.Cols
		sess.Rows = payload.Rows
	}
	c.mu.Unlock()

	if !exists {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrSessionNotFound, "session not found"), nil
	}

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]bool{"resized": true}), nil
}

type ClosePayload struct {
	SessionID string `json:"session_id"`
}

func (c *Capability) handleClose(req *protocol.Request) (*protocol.Response, error) {
	var payload ClosePayload
	if err := req.ParsePayload(&payload); err != nil || payload.SessionID == "" {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInvalidRequest, "missing session_id"), nil
	}

	c.mu.Lock()
	sess, exists := c.sessions[payload.SessionID]
	delete(c.sessions, payload.SessionID)
	c.mu.Unlock()

	if !exists {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrSessionNotFound, "session not found"), nil
	}

	sess.mu.Lock()
	sess.closed = true
	_ = sess.Stdin.Close()
	if sess.Cmd != nil && sess.Cmd.Process != nil {
		_ = sess.Cmd.Process.Kill()
	}
	sess.mu.Unlock()

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]bool{"closed": true}), nil
}

func (c *Capability) handleList(req *protocol.Request) (*protocol.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]map[string]any, 0, len(c.sessions))
	for id, sess := range c.sessions {
		list = append(list, map[string]any{
			"session_id": id,
			"cols":       sess.Cols,
			"rows":       sess.Rows,
			"created_at": sess.CreatedAt.Unix(),
		})
	}

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, list), nil
}

func (c *Capability) pipeOutput(sessionID string, r io.Reader) {
	buf := make([]byte, 2048)
	for {
		n, err := r.Read(buf)
		if n > 0 && c.onOutput != nil {
			c.onOutput(sessionID, buf[:n])
		}
		if err != nil {
			return
		}
	}
}

// WriteStdin directly feeds input to an active terminal session.
func (c *Capability) WriteStdin(sessionID string, data []byte) error {
	c.mu.RLock()
	sess, exists := c.sessions[sessionID]
	c.mu.RUnlock()
	if !exists {
		return fmt.Errorf("session %q not found", sessionID)
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if sess.closed {
		return fmt.Errorf("session %q closed", sessionID)
	}
	_, err := sess.Stdin.Write(data)
	return err
}

// CloseSession terminates an active session.
func (c *Capability) CloseSession(sessionID string) {
	c.mu.Lock()
	sess, exists := c.sessions[sessionID]
	delete(c.sessions, sessionID)
	c.mu.Unlock()
	if exists {
		sess.mu.Lock()
		sess.closed = true
		_ = sess.Stdin.Close()
		if sess.Cmd != nil && sess.Cmd.Process != nil {
			_ = sess.Cmd.Process.Kill()
		}
		sess.mu.Unlock()
	}
}
