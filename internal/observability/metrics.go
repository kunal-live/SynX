package observability

import (
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
	peerCount        atomic.Int64
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
	m.peerCount.Store(int64(count))
}

func (m *Metrics) GetPeerCount() int {
	return int(m.peerCount.Load())
}

func (m *Metrics) Reset() {
	m.bytesSent.Store(0)
	m.bytesReceived.Store(0)
	m.activeTransfers.Store(0)
	m.totalTransfers.Store(0)
	m.failedTransfers.Store(0)
	m.completedSuccess.Store(0)
	m.peerCount.Store(0)
	m.startTime = time.Now()
}

func (m *Metrics) Snapshot() map[string]any {
	return map[string]any{
		"uptime_seconds":    int64(time.Since(m.startTime).Seconds()),
		"bytes_sent":        m.bytesSent.Load(),
		"bytes_received":    m.bytesReceived.Load(),
		"active_transfers":  m.activeTransfers.Load(),
		"total_transfers":   m.totalTransfers.Load(),
		"failed_transfers":  m.failedTransfers.Load(),
		"completed_success": m.completedSuccess.Load(),
		"peer_count":        int(m.peerCount.Load()),
	}
}

