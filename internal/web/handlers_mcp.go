package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetMcpServers handles GET /api/mcp?profile=xxx
func (s *Server) handleGetMcpServers(w http.ResponseWriter, r *http.Request) {
	mcpSvc := s.MCPService()
	if mcpSvc == nil {
		writeJSON(w, http.StatusOK, []service.MCPServerDTO{})
		return
	}

	profileName := strings.TrimSpace(r.URL.Query().Get("profile"))
	servers, err := mcpSvc.ListServers(r.Context(), profileName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if servers == nil {
		servers = []service.MCPServerDTO{}
	}

	writeJSON(w, http.StatusOK, servers)
}

// handleAddMcpServer handles POST /api/mcp
func (s *Server) handleAddMcpServer(w http.ResponseWriter, r *http.Request) {
	mcpSvc := s.MCPService()
	if mcpSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "mcp service not configured")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req service.AddMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := mcpSvc.AddServer(r.Context(), req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// handleSetMcpServerState handles PATCH /api/mcp/{name}/state
func (s *Server) handleSetMcpServerState(w http.ResponseWriter, r *http.Request) {
	mcpSvc := s.MCPService()
	if mcpSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "mcp service not configured")
		return
	}

	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "server name is required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req service.SetMCPServerStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if err := mcpSvc.SetServerState(r.Context(), name, req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// handleDeleteMcpServer handles DELETE /api/mcp/{name}?profile=xxx&clients=xxx
func (s *Server) handleDeleteMcpServer(w http.ResponseWriter, r *http.Request) {
	mcpSvc := s.MCPService()
	if mcpSvc == nil {
		writeJSONError(w, http.StatusInternalServerError, "mcp service not configured")
		return
	}

	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "server name is required")
		return
	}

	profileName := strings.TrimSpace(r.URL.Query().Get("profile"))
	clientsParam := strings.TrimSpace(r.URL.Query().Get("clients"))
	var clients []string
	if clientsParam != "" {
		clients = strings.Split(clientsParam, ",")
	}

	if err := mcpSvc.RemoveServer(r.Context(), name, profileName, clients); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
