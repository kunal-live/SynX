package security

import (
	"path/filepath"
	"testing"
	"time"

	"synx/internal/storage"
)

func TestTrustStore(t *testing.T) {
	tempDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tempDir, "trust_test.db"))
	if err != nil {
		t.Fatalf("storage.Open failed: %v", err)
	}
	defer db.Close()

	ts := NewTrustStore(db)

	// Validate empty ID
	if err := ts.TrustPeer(TrustedPeer{PeerID: ""}); err == nil {
		t.Errorf("expected error trusting peer with empty ID")
	}

	// Trust valid peer
	peer := TrustedPeer{
		PeerID:    "trusted_01",
		Name:      "Developer Rig",
		PublicKey: "ed25519_pub_key_01",
		Token:     "tok_secure_123",
		TrustedAt: time.Now(),
	}

	if err := ts.TrustPeer(peer); err != nil {
		t.Fatalf("TrustPeer failed: %v", err)
	}

	if !ts.IsTrusted("trusted_01") {
		t.Errorf("expected trusted_01 to be trusted")
	}
	if ts.IsTrusted("unknown_peer") {
		t.Errorf("expected unknown_peer not to be trusted")
	}

	pGot, ok := ts.Get("trusted_01")
	if !ok || pGot.Name != "Developer Rig" {
		t.Errorf("Get failed or returned wrong data: %+v", pGot)
	}

	if !ts.VerifyToken("trusted_01", "tok_secure_123") {
		t.Errorf("expected token verification to succeed")
	}
	if ts.VerifyToken("trusted_01", "wrong_token") {
		t.Errorf("expected token verification with wrong token to fail")
	}

	trustedList := ts.ListTrusted()
	if len(trustedList) != 1 {
		t.Errorf("expected 1 trusted peer, got %d", len(trustedList))
	}

	// Revoke
	if err := ts.RevokePeer("trusted_01"); err != nil {
		t.Fatalf("RevokePeer failed: %v", err)
	}
	if ts.IsTrusted("trusted_01") {
		t.Errorf("expected trusted_01 not to be trusted after revocation")
	}
}
