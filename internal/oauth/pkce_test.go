package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE returned error: %v", err)
	}
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("verifier length invalid: %d", len(verifier))
	}

	h := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h[:])
	if challenge != expectedChallenge {
		t.Errorf("expected challenge %s, got %s", expectedChallenge, challenge)
	}

	// Verify randomness
	v2, c2, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("second GeneratePKCE returned error: %v", err)
	}
	if verifier == v2 || challenge == c2 {
		t.Errorf("expected distinct PKCE pairs, got identical")
	}
}

func TestGenerateRandomState(t *testing.T) {
	s1, err := GenerateRandomState()
	if err != nil {
		t.Fatalf("GenerateRandomState error: %v", err)
	}
	s2, err := GenerateRandomState()
	if err != nil {
		t.Fatalf("GenerateRandomState error: %v", err)
	}
	if s1 == "" || s2 == "" {
		t.Error("GenerateRandomState returned empty string")
	}
	if s1 == s2 {
		t.Errorf("expected distinct states, got identical: %s", s1)
	}
}
