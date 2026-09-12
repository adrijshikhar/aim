package profile

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTokenAccountInfo_JWT(t *testing.T) {
	claimsJSON := `{"iss":"https://accounts.google.com","sub":"123","email":"testuser@example.com","name":"Test User"}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".mockSignature"

	tokenJSON := `{
		"token": {
			"access_token": "ya29.mock"
		},
		"auth_method": "consumer",
		"id_token": "` + mockJWT + `"
	}`

	info := ParseTokenAccountInfo([]byte(tokenJSON))
	if info.Email != "testuser@example.com" {
		t.Errorf("expected email 'testuser@example.com', got %q", info.Email)
	}
	if info.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %q", info.Name)
	}
	if info.AuthMethod != "Google OAuth (consumer)" {
		t.Errorf("expected auth method 'Google OAuth (consumer)', got %q", info.AuthMethod)
	}
}

func TestParseTokenAccountInfo_FallbackEmail(t *testing.T) {
	tokenJSON := `{
		"token": {
			"access_token": "ya29.mock"
		},
		"user_email": "direct@example.com",
		"auth_method": "Workplace"
	}`

	info := ParseTokenAccountInfo([]byte(tokenJSON))
	if info.Email != "direct@example.com" {
		t.Errorf("expected email 'direct@example.com', got %q", info.Email)
	}
	if info.AuthMethod != "Workplace" {
		t.Errorf("expected auth method 'Workplace', got %q", info.AuthMethod)
	}
}

func TestGetProfileAccountInfo_ADC(t *testing.T) {
	tmpDir := t.TempDir()
	adcPath := filepath.Join(tmpDir, ".config", "gcloud", "application_default_credentials.json")
	_ = os.MkdirAll(filepath.Dir(adcPath), 0700)
	adcJSON := `{
		"client_email": "sa@project.iam.gserviceaccount.com",
		"type": "service_account",
		"project_id": "test-gcp-project"
	}`
	_ = os.WriteFile(adcPath, []byte(adcJSON), 0600)

	info := GetProfileAccountInfo(tmpDir)
	if info.Email != "sa@project.iam.gserviceaccount.com" {
		t.Errorf("expected ADC email, got %q", info.Email)
	}
	if info.AuthMethod != "Google ADC (service_account)" {
		t.Errorf("expected auth method 'Google ADC (service_account)', got %q", info.AuthMethod)
	}
	if info.ProjectID != "test-gcp-project" {
		t.Errorf("expected project_id 'test-gcp-project', got %q", info.ProjectID)
	}
}

func TestGetProfileAccountInfo_EmptyOrMissing(t *testing.T) {
	info := GetProfileAccountInfo("")
	if info.Email != "" || info.Name != "" || info.AuthMethod != "" {
		t.Errorf("expected empty AccountInfo for empty dir, got %+v", info)
	}

	emptyDir := t.TempDir()
	infoEmpty := GetProfileAccountInfo(emptyDir)
	if infoEmpty.Email != "" || infoEmpty.Name != "" || infoEmpty.AuthMethod != "" {
		t.Errorf("expected empty AccountInfo for empty profile, got %+v", infoEmpty)
	}
}
