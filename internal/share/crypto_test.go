package share

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCrypto_SignAndVerifyLease(t *testing.T) {
	tempDir := t.TempDir()
	pubKey, privKey, err := EnsureNodeKeypair(tempDir)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	leaseID := "ls_test123"
	lenderID := "alice-mbp"
	borrowerID := "bob-mbp"
	profileName := "work"
	agent := "codex"
	expiresAt := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	sig := SignLease(privKey, leaseID, lenderID, borrowerID, profileName, agent, expiresAt)
	if sig == "" {
		t.Fatalf("expected non-empty signature")
	}

	valid := VerifyLeaseSignature(pubKey, leaseID, lenderID, borrowerID, profileName, agent, expiresAt, sig)
	if !valid {
		t.Errorf("expected signature to verify successfully")
	}

	// Test tampering detection
	tamperedExp := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	if VerifyLeaseSignature(pubKey, leaseID, lenderID, borrowerID, profileName, agent, tamperedExp, sig) {
		t.Errorf("tampered expiration must fail verification!")
	}
}

func TestCrypto_EnsureNodeKeypairPersistenceAndPermissions(t *testing.T) {
	tempDir := t.TempDir()
	pubKey1, privKey1, err := EnsureNodeKeypair(tempDir)
	if err != nil {
		t.Fatalf("first EnsureNodeKeypair failed: %v", err)
	}

	keyPath := filepath.Join(tempDir, "node.key")
	pubPath := filepath.Join(tempDir, "node.pub")

	fiKey, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("node.key not found: %v", err)
	}
	if fiKey.Mode().Perm() != 0600 {
		t.Errorf("expected node.key permissions 0600, got %o", fiKey.Mode().Perm())
	}

	fiPub, err := os.Stat(pubPath)
	if err != nil {
		t.Fatalf("node.pub not found: %v", err)
	}
	if fiPub.Mode().Perm() != 0644 {
		t.Errorf("expected node.pub permissions 0644, got %o", fiPub.Mode().Perm())
	}

	// Subsequent call should return identical keys
	pubKey2, privKey2, err := EnsureNodeKeypair(tempDir)
	if err != nil {
		t.Fatalf("second EnsureNodeKeypair failed: %v", err)
	}

	if !pubKey1.Equal(pubKey2) {
		t.Errorf("expected public keys to match on reload")
	}
	if !privKey1.Equal(privKey2) {
		t.Errorf("expected private keys to match on reload")
	}
}

func TestCrypto_PeerInfoPublicKeyJSON(t *testing.T) {
	tempDir := t.TempDir()
	pubKey, _, err := EnsureNodeKeypair(tempDir)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	peer := PeerInfo{
		ID:        "node-1",
		HostName:  "node-1.local",
		PublicKey: pubKey,
	}

	// 1. Marshal writes hex string
	data, err := json.Marshal(peer)
	if err != nil {
		t.Fatalf("failed to marshal peer: %v", err)
	}
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"public_key":`) {
		t.Errorf("expected public_key in json output: %s", jsonStr)
	}

	// 2. Unmarshal parses hex string into ed25519.PublicKey
	var decoded PeerInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal peer: %v", err)
	}
	if !decoded.PublicKey.Equal(pubKey) {
		t.Errorf("expected unmarshaled public key to match original")
	}

	// 3. Unmarshal parses base64-encoded string
	b64PubKey := base64.StdEncoding.EncodeToString(pubKey)
	rawJSON := fmt.Sprintf(`{"id":"node-b64","hostname":"test","public_key":"%s"}`, b64PubKey)
	var decodedB64 PeerInfo
	if err := json.Unmarshal([]byte(rawJSON), &decodedB64); err != nil {
		t.Fatalf("failed to unmarshal base64 public key: %v", err)
	}
	if !decodedB64.PublicKey.Equal(pubKey) {
		t.Errorf("expected base64 unmarshaled public key to match original")
	}
}

func TestCrypto_SignAndVerifyRevocation(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	leaseID := "ls_test_revoke_123"
	lenderID := "lender.node.local"

	sig := SignRevocation(priv, leaseID, lenderID)
	if sig == "" {
		t.Fatalf("expected non-empty revocation signature")
	}

	// 1. Valid revocation signature
	if !VerifyRevocationSignature(pub, leaseID, lenderID, sig) {
		t.Errorf("expected valid revocation signature to pass verification")
	}

	// 2. Tampered leaseID
	if VerifyRevocationSignature(pub, "ls_different_id", lenderID, sig) {
		t.Errorf("expected tampered leaseID to fail verification")
	}

	// 3. Tampered lenderID
	if VerifyRevocationSignature(pub, leaseID, "attacker.node.local", sig) {
		t.Errorf("expected tampered lenderID to fail verification")
	}

	// 4. Corrupted signature hex
	if VerifyRevocationSignature(pub, leaseID, lenderID, "invalid_hex") {
		t.Errorf("expected corrupt hex to fail verification")
	}

	// 5. Short or invalid public key
	if VerifyRevocationSignature([]byte("short"), leaseID, lenderID, sig) {
		t.Errorf("expected short pubkey to fail verification")
	}

	// 6. Invalid private key length in SignRevocation
	if badSig := SignRevocation([]byte("bad_priv"), leaseID, lenderID); badSig != "" {
		t.Errorf("expected empty signature on invalid private key, got: %s", badSig)
	}
}
