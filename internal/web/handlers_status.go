package web

import (
	"net/http"
	"runtime"
	"time"
)

// StatusDTO represents system and build information returned by /api/status.
type StatusDTO struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Uptime  string `json:"uptime"`
}

// handleGetStatus handles GET /api/status returning AIM version, OS, arch, and server uptime.
func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	version := s.version
	if version == "" {
		version = "dev"
	}

	uptime := time.Since(s.startTime).Truncate(time.Second).String()

	status := StatusDTO{
		Version: version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Uptime:  uptime,
	}

	writeJSON(w, http.StatusOK, status)
}
