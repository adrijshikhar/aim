package share

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureNodeKeypair loads or generates the Ed25519 node keypair in aimDir.
// node.key is written with 0600 permissions, node.pub with 0644 permissions.
// Atomic disk writes are enforced to prevent partial file corruption.
func EnsureNodeKeypair(aimDir string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if aimDir == "" {
		return nil, nil, fmt.Errorf("aimDir cannot be empty")
	}

	keyPath := filepath.Join(aimDir, "node.key")
	pubPath := filepath.Join(aimDir, "node.pub")

	if keyData, err := os.ReadFile(keyPath); err == nil && len(keyData) == ed25519.PrivateKeySize {
		pubData, _ := os.ReadFile(pubPath)
		if len(pubData) == ed25519.PublicKeySize {
			return ed25519.PublicKey(pubData), ed25519.PrivateKey(keyData), nil
		}
		// If pub file is missing or corrupt, derive public key from private key
		pub := ed25519.PrivateKey(keyData).Public().(ed25519.PublicKey)
		_ = os.WriteFile(pubPath, pub, 0644)
		return pub, ed25519.PrivateKey(keyData), nil
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate node keypair: %w", err)
	}

	if err := os.MkdirAll(aimDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("failed to create aim directory: %w", err)
	}

	pid := os.Getpid()
	tmpKey := fmt.Sprintf("%s.tmp.%d", keyPath, pid)
	if err := os.WriteFile(tmpKey, priv, 0600); err != nil {
		return nil, nil, err
	}
	if err := os.Rename(tmpKey, keyPath); err != nil {
		_ = os.Remove(tmpKey)
		return nil, nil, err
	}

	tmpPub := fmt.Sprintf("%s.tmp.%d", pubPath, pid)
	if err := os.WriteFile(tmpPub, pub, 0644); err != nil {
		return nil, nil, err
	}
	if err := os.Rename(tmpPub, pubPath); err != nil {
		_ = os.Remove(tmpPub)
		return nil, nil, err
	}

	return pub, priv, nil
}

// CanonicalLeaseString formats the lease parameters into a deterministic pipe-delimited manifest string.
func CanonicalLeaseString(leaseID, lenderID, borrowerID, profileName, agent, expiresAt string) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", leaseID, lenderID, borrowerID, profileName, agent, expiresAt)
}

// SignLease signs the canonical lease manifest using the lender's private key and returns a hex string.
func SignLease(privKey ed25519.PrivateKey, leaseID, lenderID, borrowerID, profileName, agent, expiresAt string) string {
	if len(privKey) != ed25519.PrivateKeySize {
		return ""
	}
	canonical := CanonicalLeaseString(leaseID, lenderID, borrowerID, profileName, agent, expiresAt)
	sig := ed25519.Sign(privKey, []byte(canonical))
	return hex.EncodeToString(sig)
}

// VerifyLeaseSignature verifies a hex-encoded Ed25519 signature against the canonical lease manifest.
func VerifyLeaseSignature(pubKey ed25519.PublicKey, leaseID, lenderID, borrowerID, profileName, agent, expiresAt, signatureHex string) bool {
	if len(pubKey) != ed25519.PublicKeySize {
		return false
	}
	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}
	canonical := CanonicalLeaseString(leaseID, lenderID, borrowerID, profileName, agent, expiresAt)
	return ed25519.Verify(pubKey, []byte(canonical), sigBytes)
}
