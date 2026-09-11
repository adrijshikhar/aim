package oauth

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/oauth2"
)

// GeneratePKCE generates a cryptographically random code verifier and its SHA-256 S256 code challenge.
func GeneratePKCE() (verifier, challenge string, err error) {
	verifier = oauth2.GenerateVerifier()
	challenge = oauth2.S256ChallengeFromVerifier(verifier)
	return verifier, challenge, nil
}

// GenerateRandomState generates a secure random URL-safe string for OAuth state parameter validation.
func GenerateRandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
