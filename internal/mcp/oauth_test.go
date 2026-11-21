package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestDiscoverEndpointAuth_Unprotected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	disc, err := DiscoverEndpointAuth(ctx, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disc.AuthType != "none" {
		t.Errorf("expected AuthType 'none', got %q", disc.AuthType)
	}
}

func TestDiscoverEndpointAuth_BearerDirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Protected API"`)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	disc, err := DiscoverEndpointAuth(ctx, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disc.AuthType != "bearer" {
		t.Errorf("expected AuthType 'bearer', got %q", disc.AuthType)
	}
}

func TestDiscoverEndpointAuth_OAuthMetadataDiscovery(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mcp":
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s/.well-known/oauth-protected-resource/mcp"`, ts.URL))
			w.WriteHeader(http.StatusUnauthorized)
		case "/.well-known/oauth-protected-resource/mcp":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"authorization_servers": []string{ts.URL},
				"scopes_supported":      []string{"read", "write"},
			})
		case "/.well-known/oauth-authorization-server":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"authorization_endpoint": ts.URL + "/oauth/authorize",
				"token_endpoint":         ts.URL + "/oauth/token",
				"scopes_supported":       []string{"read", "write"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	disc, err := DiscoverEndpointAuth(ctx, ts.URL+"/mcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disc.AuthType != "oauth" {
		t.Errorf("expected AuthType 'oauth', got %q", disc.AuthType)
	}
	if disc.AuthURL != ts.URL+"/oauth/authorize" {
		t.Errorf("expected AuthURL %s, got %s", ts.URL+"/oauth/authorize", disc.AuthURL)
	}
	if disc.TokenURL != ts.URL+"/oauth/token" {
		t.Errorf("expected TokenURL %s, got %s", ts.URL+"/oauth/token", disc.TokenURL)
	}
}

func TestAuthorizeMCPOAuth_LoopbackHandshake(t *testing.T) {
	var tokenServer *httptest.Server
	tokenServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" && r.Method == http.MethodPost {
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "authorization_code" {
				http.Error(w, "invalid grant_type", http.StatusBadRequest)
				return
			}
			if r.Form.Get("code") != "auth-code-123" {
				http.Error(w, "invalid code", http.StatusBadRequest)
				return
			}
			if r.Form.Get("code_verifier") == "" {
				http.Error(w, "missing code_verifier", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-token-xyz",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-abc",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer tokenServer.Close()

	disc := &OAuthDiscovery{
		AuthURL:  "https://auth.example.com/oauth/authorize",
		TokenURL: tokenServer.URL + "/token",
		ClientID: "test-client",
		AuthType: "oauth",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	promptCalled := false
	tokens, err := AuthorizeMCPOAuth(ctx, disc, func(authURL string) {
		promptCalled = true
		u, err := url.Parse(authURL)
		if err != nil {
			t.Errorf("failed to parse authURL: %v", err)
			return
		}
		redirectURI := u.Query().Get("redirect_uri")
		state := u.Query().Get("state")

		// Simulate browser redirect back to loopback listener
		callbackURL := fmt.Sprintf("%s?code=auth-code-123&state=%s", redirectURI, state)
		go func() {
			time.Sleep(50 * time.Millisecond)
			resp, err := http.Get(callbackURL)
			if err != nil {
				t.Errorf("callback GET failed: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
	})

	if err != nil {
		t.Fatalf("AuthorizeMCPOAuth failed: %v", err)
	}
	if !promptCalled {
		t.Errorf("expected promptFn to be called")
	}
	if tokens["access_token"] != "access-token-xyz" {
		t.Errorf("expected access_token 'access-token-xyz', got %v", tokens["access_token"])
	}
	if tokens["refresh_token"] != "refresh-token-abc" {
		t.Errorf("expected refresh_token 'refresh-token-abc', got %v", tokens["refresh_token"])
	}
}

func TestAuthorizeMCPOAuth_ContextCancellation(t *testing.T) {
	disc := &OAuthDiscovery{
		AuthURL:  "https://auth.example.com/oauth/authorize",
		TokenURL: "https://auth.example.com/oauth/token",
		AuthType: "oauth",
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately when prompt is called
	_, err := AuthorizeMCPOAuth(ctx, disc, func(authURL string) {
		cancel()
	})

	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}
