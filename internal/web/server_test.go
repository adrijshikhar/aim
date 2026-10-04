package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aim-cli/aim/internal/service"
	"github.com/aim-cli/aim/internal/web"
)

// mockProfileService is a mock implementation of service.ProfileService.
type mockProfileService struct {
	listProfilesFn  func(ctx context.Context, agent string) ([]service.ProfileDTO, error)
	createProfileFn func(ctx context.Context, req service.CreateProfileRequest) (*service.ProfileDTO, error)
	removeProfileFn func(ctx context.Context, agent, name string) error
	renameProfileFn func(ctx context.Context, agent, oldName, newName string) error
}

func (m *mockProfileService) ListProfiles(ctx context.Context, agent string) ([]service.ProfileDTO, error) {
	if m.listProfilesFn != nil {
		return m.listProfilesFn(ctx, agent)
	}
	return []service.ProfileDTO{}, nil
}

func (m *mockProfileService) CreateProfile(ctx context.Context, req service.CreateProfileRequest) (*service.ProfileDTO, error) {
	if m.createProfileFn != nil {
		return m.createProfileFn(ctx, req)
	}
	return &service.ProfileDTO{Name: req.Name, Agent: req.Agent}, nil
}

func (m *mockProfileService) RemoveProfile(ctx context.Context, agent, name string) error {
	if m.removeProfileFn != nil {
		return m.removeProfileFn(ctx, agent, name)
	}
	return nil
}

func (m *mockProfileService) RenameProfile(ctx context.Context, agent, oldName, newName string) error {
	if m.renameProfileFn != nil {
		return m.renameProfileFn(ctx, agent, oldName, newName)
	}
	return nil
}

func (m *mockProfileService) UpdateProfileConfig(ctx context.Context, name string, mcpGlobal, pluginsGlobal *bool) error {
	return nil
}

// mockSessionService is a mock implementation of service.SessionService.
type mockSessionService struct {
	listSessionsFn            func(ctx context.Context, filter service.SessionFilter) ([]service.SessionDTO, error)
	resumeSessionInTerminalFn func(ctx context.Context, req service.ResumeRequest) error
}

func (m *mockSessionService) ListSessions(ctx context.Context, filter service.SessionFilter) ([]service.SessionDTO, error) {
	if m.listSessionsFn != nil {
		return m.listSessionsFn(ctx, filter)
	}
	return []service.SessionDTO{}, nil
}

func (m *mockSessionService) ResumeSessionInTerminal(ctx context.Context, req service.ResumeRequest) error {
	if m.resumeSessionInTerminalFn != nil {
		return m.resumeSessionInTerminalFn(ctx, req)
	}
	return nil
}

// mockLauncherService is a mock implementation of service.LauncherService.
type mockLauncherService struct {
	launchTerminalFn     func(ctx context.Context, cmdStr string) error
	availableTerminalsFn func() []string
}

func (m *mockLauncherService) LaunchTerminal(ctx context.Context, cmdStr string) error {
	if m.launchTerminalFn != nil {
		return m.launchTerminalFn(ctx, cmdStr)
	}
	return nil
}

func (m *mockLauncherService) AvailableTerminals() []string {
	if m.availableTerminalsFn != nil {
		return m.availableTerminalsFn()
	}
	return []string{"Ghostty", "Terminal"}
}

func setupTestServer(t *testing.T, profs *mockProfileService, sess *mockSessionService, launch *mockLauncherService, devMode bool) (*httptest.Server, *web.Server) {
	t.Helper()
	if profs == nil {
		profs = &mockProfileService{}
	}
	if sess == nil {
		sess = &mockSessionService{}
	}
	if launch == nil {
		launch = &mockLauncherService{}
	}

	srv := web.NewServer(profs, sess, launch, 0, devMode)
	srv.SetVersion("0.13.0-test")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, srv
}

func TestGetProfiles(t *testing.T) {
	profs := &mockProfileService{
		listProfilesFn: func(ctx context.Context, agent string) ([]service.ProfileDTO, error) {
			if agent == "codex" {
				return []service.ProfileDTO{
					{Name: "work", Agent: "codex", HasCredentials: true},
				}, nil
			}
			return []service.ProfileDTO{
				{Name: "work", Agent: "codex", HasCredentials: true},
				{Name: "personal", Agent: "claude", HasCredentials: false},
			}, nil
		},
	}

	ts, _ := setupTestServer(t, profs, nil, nil, false)

	// Test GET /api/profiles without filter
	resp, err := http.Get(ts.URL + "/api/profiles")
	if err != nil {
		t.Fatalf("GET /api/profiles failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var list []service.ProfileDTO
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(list))
	}

	// Test GET /api/profiles with ?agent=codex filter
	respFiltered, err := http.Get(ts.URL + "/api/profiles?agent=codex")
	if err != nil {
		t.Fatalf("GET /api/profiles?agent=codex failed: %v", err)
	}
	defer respFiltered.Body.Close()

	if respFiltered.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", respFiltered.StatusCode)
	}

	var filteredList []service.ProfileDTO
	if err := json.NewDecoder(respFiltered.Body).Decode(&filteredList); err != nil {
		t.Fatalf("failed to decode filtered response: %v", err)
	}
	if len(filteredList) != 1 || filteredList[0].Agent != "codex" {
		t.Fatalf("expected 1 codex profile, got %v", filteredList)
	}
}

func TestGetProfiles_EmptyAndError(t *testing.T) {
	// Empty list should return [] instead of null
	profs := &mockProfileService{
		listProfilesFn: func(ctx context.Context, agent string) ([]service.ProfileDTO, error) {
			return nil, nil
		},
	}
	ts, _ := setupTestServer(t, profs, nil, nil, false)

	resp, err := http.Get(ts.URL + "/api/profiles")
	if err != nil {
		t.Fatalf("GET /api/profiles failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if string(bytes.TrimSpace(bodyBytes)) != "[]" {
		t.Fatalf("expected '[]', got %q", string(bodyBytes))
	}

	// Service error should return 500
	profsErr := &mockProfileService{
		listProfilesFn: func(ctx context.Context, agent string) ([]service.ProfileDTO, error) {
			return nil, errors.New("db error")
		},
	}
	tsErr, _ := setupTestServer(t, profsErr, nil, nil, false)
	respErr, err := http.Get(tsErr.URL + "/api/profiles")
	if err != nil {
		t.Fatalf("GET /api/profiles failed: %v", err)
	}
	defer respErr.Body.Close()

	if respErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", respErr.StatusCode)
	}
}

func TestCreateProfile(t *testing.T) {
	profs := &mockProfileService{
		createProfileFn: func(ctx context.Context, req service.CreateProfileRequest) (*service.ProfileDTO, error) {
			if req.Name == "existing" {
				return nil, errors.New("profile already exists")
			}
			return &service.ProfileDTO{
				Name:           req.Name,
				Agent:          req.Agent,
				HasCredentials: false,
			}, nil
		},
	}

	ts, _ := setupTestServer(t, profs, nil, nil, false)

	// Valid profile creation
	reqBody := service.CreateProfileRequest{
		Agent: "claude",
		Name:  "testing",
	}
	b, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/api/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/profiles failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}

	var created service.ProfileDTO
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.Name != "testing" || created.Agent != "claude" {
		t.Fatalf("unexpected profile: %+v", created)
	}

	// Missing profile name
	badReq := service.CreateProfileRequest{Agent: "claude"}
	badBytes, _ := json.Marshal(badReq)
	badResp, err := http.Post(ts.URL+"/api/profiles", "application/json", bytes.NewReader(badBytes))
	if err != nil {
		t.Fatalf("POST /api/profiles failed: %v", err)
	}
	defer badResp.Body.Close()
	if badResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", badResp.StatusCode)
	}

	// Service error on conflict
	conflictReq := service.CreateProfileRequest{Agent: "claude", Name: "existing"}
	conflictBytes, _ := json.Marshal(conflictReq)
	conflictResp, err := http.Post(ts.URL+"/api/profiles", "application/json", bytes.NewReader(conflictBytes))
	if err != nil {
		t.Fatalf("POST /api/profiles failed: %v", err)
	}
	defer conflictResp.Body.Close()
	if conflictResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for existing profile, got %d", conflictResp.StatusCode)
	}
}

func TestDeleteProfile(t *testing.T) {
	deletedAgent := ""
	deletedName := ""

	profs := &mockProfileService{
		removeProfileFn: func(ctx context.Context, agent, name string) error {
			if name == "fail" {
				return errors.New("cannot delete profile")
			}
			deletedAgent = agent
			deletedName = name
			return nil
		},
	}

	ts, _ := setupTestServer(t, profs, nil, nil, false)

	// Valid DELETE /api/profiles/{agent}/{name}
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/profiles/codex/test-prof", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /api/profiles/codex/test-prof failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if deletedAgent != "codex" || deletedName != "test-prof" {
		t.Fatalf("expected codex / test-prof, got %s / %s", deletedAgent, deletedName)
	}

	// DELETE failure
	reqFail, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/profiles/codex/fail", nil)
	respFail, err := http.DefaultClient.Do(reqFail)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer respFail.Body.Close()
	if respFail.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", respFail.StatusCode)
	}
}

func TestGetSessions(t *testing.T) {
	now := time.Now()
	var receivedFilter service.SessionFilter

	sess := &mockSessionService{
		listSessionsFn: func(ctx context.Context, filter service.SessionFilter) ([]service.SessionDTO, error) {
			receivedFilter = filter
			return []service.SessionDTO{
				{
					ID:        "sess-1",
					Agent:     "codex",
					Profile:   "work",
					Title:     "Refactor auth",
					Turns:     12,
					UpdatedAt: now,
					IsActive:  true,
				},
			}, nil
		},
	}

	ts, _ := setupTestServer(t, nil, sess, nil, false)

	// GET /api/sessions with query params
	url := ts.URL + "/api/sessions?agent=codex&profile=work&q=auth&limit=10"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET /api/sessions failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var results []service.SessionDTO
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(results) != 1 || results[0].ID != "sess-1" {
		t.Fatalf("unexpected sessions: %+v", results)
	}

	if receivedFilter.Agent != "codex" || receivedFilter.Profile != "work" || receivedFilter.Query != "auth" || receivedFilter.Limit != 10 {
		t.Fatalf("filter mismatch: %+v", receivedFilter)
	}
}

func TestResumeSession(t *testing.T) {
	var resumedReq service.ResumeRequest

	sess := &mockSessionService{
		resumeSessionInTerminalFn: func(ctx context.Context, req service.ResumeRequest) error {
			if req.SessionID == "fail" {
				return errors.New("terminal launch failed")
			}
			resumedReq = req
			return nil
		},
	}
	launch := &mockLauncherService{
		availableTerminalsFn: func() []string {
			return []string{"Ghostty", "Terminal"}
		},
	}

	ts, _ := setupTestServer(t, nil, sess, launch, false)

	// Valid resume
	body := service.ResumeRequest{
		Agent:     "claude",
		Profile:   "work",
		SessionID: "sess-abc-123",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/sessions/resume", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/sessions/resume failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var respData map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if respData["status"] != "launched" || respData["terminal"] != "Ghostty" {
		t.Fatalf("unexpected response: %+v", respData)
	}
	if resumedReq.SessionID != "sess-abc-123" {
		t.Fatalf("unexpected resume request: %+v", resumedReq)
	}

	// Missing parameters
	badBody := service.ResumeRequest{SessionID: "sess-abc"}
	bb, _ := json.Marshal(badBody)
	badResp, err := http.Post(ts.URL+"/api/sessions/resume", "application/json", bytes.NewReader(bb))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer badResp.Body.Close()
	if badResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", badResp.StatusCode)
	}
}

func TestGetStatus(t *testing.T) {
	ts, _ := setupTestServer(t, nil, nil, nil, false)

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("GET /api/status failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var status web.StatusDTO
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}
	if status.Version != "0.13.0-test" {
		t.Errorf("expected version 0.13.0-test, got %q", status.Version)
	}
	if status.OS == "" || status.Arch == "" || status.Uptime == "" {
		t.Errorf("expected non-empty OS, Arch, and Uptime, got: %+v", status)
	}
}

func TestSPAAndFallback(t *testing.T) {
	ts, _ := setupTestServer(t, nil, nil, nil, false)

	// Root path / serves index.html
	respRoot, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer respRoot.Body.Close()
	if respRoot.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /, got %d", respRoot.StatusCode)
	}
	rootBody, _ := io.ReadAll(respRoot.Body)
	if !bytes.Contains(rootBody, []byte("<div id=\"root\">")) || !bytes.Contains(rootBody, []byte("AIM")) {
		t.Fatalf("expected index.html content, got: %s", string(rootBody))
	}

	// Client-side route /sessions falls back to index.html
	respSessions, err := http.Get(ts.URL + "/sessions")
	if err != nil {
		t.Fatalf("GET /sessions failed: %v", err)
	}
	defer respSessions.Body.Close()
	if respSessions.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /sessions, got %d", respSessions.StatusCode)
	}
	sessionsBody, _ := io.ReadAll(respSessions.Body)
	if !bytes.Contains(sessionsBody, []byte("<div id=\"root\">")) || !bytes.Contains(sessionsBody, []byte("AIM")) {
		t.Fatalf("expected index.html fallback for client-side route, got: %s", string(sessionsBody))
	}

	// Non-GET method to SPA path returns 405
	respPost, err := http.Post(ts.URL+"/sessions", "text/plain", nil)
	if err != nil {
		t.Fatalf("POST /sessions failed: %v", err)
	}
	defer respPost.Body.Close()
	if respPost.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", respPost.StatusCode)
	}

	// Non-existent API route returns 404 JSON, NOT index.html!
	respAPI404, err := http.Get(ts.URL + "/api/nonexistent")
	if err != nil {
		t.Fatalf("GET /api/nonexistent failed: %v", err)
	}
	defer respAPI404.Body.Close()
	if respAPI404.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for missing API route, got %d", respAPI404.StatusCode)
	}
	apiBody, _ := io.ReadAll(respAPI404.Body)
	if !bytes.Contains(apiBody, []byte("endpoint not found")) {
		t.Fatalf("expected JSON error for missing API route, got: %s", string(apiBody))
	}
}

func TestDevModeCORS(t *testing.T) {
	ts, _ := setupTestServer(t, nil, nil, nil, true)

	// OPTIONS preflight request
	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/api/profiles", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS /api/profiles failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS in devMode, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("expected Access-Control-Allow-Methods header")
	}

	// GET request has CORS header in devMode
	respGet, err := http.Get(ts.URL + "/api/profiles")
	if err != nil {
		t.Fatalf("GET /api/profiles failed: %v", err)
	}
	defer respGet.Body.Close()
	if respGet.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header on GET in devMode, got %q", respGet.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestServerLifecycle(t *testing.T) {
	srv := web.NewServer(nil, nil, nil, 0, false)
	srv.SetVersion("1.0.0")

	if err := srv.Listen(); err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	port := srv.Port()
	if port <= 0 {
		t.Fatalf("expected positive port, got %d", port)
	}

	if srv.URL() == "" {
		t.Fatalf("expected non-empty URL")
	}

	go func() {
		_ = srv.Serve()
	}()

	// Query running server
	resp, err := http.Get(srv.URL() + "/api/status")
	if err != nil {
		t.Fatalf("failed to query running server: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
}

func TestNilServicesHandling(t *testing.T) {
	srv := web.NewServer(nil, nil, nil, 0, false)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	// All endpoints return 500 cleanly when underlying services are nil
	resp, err := http.Get(ts.URL + "/api/profiles")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil profiles, got %d", resp.StatusCode)
	}

	b, _ := json.Marshal(service.CreateProfileRequest{Name: "test"})
	respPost, err := http.Post(ts.URL+"/api/profiles", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer respPost.Body.Close()
	if respPost.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil profiles, got %d", respPost.StatusCode)
	}

	reqDel, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/profiles/codex/test", nil)
	respDel, err := http.DefaultClient.Do(reqDel)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	defer respDel.Body.Close()
	if respDel.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil profiles, got %d", respDel.StatusCode)
	}

	respSess, err := http.Get(ts.URL + "/api/sessions")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer respSess.Body.Close()
	if respSess.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil sessions, got %d", respSess.StatusCode)
	}

	bRes, _ := json.Marshal(service.ResumeRequest{Agent: "claude", SessionID: "123"})
	respRes, err := http.Post(ts.URL+"/api/sessions/resume", "application/json", bytes.NewReader(bRes))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer respRes.Body.Close()
	if respRes.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 on nil sessions, got %d", respRes.StatusCode)
	}
}

type mockMCPService struct {
	listServersFn func(ctx context.Context, profileName string) ([]service.MCPServerDTO, error)
}

func (m *mockMCPService) ListServers(ctx context.Context, profileName string) ([]service.MCPServerDTO, error) {
	if m.listServersFn != nil {
		return m.listServersFn(ctx, profileName)
	}
	return nil, nil
}

func TestGetMcpServers(t *testing.T) {
	mcp := &mockMCPService{
		listServersFn: func(ctx context.Context, profileName string) ([]service.MCPServerDTO, error) {
			if profileName == "isolated" {
				return []service.MCPServerDTO{
					{Name: "local-tool", Command: "node", Args: []string{"server.js"}, Scope: "profile"},
				}, nil
			}
			return []service.MCPServerDTO{
				{Name: "claude-mem", Command: "node", Args: []string{"mem.js"}, Scope: "global"},
				{Name: "playwright", Command: "npx", Args: []string{"playwright"}, Scope: "global"},
			}, nil
		},
	}

	srv := web.NewServer(nil, nil, nil, 0, false, mcp)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	// GET /api/mcp
	resp, err := http.Get(ts.URL + "/api/mcp")
	if err != nil {
		t.Fatalf("GET /api/mcp failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var servers []service.MCPServerDTO
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	// GET /api/mcp?profile=isolated
	respIso, err := http.Get(ts.URL + "/api/mcp?profile=isolated")
	if err != nil {
		t.Fatalf("GET /api/mcp?profile=isolated failed: %v", err)
	}
	defer respIso.Body.Close()

	var isoServers []service.MCPServerDTO
	if err := json.NewDecoder(respIso.Body).Decode(&isoServers); err != nil {
		t.Fatalf("failed to decode iso response: %v", err)
	}
	if len(isoServers) != 1 || isoServers[0].Name != "local-tool" {
		t.Fatalf("expected 1 local-tool server, got %v", isoServers)
	}
}
