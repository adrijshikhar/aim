package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenExchange(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("expected Content-Type application/x-www-form-urlencoded, got %s", ct)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "mock_access_token_123",
				"refresh_token": "mock_refresh_token_456",
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
			return
		}
		if r.URL.Path == "/userinfo" {
			if auth := r.Header.Get("Authorization"); auth != "Bearer mock_access_token_123" {
				t.Errorf("expected Bearer mock_access_token_123, got %s", auth)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"email": "test@example.com",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID:    "test-client",
		TokenURL:    mockServer.URL + "/token",
		UserInfoURL: mockServer.URL + "/userinfo",
	}

	tok, err := ExchangeCode(context.Background(), cfg, "auth-code", "verifier", "redirect-uri")
	if err != nil {
		t.Fatalf("ExchangeCode error: %v", err)
	}
	if tok.AccessToken != "mock_access_token_123" {
		t.Errorf("expected access token 'mock_access_token_123', got '%s'", tok.AccessToken)
	}
	if tok.RefreshToken != "mock_refresh_token_456" {
		t.Errorf("expected refresh token 'mock_refresh_token_456', got '%s'", tok.RefreshToken)
	}
	if tok.UserEmail != "test@example.com" {
		t.Errorf("expected user email 'test@example.com', got '%s'", tok.UserEmail)
	}
}

func TestTokenExchange_WithClientSecret(t *testing.T) {
	var receivedSecret string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		receivedSecret = r.FormValue("client_secret")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok_sec",
			"expires_in":   1800,
		})
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID:     "client-with-secret",
		ClientSecret: "super-secret-123",
		TokenURL:     mockServer.URL,
	}

	tok, err := ExchangeCode(context.Background(), cfg, "code", "ver", "uri")
	if err != nil {
		t.Fatalf("ExchangeCode error: %v", err)
	}
	if tok.AccessToken != "tok_sec" {
		t.Errorf("expected 'tok_sec', got %s", tok.AccessToken)
	}
	if receivedSecret != "super-secret-123" {
		t.Errorf("expected client_secret 'super-secret-123', got '%s'", receivedSecret)
	}
}

func TestTokenExchange_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID: "client",
		TokenURL: mockServer.URL,
	}

	_, err := ExchangeCode(context.Background(), cfg, "bad-code", "ver", "uri")
	if err == nil {
		t.Fatal("expected error on HTTP 400, got nil")
	}
}

func TestTokenExchange_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json"))
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID: "client",
		TokenURL: mockServer.URL,
	}

	_, err := ExchangeCode(context.Background(), cfg, "code", "ver", "uri")
	if err == nil {
		t.Fatal("expected error on invalid json, got nil")
	}
}

func TestAuthenticate_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := ProviderConfig{
		ClientID: "test-client",
		AuthURL:  "https://example.com/oauth/auth",
		TokenURL: "https://example.com/oauth/token",
	}

	_, err := Authenticate(ctx, cfg, false)
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
}

func TestHandleCallback_Success(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	handler := handleCallback("test-state", codeChan, errChan)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=test-auth-code", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	select {
	case code := <-codeChan:
		if code != "test-auth-code" {
			t.Errorf("expected code 'test-auth-code', got '%s'", code)
		}
	default:
		t.Error("expected code in codeChan, got none")
	}
}

func TestHandleCallback_StateMismatch(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	handler := handleCallback("expected-state", codeChan, errChan)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=wrong-state&code=123", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	select {
	case err := <-errChan:
		if err == nil || err.Error() != "CSRF state mismatch" {
			t.Errorf("expected CSRF state mismatch error, got %v", err)
		}
	default:
		t.Error("expected error in errChan, got none")
	}
}

func TestHandleCallback_AuthError(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	handler := handleCallback("state1", codeChan, errChan)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=state1&error=access_denied", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	select {
	case err := <-errChan:
		if err == nil || err.Error() != "authorization denied: access_denied" {
			t.Errorf("expected authorization denied error, got %v", err)
		}
	default:
		t.Error("expected error in errChan, got none")
	}
}

func TestHandleCallback_MissingCode(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	handler := handleCallback("state1", codeChan, errChan)

	req := httptest.NewRequest(http.MethodGet, "/callback?state=state1", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	select {
	case err := <-errChan:
		if err == nil || err.Error() != "missing authorization code" {
			t.Errorf("expected missing authorization code error, got %v", err)
		}
	default:
		t.Error("expected error in errChan, got none")
	}
}

func TestHandleCallback_NonBlockingOnMultipleRequests(t *testing.T) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	handler := handleCallback("test-state", codeChan, errChan)

	// First request sends code and fills buffer
	req1 := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=first-code", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request must not block even though codeChan buffer is full
	req2 := httptest.NewRequest(http.MethodGet, "/callback?state=test-state&code=second-code", nil)
	rec2 := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec2, req2)
		close(done)
	}()

	select {
	case <-done:
		// Success, didn't block
	case <-time.After(1 * time.Second):
		t.Fatal("handler blocked on repeat callback with full channel")
	}

	// Repeat error request must also not block when errChan is full
	errReq1 := httptest.NewRequest(http.MethodGet, "/callback?state=wrong-state", nil)
	handler.ServeHTTP(httptest.NewRecorder(), errReq1)

	errReq2 := httptest.NewRequest(http.MethodGet, "/callback?state=wrong-state-2", nil)
	errDone := make(chan struct{})
	go func() {
		handler.ServeHTTP(httptest.NewRecorder(), errReq2)
		close(errDone)
	}()

	select {
	case <-errDone:
		// Success, didn't block
	case <-time.After(1 * time.Second):
		t.Fatal("handler blocked on repeat error callback with full channel")
	}
}

func TestTokenExchange_UserInfoFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "valid_token",
				"expires_in":   3600,
			})
			return
		}
		if r.URL.Path == "/userinfo" {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID:    "client",
		TokenURL:    mockServer.URL + "/token",
		UserInfoURL: mockServer.URL + "/userinfo",
	}

	tok, err := ExchangeCode(context.Background(), cfg, "code", "ver", "uri")
	if err != nil {
		t.Fatalf("ExchangeCode should succeed even if UserInfo fails, got: %v", err)
	}
	if tok.AccessToken != "valid_token" {
		t.Errorf("expected access token 'valid_token', got '%s'", tok.AccessToken)
	}
	if tok.UserEmail != "" {
		t.Errorf("expected empty user email on failure, got '%s'", tok.UserEmail)
	}
}

func TestTokenExchange_InvalidUserInfoURL(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "valid_token",
			"expires_in":   3600,
		})
	}))
	defer mockServer.Close()

	cfg := ProviderConfig{
		ClientID:    "client",
		TokenURL:    mockServer.URL,
		UserInfoURL: "://malformed-url\x7f",
	}

	tok, err := ExchangeCode(context.Background(), cfg, "code", "ver", "uri")
	if err != nil {
		t.Fatalf("ExchangeCode should succeed with fallback even on malformed UserInfoURL, got: %v", err)
	}
	if tok.AccessToken != "valid_token" {
		t.Errorf("expected access token 'valid_token', got '%s'", tok.AccessToken)
	}
	if tok.UserEmail != "" {
		t.Errorf("expected empty user email, got '%s'", tok.UserEmail)
	}
}
