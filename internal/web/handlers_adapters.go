package web

import (
	"net/http"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetAdapters handles GET /api/adapters returning all registered agent adapters and their runnable flags.
func (s *Server) handleGetAdapters(w http.ResponseWriter, r *http.Request) {
	if s.profiles == nil {
		writeJSONError(w, http.StatusInternalServerError, "profile service unavailable")
		return
	}

	adapters, err := s.profiles.ListAdapters(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if adapters == nil {
		adapters = []service.AdapterDTO{}
	}

	writeJSON(w, http.StatusOK, adapters)
}
