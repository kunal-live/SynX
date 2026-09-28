package observability

import (
	"testing"
)

func TestMetricsTracking(t *testing.T) {
	m := NewMetrics()

	m.SetPeerCount(5)
	if m.GetPeerCount() != 5 {
		t.Errorf("expected peer count 5, got %d", m.GetPeerCount())
	}

	m.AddBytesSent(1024)
	m.AddBytesReceived(2048)
	m.IncActiveTransfers()
	m.IncTotalTransfers()
	m.IncCompletedSuccess()

	snap := m.Snapshot()
	if snap["peer_count"] != 5 {
		t.Errorf("expected snapshot peer_count 5, got %v", snap["peer_count"])
	}
	if snap["bytes_sent"] != int64(1024) {
		t.Errorf("expected bytes_sent 1024, got %v", snap["bytes_sent"])
	}
	if snap["bytes_received"] != int64(2048) {
		t.Errorf("expected bytes_received 2048, got %v", snap["bytes_received"])
	}
	if snap["active_transfers"] != int64(1) {
		t.Errorf("expected active_transfers 1, got %v", snap["active_transfers"])
	}
	if snap["total_transfers"] != int64(1) {
		t.Errorf("expected total_transfers 1, got %v", snap["total_transfers"])
	}
	if snap["completed_success"] != int64(1) {
		t.Errorf("expected completed_success 1, got %v", snap["completed_success"])
	}

	m.DecActiveTransfers()
	if m.Snapshot()["active_transfers"] != int64(0) {
		t.Errorf("expected active_transfers 0 after decrement")
	}

	m.Reset()
	snapReset := m.Snapshot()
	if snapReset["bytes_sent"] != int64(0) || snapReset["peer_count"] != 0 {
		t.Errorf("expected reset values to be 0, got %v", snapReset)
	}
}
