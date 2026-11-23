package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aim-cli/aim/internal/desktop"
	"github.com/aim-cli/aim/internal/service"
)

type mockDesktopService struct {
	statusFunc func(ctx context.Context) ([]service.DesktopStatusDTO, error)
	launchFunc func(ctx context.Context, req service.DesktopLaunchRequest) error
	quitFunc   func(ctx context.Context, agent string) error
	focusFunc  func(ctx context.Context, agent string) error
}

func (m *mockDesktopService) ListStatus(ctx context.Context) ([]service.DesktopStatusDTO, error) {
	if m.statusFunc != nil {
		return m.statusFunc(ctx)
	}
	return nil, nil
}

func (m *mockDesktopService) Launch(ctx context.Context, req service.DesktopLaunchRequest) error {
	if m.launchFunc != nil {
		return m.launchFunc(ctx, req)
	}
	return nil
}

func (m *mockDesktopService) Quit(ctx context.Context, agent string) error {
	if m.quitFunc != nil {
		return m.quitFunc(ctx, agent)
	}
	return nil
}

func (m *mockDesktopService) Focus(ctx context.Context, agent string) error {
	if m.focusFunc != nil {
		return m.focusFunc(ctx, agent)
	}
	return nil
}

func TestHandleDesktopStatus(t *testing.T) {
	srv := NewServer(nil, nil, nil, 0, false)
	mock := &mockDesktopService{
		statusFunc: func(ctx context.Context) ([]service.DesktopStatusDTO, error) {
			return []service.DesktopStatusDTO{
				{Agent: "cursor", DisplayName: "Cursor", Running: true, PID: 1234, Profile: "work"},
			}, nil
		},
	}
	srv.SetDesktopService(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/desktop/status", nil)
	rr := httptest.NewRecorder()
	srv.setupRoutes()
	srv.handleDesktopStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp []service.DesktopStatusDTO
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].Agent != "cursor" || resp[0].PID != 1234 {
		t.Errorf("unexpected status response: %+v", resp)
	}
}

func TestHandleDesktopLaunch(t *testing.T) {
	srv := NewServer(nil, nil, nil, 0, false)
	mock := &mockDesktopService{
		launchFunc: func(ctx context.Context, req service.DesktopLaunchRequest) error {
			if req.Profile == "conflict" {
				return desktop.ErrInstanceRunning
			}
			return nil
		},
	}
	srv.SetDesktopService(mock)

	// Success case
	body, _ := json.Marshal(service.DesktopLaunchRequest{Agent: "cursor", Profile: "work"})
	req := httptest.NewRequest(http.MethodPost, "/api/desktop/launch", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.handleDesktopLaunch(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Conflict case (HTTP 409)
	bodyConflict, _ := json.Marshal(service.DesktopLaunchRequest{Agent: "cursor", Profile: "conflict"})
	reqConflict := httptest.NewRequest(http.MethodPost, "/api/desktop/launch", bytes.NewReader(bodyConflict))
	rrConflict := httptest.NewRecorder()
	srv.handleDesktopLaunch(rrConflict, reqConflict)

	if rrConflict.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict, got %d: %s", rrConflict.Code, rrConflict.Body.String())
	}

	var conflictResp map[string]any
	if err := json.Unmarshal(rrConflict.Body.Bytes(), &conflictResp); err != nil {
		t.Fatalf("failed to decode conflict response: %v", err)
	}
	if conflictResp["code"] != "INSTANCE_RUNNING" {
		t.Errorf("expected code INSTANCE_RUNNING, got %v", conflictResp["code"])
	}

	// Missing agent / profile validation
	bodyBad, _ := json.Marshal(service.DesktopLaunchRequest{Agent: "", Profile: ""})
	reqBad := httptest.NewRequest(http.MethodPost, "/api/desktop/launch", bytes.NewReader(bodyBad))
	rrBad := httptest.NewRecorder()
	srv.handleDesktopLaunch(rrBad, reqBad)

	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d", rrBad.Code)
	}
}

func TestHandleDesktopQuitAndFocus(t *testing.T) {
	srv := NewServer(nil, nil, nil, 0, false)
	mock := &mockDesktopService{
		quitFunc: func(ctx context.Context, agent string) error {
			if agent == "not-running" {
				return desktop.ErrProcessNotFound
			}
			return nil
		},
		focusFunc: func(ctx context.Context, agent string) error {
			if agent == "not-running" {
				return desktop.ErrProcessNotFound
			}
			return nil
		},
	}
	srv.SetDesktopService(mock)

	// Quit success
	bodyQuit, _ := json.Marshal(map[string]string{"agent": "cursor"})
	reqQuit := httptest.NewRequest(http.MethodPost, "/api/desktop/quit", bytes.NewReader(bodyQuit))
	rrQuit := httptest.NewRecorder()
	srv.handleDesktopQuit(rrQuit, reqQuit)
	if rrQuit.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrQuit.Code)
	}

	// Quit not found (404)
	bodyNotFound, _ := json.Marshal(map[string]string{"agent": "not-running"})
	reqNotFound := httptest.NewRequest(http.MethodPost, "/api/desktop/quit", bytes.NewReader(bodyNotFound))
	rrNotFound := httptest.NewRecorder()
	srv.handleDesktopQuit(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rrNotFound.Code)
	}

	// Focus success
	bodyFocus, _ := json.Marshal(map[string]string{"agent": "cursor"})
	reqFocus := httptest.NewRequest(http.MethodPost, "/api/desktop/focus", bytes.NewReader(bodyFocus))
	rrFocus := httptest.NewRecorder()
	srv.handleDesktopFocus(rrFocus, reqFocus)
	if rrFocus.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rrFocus.Code)
	}

	// Focus not found (404)
	reqFocusNotFound := httptest.NewRequest(http.MethodPost, "/api/desktop/focus", bytes.NewReader(bodyNotFound))
	rrFocusNotFound := httptest.NewRecorder()
	srv.handleDesktopFocus(rrFocusNotFound, reqFocusNotFound)
	if rrFocusNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rrFocusNotFound.Code)
	}
}
