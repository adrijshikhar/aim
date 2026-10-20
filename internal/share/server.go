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

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
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
		NodeID:         s.nodeID,
		HostName:       host,
		VPNIP:          r.Host,
		Profiles:       names,
		SharedProfiles: profiles,
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
		expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
		resp := LeaseResponse{
			Status:      "approved",
			LeaseID:     leaseID,
			ProfileName: matched.Name,
			Agent:       matched.Agent,
			ExpiresAt:   expiresAt,
			Payload:     payload,
		}

		borrowerID := req.RequesterID
		if borrowerID == "" {
			borrowerID = r.RemoteAddr
		}
		_ = s.store.SaveGrantedLease(GrantedLease{
			LeaseID:      leaseID,
			ProfileName:  matched.Name,
			Agent:        matched.Agent,
			BorrowerID:   borrowerID,
			BorrowerIP:   r.RemoteAddr,
			GrantedAt:    time.Now(),
			ExpiresAt:    expiresAt,
			Revoked:      false,
			AutoApproved: true,
		})

		_ = s.store.AppendAudit(AuditEntry{
			Action:  "LEASE_GRANTED",
			Actor:   borrowerID,
			Profile: matched.Name,
			Agent:   matched.Agent,
			Details: fmt.Sprintf("Auto-approved lease %s (%s) for %dh to %s", leaseID, matched.Name, hours, borrowerID),
		})

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

	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_REQUESTED",
		Actor:   req.RequesterID,
		Profile: matched.Name,
		Agent:   matched.Agent,
		Details: fmt.Sprintf("Lease request %s created by %s for %dh", reqID, req.RequesterID, req.RequestedHours),
	})

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
	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
	leaseID := generateID("ls")
	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     leaseID,
		ProfileName: pr.matched.Name,
		Agent:       pr.matched.Agent,
		ExpiresAt:   expiresAt,
		Payload:     payload,
	}

	_ = s.store.SaveGrantedLease(GrantedLease{
		LeaseID:      leaseID,
		ProfileName:  pr.matched.Name,
		Agent:        pr.matched.Agent,
		BorrowerID:   pr.info.Request.RequesterID,
		BorrowerIP:   pr.info.Request.RequesterIP,
		GrantedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		Revoked:      false,
		AutoApproved: false,
	})

	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_GRANTED",
		Actor:   pr.info.Request.RequesterID,
		Profile: pr.matched.Name,
		Agent:   pr.matched.Agent,
		Details: fmt.Sprintf("Manually approved lease %s (%s) for %dh", leaseID, pr.matched.Name, hours),
	})

	return resp, nil
}

func (s *Server) DenyRequest(reqID string) error {
	s.pendingMu.Lock()
	pr, exists := s.pending[reqID]
	if exists {
		delete(s.pending, reqID)
	}
	s.pendingMu.Unlock()
	if !exists {
		return fmt.Errorf("request %s not found or expired", reqID)
	}
	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_DENIED",
		Actor:   "local",
		Profile: pr.matched.Name,
		Agent:   pr.matched.Agent,
		Details: fmt.Sprintf("Host denied request %s from %s", reqID, pr.info.Request.RequesterID),
	})
	return nil
}

func (s *Server) packageCredentials(profileName, agent string) (*SharePayload, error) {
	profDir := filepath.Join(s.profilesDir, profileName)
	files := make(map[string]string)

	switch agent {
	case "claude":
		_ = profile.HarvestKeychainTokenToProfile("claude", profDir)
		for _, rel := range []string{
			".claude/.credentials.json",
			".claude/auth.json",
			".claude.json",
			".claude/.claude.json",
			".claude/settings.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	case "codex":
		codexAuth := filepath.Join(profDir, ".codex", "auth.json")
		if _, err := os.Stat(codexAuth); os.IsNotExist(err) {
			realHome := config.RealHomeDir()
			hostAuth := filepath.Join(realHome, ".codex", "auth.json")
			if data, err := os.ReadFile(hostAuth); err == nil && len(data) > 0 {
				_ = os.MkdirAll(filepath.Dir(codexAuth), 0700)
				_ = os.WriteFile(codexAuth, data, 0600)
			}
		}
		for _, rel := range []string{
			".codex/auth.json",
			".codex/config.toml",
			".config/codex/auth.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	case "agy":
		agyAuth := filepath.Join(profDir, ".gemini", "antigravity-cli", "antigravity-oauth-token")
		if _, err := os.Stat(agyAuth); os.IsNotExist(err) {
			realHome := config.RealHomeDir()
			hostAgy := filepath.Join(realHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")
			if data, err := os.ReadFile(hostAgy); err == nil && len(data) > 0 {
				_ = os.MkdirAll(filepath.Dir(agyAuth), 0700)
				_ = os.WriteFile(agyAuth, data, 0600)
			}
		}
		for _, rel := range []string{
			".gemini/antigravity-cli/antigravity-oauth-token",
			".gemini/antigravity-cli/config.json",
			".config/gcloud/application_default_credentials.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	case "gemini":
		for _, rel := range []string{
			".gemini/gemini-oauth-token",
			".gemini/oauth.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	}

	return &SharePayload{Files: files}, nil
}

func generateID(prefix string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
