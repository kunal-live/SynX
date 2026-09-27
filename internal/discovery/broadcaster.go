package discovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"synx/internal/observability"
)

// Broadcaster sends periodic discovery announcements over UDP multicast and broadcast.
type Broadcaster struct {
	port      int
	messageFn func() *Message
	interval  time.Duration
	stopCh    chan struct{}
	running   bool
	mu        sync.Mutex
}

// NewBroadcaster constructs a periodic discovery announcement sender.
func NewBroadcaster(port int, interval time.Duration, messageFn func() *Message) *Broadcaster {
	if interval <= 0 {
		interval = 2500 * time.Millisecond
	}
	return &Broadcaster{
		port:      port,
		interval:  interval,
		messageFn: messageFn,
		stopCh:    make(chan struct{}),
	}
}

// Start launches the background broadcast loop.
func (b *Broadcaster) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return nil
	}
	b.running = true

	go b.loop()
	return nil
}

// Stop shuts down the broadcast loop.
func (b *Broadcaster) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}
	b.running = false
	close(b.stopCh)
}

func (b *Broadcaster) loop() {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	// Initial announce
	b.broadcastOnce()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.broadcastOnce()
		}
	}
}

func (b *Broadcaster) broadcastOnce() {
	if b.messageFn == nil {
		return
	}
	msg := b.messageFn()
	if msg == nil {
		return
	}

	payload, err := msg.Encode()
	if err != nil {
		observability.Warn("Failed to encode discovery announcement: %v", err)
		return
	}

	// 1. Broadcast to 255.255.255.255:port
	dst, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", b.port))
	if err == nil {
		if conn, err := net.DialUDP("udp4", nil, dst); err == nil {
			_, _ = conn.Write(payload)
			_ = conn.Close()
		}
	}

	// 2. Direct broadcast to all active IPv4 broadcast interface addresses
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			// Compute directed broadcast address for subnet
			ip := ipNet.IP.To4()
			mask := ipNet.Mask
			bcast := net.IPv4(
				ip[0]|^mask[0],
				ip[1]|^mask[1],
				ip[2]|^mask[2],
				ip[3]|^mask[3],
			)
			if bcastAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", bcast.String(), b.port)); err == nil {
				if conn, err := net.DialUDP("udp4", nil, bcastAddr); err == nil {
					_, _ = conn.Write(payload)
					_ = conn.Close()
				}
			}
		}
	}
}
