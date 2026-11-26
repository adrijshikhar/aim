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

func TestParseTokenAccountInfo_IDTokenEmail(t *testing.T) {
	// Real-world Antigravity CLI token format
	tokenJSON := `{
		"token": {
			"access_token": "ya29.mock_token",
			"token_type": "Bearer",
			"refresh_token": "1//mock_refresh",
			"expiry": "2026-11-17T14:43:02.208746+05:30",
			"expires_in": 3599
		},
		"auth_method": "consumer",
		"id_token_email": "bhomika1.1.2000@gmail.com"
	}`

	info := ParseTokenAccountInfo([]byte(tokenJSON))
	if info.Email != "bhomika1.1.2000@gmail.com" {
		t.Errorf("expected email 'bhomika1.1.2000@gmail.com', got %q", info.Email)
	}
	if info.AuthMethod != "Google OAuth (consumer)" {
		t.Errorf("expected auth method 'Google OAuth (consumer)', got %q", info.AuthMethod)
	}
}

func TestParseTokenAccountInfo_NestedIDTokenJWT(t *testing.T) {
	claimsJSON := `{"iss":"https://accounts.google.com","sub":"456","email":"nested@example.com","name":"Nested User"}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".mockSignature"

	tokenJSON := `{
		"token": {
			"access_token": "ya29.mock",
			"id_token": "` + mockJWT + `"
		},
		"auth_method": "consumer"
	}`

	info := ParseTokenAccountInfo([]byte(tokenJSON))
	if info.Email != "nested@example.com" {
		t.Errorf("expected email 'nested@example.com', got %q", info.Email)
	}
	if info.Name != "Nested User" {
		t.Errorf("expected name 'Nested User', got %q", info.Name)
	}
	if info.UserID != "456" {
		t.Errorf("expected user_id '456', got %q", info.UserID)
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

func TestParseCodexAuthAccountInfo_JWT(t *testing.T) {
	claimsJSON := `{
		"email": "codex-user@example.com",
		"name": "Codex Engineer",
		"https://api.openai.com/auth": {
			"chatgpt_plan_type": "pro",
			"user_id": "user-12345"
		}
	}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".mockSignature"

	authJSON := `{
		"tokens": {
			"id_token": "` + mockJWT + `",
			"access_token": "mock-access-token"
		}
	}`

	info := ParseCodexAuthAccountInfo([]byte(authJSON))
	if info.Email != "codex-user@example.com" {
		t.Errorf("expected email 'codex-user@example.com', got %q", info.Email)
	}
	if info.Name != "Codex Engineer" {
		t.Errorf("expected name 'Codex Engineer', got %q", info.Name)
	}
	if info.AuthMethod != "ChatGPT Pro" {
		t.Errorf("expected auth method 'ChatGPT Pro', got %q", info.AuthMethod)
	}
}

func TestParseCodexAuthAccountInfo_APIKey(t *testing.T) {
	authJSON := `{
		"openai_api_key": "sk-mock-123456789"
	}`

	info := ParseCodexAuthAccountInfo([]byte(authJSON))
	if info.AuthMethod != "OpenAI API Key" {
		t.Errorf("expected auth method 'OpenAI API Key', got %q", info.AuthMethod)
	}
}

func TestGetProfileAccountInfo_Codex(t *testing.T) {
	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)

	claimsJSON := `{"email":"codex@test.io","name":"Codex Pro","https://api.openai.com/auth":{"chatgpt_plan_type":"plus"}}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".sig"

	authJSON := `{"tokens":{"id_token":"` + mockJWT + `"}}`
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(authJSON), 0600)

	// General extraction
	info := GetProfileAccountInfo(tmpDir)
	if info.Email != "codex@test.io" {
		t.Errorf("expected email 'codex@test.io', got %q", info.Email)
	}
	if info.AuthMethod != "ChatGPT Plus" {
		t.Errorf("expected auth method 'ChatGPT Plus', got %q", info.AuthMethod)
	}

	// Agent-specific extraction
	agentInfo := GetProfileAccountInfoForAgent(tmpDir, "codex")
	if agentInfo.Email != "codex@test.io" {
		t.Errorf("expected agent email 'codex@test.io', got %q", agentInfo.Email)
	}
}

func TestGetProfileAccountInfo_Claude(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	_ = os.MkdirAll(claudeDir, 0700)

	claudeJSON := `{
		"oauthAccount": {
			"emailAddress": "adrij@hevodata.com",
			"fullName": "Adrij Shikhar",
			"organizationName": "Hevo-PED",
			"organizationType": "claude_team",
			"billingType": "stripe_subscription"
		}
	}`
	_ = os.WriteFile(filepath.Join(claudeDir, ".claude.json"), []byte(claudeJSON), 0600)

	// Agent-specific extraction
	info := GetProfileAccountInfoForAgent(tmpDir, "claude")
	if info.Email != "adrij@hevodata.com" {
		t.Errorf("expected claude email 'adrij@hevodata.com', got %q", info.Email)
	}
	if info.Name != "Adrij Shikhar" {
		t.Errorf("expected claude name 'Adrij Shikhar', got %q", info.Name)
	}
	if info.AuthMethod != "Claude Team (Hevo-PED)" {
		t.Errorf("expected auth method 'Claude Team (Hevo-PED)', got %q", info.AuthMethod)
	}
}

func TestGetProfileAccountInfo_AgentIsolation(t *testing.T) {
	// A directory with ONLY Codex auth must NOT return anything when claude is requested
	tmpDir := t.TempDir()
	codexDir := filepath.Join(tmpDir, ".codex")
	_ = os.MkdirAll(codexDir, 0700)

	authJSON := `{"tokens":{"access_token":"mock"},"email":"codex-only@test.io"}`
	_ = os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(authJSON), 0600)

	// Requesting claude must return EMPTY, never Codex's credentials
	claudeInfo := GetProfileAccountInfoForAgent(tmpDir, "claude")
	if claudeInfo.Email != "" || claudeInfo.AuthMethod != "" {
		t.Fatalf("CRITICAL ISOLATION BREACH: claude requested but got codex credentials: %+v", claudeInfo)
	}

	// Requesting agy must return EMPTY
	agyInfo := GetProfileAccountInfoForAgent(tmpDir, "agy")
	if agyInfo.Email != "" || agyInfo.AuthMethod != "" {
		t.Fatalf("CRITICAL ISOLATION BREACH: agy requested but got codex credentials: %+v", agyInfo)
	}

	// Requesting codex must return the credentials
	codexInfo := GetProfileAccountInfoForAgent(tmpDir, "codex")
	if codexInfo.Email != "codex-only@test.io" {
		t.Errorf("expected codex email 'codex-only@test.io', got %q", codexInfo.Email)
	}
}

func TestGetProfileAccountInfo_Agy_RealWorldToken(t *testing.T) {
	tmpDir := t.TempDir()
	agyDir := filepath.Join(tmpDir, ".gemini", "antigravity-cli")
	_ = os.MkdirAll(agyDir, 0700)

	realToken := `{"token":{"access_token":"ya29.mock","token_type":"Bearer","refresh_token":"1//mock","expiry":"2026-11-17T14:43:02+05:30","expires_in":3599},"auth_method":"consumer","id_token_email":"bhomika1.1.2000@gmail.com"}`
	_ = os.WriteFile(filepath.Join(agyDir, "antigravity-oauth-token"), []byte(realToken), 0600)

	info := GetProfileAccountInfoForAgent(tmpDir, "agy")
	if info.Email != "bhomika1.1.2000@gmail.com" {
		t.Errorf("expected agy email 'bhomika1.1.2000@gmail.com', got %q", info.Email)
	}
	if info.AuthMethod != "Google OAuth (consumer)" {
		t.Errorf("expected auth method 'Google OAuth (consumer)', got %q", info.AuthMethod)
	}
}

func TestGetProfileAccountInfo_IgnoresAmbientAnthropicAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-ambient-host-key")
	tmpDir := t.TempDir()
	info := GetProfileAccountInfoForAgent(tmpDir, "claude")
	if info.AuthMethod != "" || info.Email != "" {
		t.Fatalf("LEAK: unauthenticated profile inherited host ANTHROPIC_API_KEY: %+v", info)
	}
}
