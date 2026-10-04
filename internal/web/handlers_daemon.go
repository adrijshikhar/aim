package web

import (
	"encoding/json"
	"net/http"

	"github.com/aim-cli/aim/internal/service"
)

// handleGetDaemon handles GET /api/daemon to check operational state of the OS background daemon.
func (s *Server) handleGetDaemon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	daemonSvc := s.DaemonService()
	if daemonSvc == nil {
		writeJSON(w, http.StatusOK, &service.DaemonDTO{})
		return
	}

	info, err := daemonSvc.GetStatus(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// handleDaemonInstall handles POST /api/daemon/install to install and start the background service.
func (s *Server) handleDaemonInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	daemonSvc := s.DaemonService()
	if daemonSvc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "daemon service unavailable")
		return
	}

	var req struct {
		BinaryPath string `json:"binary_path,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	info, err := daemonSvc.Install(r.Context(), req.BinaryPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// handleDaemonUninstall handles POST /api/daemon/uninstall to stop and remove the background service.
func (s *Server) handleDaemonUninstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	daemonSvc := s.DaemonService()
	if daemonSvc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "daemon service unavailable")
		return
	}

	if err := daemonSvc.Uninstall(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "uninstalled"})
}

// handleDaemonRun handles POST /api/daemon/run to manually trigger quota cache pre-warming.
func (s *Server) handleDaemonRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	daemonSvc := s.DaemonService()
	if daemonSvc == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "daemon service unavailable")
		return
	}

	if err := daemonSvc.RunOnce(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch updated status after execution
	info, _ := daemonSvc.GetStatus(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "completed",
		"message": "Quota caches refreshed successfully across all profiles",
		"info":    info,
	})
}
