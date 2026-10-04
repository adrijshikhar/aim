package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetSessions handles GET /api/sessions with query params agent, profile, q, limit.
func (s *Server) handleGetSessions(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		writeJSONError(w, http.StatusInternalServerError, "session service unavailable")
		return
	}

	q := r.URL.Query()
	limit := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	filter := service.SessionFilter{
		Agent:   q.Get("agent"),
		Profile: q.Get("profile"),
		Query:   q.Get("q"),
		Limit:   limit,
	}

	sessions, err := s.sessions.ListSessions(r.Context(), filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sessions == nil {
		sessions = []service.SessionDTO{}
	}

	writeJSON(w, http.StatusOK, sessions)
}

// handleResumeSession handles POST /api/sessions/resume to resume a session in a terminal.
func (s *Server) handleResumeSession(w http.ResponseWriter, r *http.Request) {
	if s.sessions == nil {
		writeJSONError(w, http.StatusInternalServerError, "session service unavailable")
		return
	}

	var req service.ResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Agent) == "" || strings.TrimSpace(req.SessionID) == "" {
		writeJSONError(w, http.StatusBadRequest, "agent and session_id are required")
		return
	}

	if err := s.sessions.ResumeSessionInTerminal(r.Context(), req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "illegal") {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	terminalName := "terminal"
	if s.launcher != nil {
		terms := s.launcher.AvailableTerminals()
		if len(terms) > 0 {
			terminalName = terms[0]
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "launched",
		"terminal": terminalName,
	})
}
