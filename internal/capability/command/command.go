package command

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"time"

	"synx/internal/observability"
	"synx/internal/protocol"
)

// PermissionChecker validates whether the caller is permitted to execute commands.
type PermissionChecker func(senderID string) bool

// Capability implements secure remote command execution.
type Capability struct {
	checkPermission PermissionChecker
	defaultTimeout  time.Duration
}

// New creates a new Command execution capability.
func New(checkPerm PermissionChecker) *Capability {
	return &Capability{
		checkPermission: checkPerm,
		defaultTimeout:  30 * time.Second,
	}
}

func (c *Capability) Name() string {
	return "command"
}

func (c *Capability) Version() string {
	return "1.0.0"
}

func (c *Capability) Actions() []string {
	return []string{
		"execute",
	}
}

type ExecutePayload struct {
	Command string `json:"command"`
	Dir     string `json:"dir,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // in seconds
}

func (c *Capability) Handle(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if req.Action != "execute" {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrActionNotSupported, "unsupported action"), nil
	}

	// 1. Permission verification
	if c.checkPermission != nil && !c.checkPermission(req.SenderID) {
		observability.Warn("Unauthorized command execution attempt by sender %q", req.SenderID)
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrForbidden, "peer does not have permission to execute commands"), nil
	}

	var payload ExecutePayload
	if err := req.ParsePayload(&payload); err != nil || payload.Command == "" {
		return protocol.ErrorResponse(req.MessageID, c.Name(), req.Action, protocol.ErrInvalidRequest, "command string required"), nil
	}

	timeout := c.defaultTimeout
	if payload.Timeout > 0 {
		timeout = time.Duration(payload.Timeout) * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd.exe", "/c", payload.Command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", payload.Command)
	}

	if payload.Dir != "" {
		cmd.Dir = payload.Dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	observability.Info("Executed remote command %q (exit: %d, duration: %v)", payload.Command, exitCode, duration)

	return protocol.SuccessResponse(req.MessageID, c.Name(), req.Action, map[string]any{
		"command":     payload.Command,
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
	}), nil
}
