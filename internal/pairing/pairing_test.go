package pairing

import (
	"testing"
	"time"
)

func TestPairingPIN(t *testing.T) {
	svc := NewService(nil, nil, "local-token")

	pin := svc.GeneratePIN()
	if len(pin) != 6 {
		t.Fatalf("expected 6-digit pin, got %s", pin)
	}

	cur, ok := svc.CurrentPIN()
	if !ok || cur != pin {
		t.Fatalf("CurrentPIN mismatch: cur=%s, ok=%v", cur, ok)
	}

	if svc.VerifyPIN("000000") {
		t.Errorf("wrong pin should fail")
	}

	if !svc.VerifyPIN(pin) {
		t.Errorf("valid pin should succeed")
	}

	// Should be consumed
	if svc.VerifyPIN(pin) {
		t.Errorf("pin should be consumed after successful verification")
	}

	// Expiry test
	svc.GeneratePIN()
	svc.pinExpiry = time.Now().Add(-1 * time.Minute)
	if _, ok := svc.CurrentPIN(); ok {
		t.Errorf("expired pin should not be valid")
	}
}
