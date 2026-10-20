package share

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PendingRequestInfo struct {
	RequestID string       `json:"request_id"`
	Request   LeaseRequest `json:"request"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type Server struct {
	store       *ShareStore
	profilesDir string
	nodeID      string
	port        int
	mux         *http.ServeMux
	httpServer  *http.Server
	listener    net.Listener
	pendingMu   sync.Mutex
	pending     map[string]*pendingRequest
}

type pendingRequest struct {
	info    PendingRequestInfo
	matched *SharedProfileConfig
	respCh  chan *LeaseResponse
}

func NewServer(store *ShareStore, profilesDir string, nodeID string, port int) *Server {
	s := &Server{
		store:       store,
		profilesDir: profilesDir,
		nodeID:      nodeID,
		port:        port,
		mux:         http.NewServeMux(),
		pending:     make(map[string]*pendingRequest),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = ln
	s.httpServer = &http.Server{
		Handler: s.mux,
	}
	return s.httpServer.Serve(ln)
}

func (s *Server) Close() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/share/ping", s.handlePing)
	s.mux.HandleFunc("/api/share/profiles", s.handleListProfiles)
	s.mux.HandleFunc("/api/share/request", s.handleRequest)
	s.mux.HandleFunc("/api/share/pending", s.handlePending)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profiles := s.store.GetSharedProfiles()
	var names []string
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	host, _ := os.Hostname()
	resp := PingResponse{
		NodeID:   s.nodeID,
		HostName: host,
		VPNIP:    r.Host,
		Profiles: names,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	profiles := s.store.GetSharedProfiles()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(profiles)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	profiles := s.store.GetSharedProfiles()
	var matched *SharedProfileConfig
	for _, p := range profiles {
		if p.Name == req.ProfileName {
			matched = &p
			break
		}
	}
	if matched == nil {
		http.Error(w, "profile not shared or found", http.StatusNotFound)
		return
	}

	if matched.AutoApprove {
		payload, err := s.packageCredentials(matched.Name, matched.Agent)
		if err != nil {
			http.Error(w, "failed to read credentials: "+err.Error(), http.StatusInternalServerError)
			return
		}
		leaseID := generateID("ls")
		hours := req.RequestedHours
		if hours <= 0 || time.Duration(hours)*time.Hour > matched.MaxLeaseTTL {
			hours = int(matched.MaxLeaseTTL.Hours())
		}
		resp := LeaseResponse{
			Status:      "approved",
			LeaseID:     leaseID,
			ProfileName: matched.Name,
			Agent:       matched.Agent,
			ExpiresAt:   time.Now().Add(time.Duration(hours) * time.Hour),
			Payload:     payload,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Manual approval pending
	reqID := generateID("req")
	s.pendingMu.Lock()
	pr := &pendingRequest{
		info: PendingRequestInfo{
			RequestID: reqID,
			Request:   req,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(5 * time.Minute),
		},
		matched: matched,
		respCh:  make(chan *LeaseResponse, 1),
	}
	s.pending[reqID] = pr
	s.pendingMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(LeaseResponse{
		Status:    "pending",
		RequestID: reqID,
		Message:   "approval pending on host",
	})
}

func (s *Server) handlePending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	pending := s.GetPendingRequests()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pending)
}

func (s *Server) GetPendingRequests() []PendingRequestInfo {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	var result []PendingRequestInfo
	now := time.Now()
	for id, pr := range s.pending {
		if now.After(pr.info.ExpiresAt) {
			delete(s.pending, id)
			continue
		}
		result = append(result, pr.info)
	}
	return result
}

func (s *Server) ApproveRequest(reqID string) (*LeaseResponse, error) {
	s.pendingMu.Lock()
	pr, exists := s.pending[reqID]
	if exists {
		delete(s.pending, reqID)
	}
	s.pendingMu.Unlock()

	if !exists {
		return nil, fmt.Errorf("request %s not found or expired", reqID)
	}

	payload, err := s.packageCredentials(pr.matched.Name, pr.matched.Agent)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}
	hours := pr.info.Request.RequestedHours
	if hours <= 0 || time.Duration(hours)*time.Hour > pr.matched.MaxLeaseTTL {
		hours = int(pr.matched.MaxLeaseTTL.Hours())
	}
	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     generateID("ls"),
		ProfileName: pr.matched.Name,
		Agent:       pr.matched.Agent,
		ExpiresAt:   time.Now().Add(time.Duration(hours) * time.Hour),
		Payload:     payload,
	}
	return resp, nil
}

func (s *Server) DenyRequest(reqID string) error {
	s.pendingMu.Lock()
	_, exists := s.pending[reqID]
	if exists {
		delete(s.pending, reqID)
	}
	s.pendingMu.Unlock()
	if !exists {
		return fmt.Errorf("request %s not found or expired", reqID)
	}
	return nil
}

func (s *Server) packageCredentials(profileName, agent string) (*SharePayload, error) {
	profDir := filepath.Join(s.profilesDir, profileName)
	files := make(map[string]string)

	switch agent {
	case "claude":
		credsPath := filepath.Join(profDir, ".claude", ".credentials.json")
		if data, err := os.ReadFile(credsPath); err == nil {
			files[".claude/.credentials.json"] = base64.StdEncoding.EncodeToString(data)
		}
		authPath := filepath.Join(profDir, ".claude", "auth.json")
		if data, err := os.ReadFile(authPath); err == nil {
			files[".claude/auth.json"] = base64.StdEncoding.EncodeToString(data)
		}
	case "codex":
		authPath := filepath.Join(profDir, ".config", "codex", "auth.json")
		if data, err := os.ReadFile(authPath); err == nil {
			files[".config/codex/auth.json"] = base64.StdEncoding.EncodeToString(data)
		}
	case "gemini":
		oauthPath := filepath.Join(profDir, ".gemini", "oauth.json")
		if data, err := os.ReadFile(oauthPath); err == nil {
			files[".gemini/oauth.json"] = base64.StdEncoding.EncodeToString(data)
		}
	}

	return &SharePayload{Files: files}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
