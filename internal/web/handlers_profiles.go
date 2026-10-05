package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetProfiles handles GET /api/profiles (supports optional ?agent= query filter).
func (s *Server) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	agent := r.URL.Query().Get("agent")
	profiles, err := s.profiles.ListProfiles(r.Context(), agent)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profiles == nil {
		profiles = []service.ProfileDTO{}
	}
	writeJSON(w, http.StatusOK, profiles)
}

// handleCreateProfile handles POST /api/profiles to create or scaffold a new profile.
func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	var req service.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "profile name is required")
		return
	}

	created, err := s.profiles.CreateProfile(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// handleDeleteProfile handles DELETE /api/profiles/{agent}/{name} to remove an agent profile.
func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	agent := r.PathValue("agent")
	name := r.PathValue("name")
	if agent == "" || name == "" {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/profiles/"), "/")
		if len(parts) >= 2 {
			agent = parts[0]
			name = parts[1]
		}
	}

	if strings.TrimSpace(agent) == "" || strings.TrimSpace(name) == "" {
		writeJSONError(w, http.StatusBadRequest, "agent and profile name are required")
		return
	}

	if err := s.profiles.RemoveProfile(r.Context(), agent, name); err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "not allowed") {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type UpdateProfileConfigRequest struct {
	MCPGlobal     *bool `json:"mcp_global,omitempty"`
	PluginsGlobal *bool `json:"plugins_global,omitempty"`
}

// handleUpdateProfileConfig handles PATCH /api/profiles/{name}/config
func (s *Server) handleUpdateProfileConfig(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/profiles/"), "/")
		if len(parts) >= 1 {
			name = parts[0]
		}
	}

	if strings.TrimSpace(name) == "" {
		writeJSONError(w, http.StatusBadRequest, "profile name is required")
		return
	}

	var req UpdateProfileConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := s.profiles.UpdateProfileConfig(r.Context(), name, req.MCPGlobal, req.PluginsGlobal); err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "not allowed") {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
