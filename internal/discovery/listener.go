package discovery

import (
	"fmt"
	"net"
	"sync"

	"synx/internal/observability"
)

// HandlerFunc is invoked whenever a valid remote peer discovery message is received.
type HandlerFunc func(msg *Message, remoteIP net.IP)

// Listener binds to a UDP port and listens for incoming LAN discovery announcements.
type Listener struct {
	port       int
	selfID     string
	handler    HandlerFunc
	conn       *net.UDPConn
	stopCh     chan struct{}
	running    bool
	mu         sync.Mutex
}

// NewListener creates a new discovery datagram listener.
func NewListener(port int, selfID string, handler HandlerFunc) *Listener {
	return &Listener{
		port:    port,
		selfID:  selfID,
		handler: handler,
		stopCh:  make(chan struct{}),
	}
}

// Start binds the socket and spawns the background listening routine.
func (l *Listener) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return nil
	}

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", l.port))
	if err != nil {
		return fmt.Errorf("failed to resolve discovery UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("failed to bind discovery UDP port %d: %w", l.port, err)
	}

	l.conn = conn
	l.running = true

	go l.loop()
	observability.Info("Discovery listener active on UDP 0.0.0.0:%d", l.port)
	return nil
}

// Stop closes the UDP socket and stops reading.
func (l *Listener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}
	l.running = false
	close(l.stopCh)
	if l.conn != nil {
		_ = l.conn.Close()
	}
}

func (l *Listener) loop() {
	buf := make([]byte, 4096)

	for {
		select {
		case <-l.stopCh:
			return
		default:
		}

		n, remoteAddr, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-l.stopCh:
				return
			default:
				continue
			}
		}

		if n == 0 {
			continue
		}

		msg, err := DecodeMessage(buf[:n])
		if err != nil {
			continue
		}

		// Filter self-announcements
		if l.selfID != "" && msg.DeviceID == l.selfID {
			continue
		}

		if l.handler != nil {
			l.handler(msg, remoteAddr.IP)
		}
	}
}
