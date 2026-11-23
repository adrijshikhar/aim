package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aim-cli/aim/internal/desktop"
	"github.com/aim-cli/aim/internal/service"
)

// handleDesktopStatus handles GET /api/desktop/status to list running state of desktop apps.
func (s *Server) handleDesktopStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	desktopSvc := s.DesktopService()
	if desktopSvc == nil {
		writeJSON(w, http.StatusOK, []service.DesktopStatusDTO{})
		return
	}

	status, err := desktopSvc.ListStatus(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if status == nil {
		status = []service.DesktopStatusDTO{}
	}

	writeJSON(w, http.StatusOK, status)
}

// handleDesktopLaunch handles POST /api/desktop/launch to launch or switch to an isolated desktop app instance.
func (s *Server) handleDesktopLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	desktopSvc := s.DesktopService()
	if desktopSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "desktop service not configured")
		return
	}

	var req service.DesktopLaunchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	req.Agent = strings.TrimSpace(req.Agent)
	req.Profile = strings.TrimSpace(req.Profile)
	if req.Agent == "" {
		writeJSONError(w, http.StatusBadRequest, "agent is required")
		return
	}
	if req.Profile == "" {
		writeJSONError(w, http.StatusBadRequest, "profile is required")
		return
	}

	if err := desktopSvc.Launch(r.Context(), req); err != nil {
		if errors.Is(err, desktop.ErrInstanceRunning) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":   err.Error(),
				"code":    "INSTANCE_RUNNING",
				"agent":   req.Agent,
				"profile": req.Profile,
			})
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "launched"})
}

// handleDesktopQuit handles POST /api/desktop/quit to gracefully close a desktop app.
func (s *Server) handleDesktopQuit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	desktopSvc := s.DesktopService()
	if desktopSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "desktop service not configured")
		return
	}

	var body struct {
		Agent string `json:"agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	agent := strings.TrimSpace(body.Agent)
	if agent == "" {
		writeJSONError(w, http.StatusBadRequest, "agent is required")
		return
	}

	if err := desktopSvc.Quit(r.Context(), agent); err != nil {
		if errors.Is(err, desktop.ErrProcessNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "quit"})
}

// handleDesktopFocus handles POST /api/desktop/focus to bring a running desktop app window to front.
func (s *Server) handleDesktopFocus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	desktopSvc := s.DesktopService()
	if desktopSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "desktop service not configured")
		return
	}

	var body struct {
		Agent string `json:"agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	agent := strings.TrimSpace(body.Agent)
	if agent == "" {
		writeJSONError(w, http.StatusBadRequest, "agent is required")
		return
	}

	if err := desktopSvc.Focus(r.Context(), agent); err != nil {
		if errors.Is(err, desktop.ErrProcessNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "focused"})
}
