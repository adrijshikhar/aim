package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetProfiles handles GET /api/profiles (supports optional ?agent=, ?profile=, and ?refresh=true).
func (s *Server) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	agent := r.URL.Query().Get("agent")
	profile := r.URL.Query().Get("profile")
	refresh := r.URL.Query().Get("refresh") == "true" || r.URL.Query().Get("refresh") == "1"

	var profiles []service.ProfileDTO
	var err error
	if refresh {
		profiles, err = s.profiles.RefreshUsage(r.Context(), agent, profile)
	} else {
		profiles, err = s.profiles.ListProfiles(r.Context(), agent)
	}
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

// handleDeleteProfile handles DELETE /api/profiles/{agent}/{name}, DELETE /api/profiles/{name},
// or DELETE /api/profiles?name=... to remove an agent profile or entire profile.
func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	agent := r.PathValue("agent")
	name := r.PathValue("name")
	if name == "" {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/profiles/"), "/")
		if len(parts) >= 2 && parts[0] != "" {
			agent = parts[0]
			name = parts[1]
		} else if len(parts) == 1 && parts[0] != "" {
			name = parts[0]
		}
	}

	if name == "" {
		name = r.URL.Query().Get("name")
	}
	if agent == "" {
		agent = r.URL.Query().Get("agent")
	}

	if strings.TrimSpace(name) == "" {
		writeJSONError(w, http.StatusBadRequest, "profile name is required")
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

// handleRefreshProfiles handles POST /api/profiles/refresh to force an active quota telemetry query.
func (s *Server) handleRefreshProfiles(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	var req struct {
		Agent   string `json:"agent,omitempty"`
		Profile string `json:"profile,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Agent == "" {
		req.Agent = r.URL.Query().Get("agent")
	}
	if req.Profile == "" {
		req.Profile = r.URL.Query().Get("profile")
	}

	profiles, err := s.profiles.RefreshUsage(r.Context(), req.Agent, req.Profile)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profiles == nil {
		profiles = []service.ProfileDTO{}
	}
	writeJSON(w, http.StatusOK, profiles)
}

// handleRefreshSingleProfile handles POST /api/profiles/{name}/refresh for refreshing a specific profile.
func (s *Server) handleRefreshSingleProfile(w http.ResponseWriter, r *http.Request) {
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
	agent := r.URL.Query().Get("agent")

	profiles, err := s.profiles.RefreshUsage(r.Context(), agent, name)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profiles == nil {
		profiles = []service.ProfileDTO{}
	}
	writeJSON(w, http.StatusOK, profiles)
}
