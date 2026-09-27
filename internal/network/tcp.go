package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	MaxFrameSize = 32 * 1024 * 1024 // 32MB max frame size
)

// TCPConnection implements Connection over standard net.Conn using 4-byte length-prefix framing.
type TCPConnection struct {
	conn net.Conn
	mu   sync.Mutex
}

// WrapConn wraps an active net.Conn into a framed TCPConnection.
func WrapConn(conn net.Conn) *TCPConnection {
	return &TCPConnection{conn: conn}
}

// Dial connects to a remote TCP endpoint and returns a Connection.
func Dial(addr string, timeout time.Duration) (Connection, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("tcp dial failed: %w", err)
	}
	return WrapConn(conn), nil
}

// Send sends length-prefixed data frame over the wire.
func (c *TCPConnection) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(data) > MaxFrameSize {
		return fmt.Errorf("frame size %d exceeds max frame size %d", len(data), MaxFrameSize)
	}

	length := uint32(len(data))
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], length)

	if _, err := c.conn.Write(header[:]); err != nil {
		return fmt.Errorf("write header failed: %w", err)
	}
	if _, err := c.conn.Write(data); err != nil {
		return fmt.Errorf("write payload failed: %w", err)
	}
	return nil
}

// Receive reads the next length-prefixed data frame.
func (c *TCPConnection) Receive() ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[:])
	if length > MaxFrameSize {
		return nil, fmt.Errorf("incoming frame size %d exceeds limit %d", length, MaxFrameSize)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(c.conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// Close closes the underlying socket.
func (c *TCPConnection) Close() error {
	return c.conn.Close()
}

// RemoteAddr returns the remote host:port string.
func (c *TCPConnection) RemoteAddr() string {
	if c.conn != nil && c.conn.RemoteAddr() != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}
