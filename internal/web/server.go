package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/service"
)

//go:embed dist/*
var distFS embed.FS

// Server hosts the AIM web dashboard REST API and serves the embedded SPA.
type Server struct {
	profiles  service.ProfileService
	sessions  service.SessionService
	launcher  service.LauncherService
	port      int
	devMode   bool
	version   string
	startTime time.Time

	mux        *http.ServeMux
	handler    http.Handler
	httpServer *http.Server
	listener   net.Listener
	devProxy   *httputil.ReverseProxy
	mu         sync.Mutex
}

// NewServer creates a new Server configured with domain services, port, and devMode.
func NewServer(
	profiles service.ProfileService,
	sessions service.SessionService,
	launcher service.LauncherService,
	port int,
	devMode bool,
) *Server {
	s := &Server{
		profiles:  profiles,
		sessions:  sessions,
		launcher:  launcher,
		port:      port,
		devMode:   devMode,
		version:   "dev",
		startTime: time.Now(),
	}

	if devMode {
		devTarget := os.Getenv("AIM_DEV_PROXY_URL")
		if devTarget == "" {
			devTarget = "http://127.0.0.1:5173"
		}
		if targetURL, err := url.Parse(devTarget); err == nil {
			proxy := httputil.NewSingleHostReverseProxy(targetURL)
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				// If Vite dev server is unreachable, smoothly fallback to embedded SPA.
				s.serveEmbeddedSPA(w, r)
			}
			s.devProxy = proxy
		}
	}

	s.setupRoutes()
	return s
}

// SetVersion sets the AIM version string reported by /api/status.
func (s *Server) SetVersion(v string) {
	if v != "" {
		s.version = v
	}
}

// setupRoutes registers all REST API endpoints and static SPA fallback routing.
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	// REST API Handlers
	mux.HandleFunc("GET /api/profiles", s.handleGetProfiles)
	mux.HandleFunc("POST /api/profiles", s.handleCreateProfile)
	mux.HandleFunc("PATCH /api/profiles/{name}/config", s.handleUpdateProfileConfig)
	mux.HandleFunc("DELETE /api/profiles/{agent}/{name}", s.handleDeleteProfile)
	mux.HandleFunc("GET /api/sessions", s.handleGetSessions)
	mux.HandleFunc("POST /api/sessions/resume", s.handleResumeSession)
	mux.HandleFunc("GET /api/status", s.handleGetStatus)

	// Catch-all for undefined /api/ routes -> JSON 404
	mux.HandleFunc("/api/", s.handleAPINotFound)

	// SPA & static asset fallback for all other routes
	mux.HandleFunc("/", s.handleSPA)

	s.mux = mux
	s.handler = s.corsMiddleware(mux)
}

// corsMiddleware injects CORS headers and handles OPTIONS preflight when devMode is true.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.devMode {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// handleSPA serves either proxied Vite dev server responses or embedded static assets with index.html fallback.
func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.devMode && s.devProxy != nil {
		s.devProxy.ServeHTTP(w, r)
		return
	}

	s.serveEmbeddedSPA(w, r)
}

// serveEmbeddedSPA serves embedded files from dist/ or falls back to index.html for client-side routing.
func (s *Server) serveEmbeddedSPA(w http.ResponseWriter, r *http.Request) {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		http.Error(w, "Static asset filesystem error", http.StatusInternalServerError)
		return
	}

	cleanPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	if cleanPath == "" || cleanPath == "." {
		cleanPath = "index.html"
	}

	// Check if exact file exists in embedded filesystem
	if f, err := subFS.Open(cleanPath); err == nil {
		defer f.Close()
		stat, err := f.Stat()
		if err == nil && !stat.IsDir() {
			http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
			return
		}
	}

	// SPA fallback: any non-API GET request serves index.html
	indexBytes, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(indexBytes)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(indexBytes)
	}
}

// handleAPINotFound returns a JSON 404 for unmatched /api/ endpoints.
func (s *Server) handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotFound, "endpoint not found")
}

// Handler returns the underlying http.Handler with all routing and middleware configured.
func (s *Server) Handler() http.Handler {
	return s.handler
}

// Port returns the listening TCP port.
func (s *Server) Port() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		if addr, ok := s.listener.Addr().(*net.TCPAddr); ok {
			return addr.Port
		}
	}
	return s.port
}

// URL returns the loopback URL for the running server.
func (s *Server) URL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.Port())
}

// Listen binds the TCP listener on 127.0.0.1:<port>.
func (s *Server) Listen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return nil
	}
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = ln
	s.httpServer = &http.Server{
		Handler: s.handler,
	}
	return nil
}

// Serve accepts incoming connections on the listener and blocks until Shutdown or error.
func (s *Server) Serve() error {
	s.mu.Lock()
	ln := s.listener
	srv := s.httpServer
	s.mu.Unlock()

	if ln == nil {
		if err := s.Listen(); err != nil {
			return err
		}
		s.mu.Lock()
		ln = s.listener
		srv = s.httpServer
		s.mu.Unlock()
	}

	err := srv.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Start begins serving HTTP requests.
func (s *Server) Start() error {
	return s.Serve()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Close immediately closes the server.
func (s *Server) Close() error {
	return s.Shutdown(context.Background())
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
