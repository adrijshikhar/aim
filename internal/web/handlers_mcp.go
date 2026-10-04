package web

import (
	"net/http"
	"strings"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetMcpServers handles GET /api/mcp?profile=xxx
func (s *Server) handleGetMcpServers(w http.ResponseWriter, r *http.Request) {
	if s.mcp == nil {
		writeJSON(w, http.StatusOK, []service.MCPServerDTO{})
		return
	}

	profileName := strings.TrimSpace(r.URL.Query().Get("profile"))
	servers, err := s.mcp.ListServers(r.Context(), profileName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if servers == nil {
		servers = []service.MCPServerDTO{}
	}

	writeJSON(w, http.StatusOK, servers)
}
