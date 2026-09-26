package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	startTime        time.Time
	bytesSent        atomic.Int64
	bytesReceived    atomic.Int64
	activeTransfers  atomic.Int64
	totalTransfers   atomic.Int64
	failedTransfers  atomic.Int64
	completedSuccess atomic.Int64

	mu        sync.RWMutex
	peerCount int
}

var globalMetrics = NewMetrics()

func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func DefaultMetrics() *Metrics {
	return globalMetrics
}

func (m *Metrics) AddBytesSent(n int64)     { m.bytesSent.Add(n) }
func (m *Metrics) AddBytesReceived(n int64) { m.bytesReceived.Add(n) }
func (m *Metrics) IncActiveTransfers()      { m.activeTransfers.Add(1) }
func (m *Metrics) DecActiveTransfers()      { m.activeTransfers.Add(-1) }
func (m *Metrics) IncTotalTransfers()       { m.totalTransfers.Add(1) }
func (m *Metrics) IncFailedTransfers()      { m.failedTransfers.Add(1) }
func (m *Metrics) IncCompletedSuccess()     { m.completedSuccess.Add(1) }

func (m *Metrics) SetPeerCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peerCount = count
}

func (m *Metrics) Snapshot() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"uptime_seconds":    int64(time.Since(m.startTime).Seconds()),
		"bytes_sent":        m.bytesSent.Load(),
		"bytes_received":    m.bytesReceived.Load(),
		"active_transfers":  m.activeTransfers.Load(),
		"total_transfers":   m.totalTransfers.Load(),
		"failed_transfers":  m.failedTransfers.Load(),
		"completed_success": m.completedSuccess.Load(),
		"peer_count":        m.peerCount,
	}
}
