package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// OAuthDiscovery describes discovered authentication capabilities for an MCP endpoint.
type OAuthDiscovery struct {
	AuthURL  string   `json:"authorization_endpoint,omitempty"`
	TokenURL string   `json:"token_endpoint,omitempty"`
	ClientID string   `json:"client_id,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
	AuthType string   `json:"auth_type"` // "oauth" | "bearer" | "none"
}

var reAuthParam = regexp.MustCompile(`([a-zA-Z0-9_-]+)="?([^",]+)"?`)

func parseAuthParams(header string) map[string]string {
	params := make(map[string]string)
	matches := reAuthParam.FindAllStringSubmatch(header, -1)
	for _, m := range matches {
		if len(m) == 3 {
			params[m[1]] = m[2]
		}
	}
	return params
}

// DiscoverEndpointAuth probes an MCP HTTP/SSE endpoint to determine its authentication requirements.
func DiscoverEndpointAuth(ctx context.Context, serverURL string) (*OAuthDiscovery, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, serverURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create probe request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json, text/event-stream")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("probe request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return &OAuthDiscovery{AuthType: "none"}, nil
	}

	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		return &OAuthDiscovery{AuthType: "none"}, nil
	}

	authHeader := resp.Header.Get("WWW-Authenticate")
	params := parseAuthParams(authHeader)

	// 1. Check if resource_metadata or authorization_uri URL is directly provided
	metadataURL := params["resource_metadata"]
	if metadataURL == "" {
		metadataURL = params["authorization_uri"]
	}

	origin := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Candidate URLs to probe
	candidateURLs := []string{}
	if metadataURL != "" {
		candidateURLs = append(candidateURLs, metadataURL)
	}
	candidateURLs = append(candidateURLs,
		origin+"/.well-known/oauth-protected-resource/mcp",
		origin+"/.well-known/oauth-protected-resource",
		origin+"/.well-known/oauth-authorization-server",
		origin+"/.well-known/openid-configuration",
	)

	for _, cand := range candidateURLs {
		disc, err := probeMetadataURL(reqCtx, client, cand)
		if err == nil && disc != nil && disc.AuthURL != "" && disc.TokenURL != "" {
			disc.AuthType = "oauth"
			return disc, nil
		}
	}

	// If 401 has Bearer scheme but no discoverable OAuth endpoints, mark as bearer
	if strings.Contains(strings.ToLower(authHeader), "bearer") {
		return &OAuthDiscovery{AuthType: "bearer"}, nil
	}

	return &OAuthDiscovery{AuthType: "none"}, nil
}

func probeMetadataURL(ctx context.Context, client *http.Client, metaURL string) (*OAuthDiscovery, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	disc := &OAuthDiscovery{}
	if auth, ok := data["authorization_endpoint"].(string); ok {
		disc.AuthURL = auth
	}
	if tok, ok := data["token_endpoint"].(string); ok {
		disc.TokenURL = tok
	}
	if clientID, ok := data["client_id"].(string); ok {
		disc.ClientID = clientID
	}
	if scopes, ok := data["scopes_supported"].([]any); ok {
		for _, s := range scopes {
			if str, ok := s.(string); ok {
				disc.Scopes = append(disc.Scopes, str)
			}
		}
	}

	// If this was an oauth-protected-resource pointing to authorization_servers
	if disc.AuthURL == "" {
		if authServers, ok := data["authorization_servers"].([]any); ok && len(authServers) > 0 {
			if asURL, ok := authServers[0].(string); ok && asURL != "" {
				asURL = strings.TrimRight(asURL, "/")
				// Probe the authorization server
				asCandidate := asURL + "/.well-known/oauth-authorization-server"
				return probeMetadataURL(ctx, client, asCandidate)
			}
		}
	}

	return disc, nil
}

// AuthorizeMCPOAuth executes the RFC 8414/9207 PKCE OAuth loopback handshake.
func AuthorizeMCPOAuth(ctx context.Context, disc *OAuthDiscovery, promptFn func(authURL string)) (map[string]any, error) {
	if disc == nil || disc.AuthURL == "" || disc.TokenURL == "" {
		return nil, errors.New("incomplete OAuth discovery: missing authorization or token endpoint")
	}

	handshakeCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// 1. Generate PKCE verifier and challenge (S256)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random verifier: %w", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// 2. Generate random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// 3. Bind strictly to IPv4 loopback to avoid macOS IPv6 localhost mismatch
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to bind loopback listener: %w", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// 4. Construct Authorization URL
	authURLParsed, err := url.Parse(disc.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth URL: %w", err)
	}
	q := authURLParsed.Query()
	q.Set("response_type", "code")
	clientID := disc.ClientID
	if clientID == "" {
		clientID = "aim"
	}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	if len(disc.Scopes) > 0 {
		q.Set("scope", strings.Join(disc.Scopes, " "))
	}
	authURLParsed.RawQuery = q.Encode()

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		reqState := r.URL.Query().Get("state")
		if reqState != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errChan <- errors.New("state parameter mismatch during OAuth callback")
			return
		}

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			http.Error(w, fmt.Sprintf("Authentication error: %s (%s)", errParam, errDesc), http.StatusBadRequest)
			errChan <- fmt.Errorf("oauth provider returned error: %s (%s)", errParam, errDesc)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errChan <- errors.New("missing authorization code in callback")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>AIM — MCP Authentication Successful</title>
  <style>
    body {
      background-color: #181a1f;
      color: #abb2bf;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100vh;
      margin: 0;
    }
    .card {
      background-color: #21252b;
      border: 1px solid #3e4451;
      border-radius: 8px;
      padding: 2.5rem;
      text-align: center;
      max-width: 440px;
      box-shadow: 0 10px 25px rgba(0,0,0,0.5);
    }
    h2 { color: #98c379; margin-top: 0; font-size: 1.25rem; }
    p { font-size: 0.875rem; line-height: 1.5; color: #abb2bf; }
    .badge {
      display: inline-block;
      margin-top: 1rem;
      font-family: monospace;
      font-size: 0.75rem;
      background: #282c34;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      color: #61afef;
    }
  </style>
</head>
<body>
  <div class="card">
    <h2>Authentication Successful</h2>
    <p>Your session has successfully authorized with the MCP server. You may close this tab and return to your terminal.</p>
    <div class="badge">AIM MCP Client</div>
  </div>
</body>
</html>`))

		codeChan <- code
	})

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(listener)
	}()

	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	// 5. Notify caller with Auth URL
	promptFn(authURLParsed.String())

	// 6. Wait for code or cancellation
	var authCode string
	select {
	case <-handshakeCtx.Done():
		return nil, handshakeCtx.Err()
	case err := <-errChan:
		return nil, err
	case authCode = <-codeChan:
	}

	// 7. Exchange authorization code for tokens
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", authCode)
	tokenForm.Set("redirect_uri", redirectURI)
	tokenForm.Set("client_id", clientID)
	tokenForm.Set("code_verifier", codeVerifier)

	tokenReq, err := http.NewRequestWithContext(handshakeCtx, http.MethodPost, disc.TokenURL, strings.NewReader(tokenForm.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	tokenClient := &http.Client{Timeout: 15 * time.Second}
	tokenResp, err := tokenClient.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer func() { _ = tokenResp.Body.Close() }()

	bodyBytes, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange returned status %d: %s", tokenResp.StatusCode, string(bodyBytes))
	}

	var tokenMap map[string]any
	if err := json.Unmarshal(bodyBytes, &tokenMap); err != nil {
		return nil, fmt.Errorf("failed to parse token response JSON: %w", err)
	}

	return tokenMap, nil
}
