package share

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/profile"
)

type Server struct {
	store           *ShareStore
	profilesDir     string
	nodeID          string
	port            int
	mux             *http.ServeMux
	httpServer      *http.Server
	listener        net.Listener
	pendingMu       sync.Mutex
	pending         map[string]*pendingRequest
	approvedResults map[string]*LeaseResponse
	deniedResults   map[string]bool
}

type pendingRequest struct {
	info    PendingRequestInfo
	matched *SharedProfileConfig
	respCh  chan *LeaseResponse
}

func NewServer(store *ShareStore, profilesDir string, nodeID string, port int) *Server {
	s := &Server{
		store:           store,
		profilesDir:     profilesDir,
		nodeID:          nodeID,
		port:            port,
		mux:             http.NewServeMux(),
		pending:         make(map[string]*pendingRequest),
		approvedResults: make(map[string]*LeaseResponse),
		deniedResults:   make(map[string]bool),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) SetListener(ln net.Listener) {
	s.listener = ln
}

func (s *Server) Serve(ln net.Listener) error {
	s.listener = ln
	s.httpServer = &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s.httpServer.Serve(ln)
}

func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = ln
	s.httpServer = &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s.httpServer.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
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

func isLocalhost(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return host == "localhost"
	}
	return ip.IsLoopback()
}

func isPrivateOrVPNIP(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return host == "localhost"
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}
	_, cgnat, _ := net.ParseCIDR("100.64.0.0/10")
	if cgnat != nil && cgnat.Contains(ip) {
		return true
	}
	return false
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/share/ping", s.handlePing)
	s.mux.HandleFunc("/api/share/profiles", s.handleListProfiles)
	s.mux.HandleFunc("/api/share/request", s.handleRequest)
	s.mux.HandleFunc("/api/share/poll", s.handlePoll)
	s.mux.HandleFunc("/api/share/pending", s.handlePending)
	s.mux.HandleFunc("/api/share/approve", s.handleApprove)
	s.mux.HandleFunc("/api/share/deny", s.handleDeny)
	s.mux.HandleFunc("/api/share/cancel", s.handleCancel)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isPrivateOrVPNIP(r.RemoteAddr) {
		http.Error(w, "Forbidden: ping only accepted from private VPN / LAN subnets", http.StatusForbidden)
		return
	}
	profiles := make([]SharedProfileConfig, 0)
	names := make([]string, 0)
	for _, p := range s.store.GetAllSharedProfiles() {
		enabled, autoApprove := s.store.GetEffectiveProfileMode(p.Name)
		if enabled {
			p.AutoApprove = autoApprove
			p.Enabled = true
			profiles = append(profiles, p)
			names = append(names, p.Name)
		}
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
	if !isPrivateOrVPNIP(r.RemoteAddr) {
		http.Error(w, "Forbidden: profile listing only accepted from private VPN / LAN subnets", http.StatusForbidden)
		return
	}
	allProfiles := s.store.GetAllSharedProfiles()
	exposed := make([]SharedProfileConfig, 0)
	for _, p := range allProfiles {
		enabled, autoApprove := s.store.GetEffectiveProfileMode(p.Name)
		if enabled {
			p.AutoApprove = autoApprove
			p.Enabled = true
			exposed = append(exposed, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(exposed)
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isPrivateOrVPNIP(r.RemoteAddr) {
		http.Error(w, "Forbidden: lease requests only accepted from private VPN / LAN subnets", http.StatusForbidden)
		return
	}
	var req LeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	enabled, autoApprove := s.store.GetEffectiveProfileMode(req.ProfileName)
	if !enabled {
		http.Error(w, fmt.Sprintf("profile %q is not currently available for sharing (host policy is off or restricted)", req.ProfileName), http.StatusForbidden)
		return
	}

	profiles := s.store.GetAllSharedProfiles()
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
	matched.AutoApprove = autoApprove

	if matched.AutoApprove {
		payload, err := s.packageCredentials(matched.Name, matched.Agent)
		if err != nil {
			http.Error(w, "failed to read credentials: "+err.Error(), http.StatusInternalServerError)
			return
		}
		leaseID := generateID("ls")
		duration := time.Duration(req.RequestedHours) * time.Hour
		if duration <= 0 || (matched.MaxLeaseTTL > 0 && duration > matched.MaxLeaseTTL) {
			if matched.MaxLeaseTTL > 0 {
				duration = matched.MaxLeaseTTL
			} else {
				duration = 2 * time.Hour
			}
		}
		expiresAt := time.Now().Add(duration)
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
		if err := s.store.SaveGrantedLease(GrantedLease{
			LeaseID:      leaseID,
			ProfileName:  matched.Name,
			Agent:        matched.Agent,
			BorrowerID:   borrowerID,
			BorrowerIP:   r.RemoteAddr,
			GrantedAt:    time.Now(),
			ExpiresAt:    expiresAt,
			Revoked:      false,
			AutoApproved: true,
		}); err != nil {
			http.Error(w, "failed to record granted lease: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_ = s.store.AppendAudit(AuditEntry{
			Action:  "LEASE_GRANTED",
			Actor:   borrowerID,
			Profile: matched.Name,
			Agent:   matched.Agent,
			Details: fmt.Sprintf("Auto-approved lease %s (%s) for %v to %s", leaseID, matched.Name, duration, borrowerID),
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Manual approval pending
	reqID := generateID("req")
	expiresAt := time.Now().Add(30 * time.Minute)
	s.pendingMu.Lock()
	pr := &pendingRequest{
		info: PendingRequestInfo{
			RequestID: reqID,
			Request:   req,
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		},
		matched: matched,
		respCh:  make(chan *LeaseResponse, 1),
	}
	s.pending[reqID] = pr
	s.pendingMu.Unlock()

	_ = s.store.SaveShareRequest(ShareRequestRecord{
		ID:             reqID,
		Direction:      "incoming",
		PeerID:         req.RequesterID,
		PeerIP:         req.RequesterIP,
		ProfileName:    matched.Name,
		Agent:          matched.Agent,
		RequestedHours: req.RequestedHours,
		Status:         "pending",
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
	})

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
	if !isLocalhost(r) {
		http.Error(w, "Forbidden: pending list requires local host authorization", http.StatusForbidden)
		return
	}
	pending := s.GetPendingRequests()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pending)
}

func (s *Server) handlePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isPrivateOrVPNIP(r.RemoteAddr) {
		http.Error(w, "Forbidden: poll requests only accepted from private VPN / LAN subnets", http.StatusForbidden)
		return
	}
	reqID := r.URL.Query().Get("request_id")
	if reqID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	s.pendingMu.Lock()
	resp, approved := s.approvedResults[reqID]
	denied := s.deniedResults[reqID]
	pr, pending := s.pending[reqID]
	s.pendingMu.Unlock()

	if approved {
		if time.Now().After(resp.ExpiresAt) {
			s.pendingMu.Lock()
			delete(s.approvedResults, reqID)
			s.pendingMu.Unlock()
			http.Error(w, "lease has expired", http.StatusGone)
			return
		}
		// Verify lease is not revoked
		leases := s.store.GetGrantedLeases()
		for _, gl := range leases {
			if gl.LeaseID == resp.LeaseID && gl.Revoked {
				http.Error(w, "lease was revoked", http.StatusGone)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if denied {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "denied",
			"request_id": reqID,
			"message":    "Lease request was denied by host",
		})
		return
	}

	if pending {
		if time.Now().After(pr.info.ExpiresAt) {
			s.pendingMu.Lock()
			delete(s.pending, reqID)
			s.pendingMu.Unlock()
			http.Error(w, "request expired", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "pending",
			"request_id": reqID,
			"message":    "approval pending on host",
		})
		return
	}

	// Check SQLite for durable state
	rec, err := s.store.GetShareRequest(reqID)
	if err == nil && rec != nil {
		if rec.Status == "approved" && rec.LeaseID != "" {
			leases := s.store.GetGrantedLeases()
			for _, gl := range leases {
				if gl.LeaseID == rec.LeaseID {
					if gl.Revoked || time.Now().After(gl.ExpiresAt) {
						http.Error(w, "lease has expired or was revoked", http.StatusGone)
						return
					}
					payload, _ := s.packageCredentials(gl.ProfileName, gl.Agent)
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(LeaseResponse{
						Status:      "approved",
						LeaseID:     gl.LeaseID,
						ProfileName: gl.ProfileName,
						Agent:       gl.Agent,
						ExpiresAt:   gl.ExpiresAt,
						Payload:     payload,
					})
					return
				}
			}
		}
		if rec.Status == "denied" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":     "denied",
				"request_id": reqID,
				"message":    "Lease request was denied by host",
			})
			return
		}
		if rec.Status == "cancelled" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":     "cancelled",
				"request_id": reqID,
				"message":    "Lease request was cancelled",
			})
			return
		}
		if rec.Status == "pending" {
			if time.Now().After(rec.ExpiresAt) {
				_ = s.store.UpdateRequestStatus(reqID, "expired", "", "Request expired")
				http.Error(w, "request expired", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":     "pending",
				"request_id": reqID,
				"message":    "approval pending on host",
			})
			return
		}
	}

	http.Error(w, "request not found or expired", http.StatusNotFound)
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isPrivateOrVPNIP(r.RemoteAddr) {
		http.Error(w, "Forbidden: cancellation requests only accepted from private VPN / LAN subnets", http.StatusForbidden)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	s.pendingMu.Lock()
	pr, exists := s.pending[req.RequestID]
	if exists {
		delete(s.pending, req.RequestID)
		if pr.respCh != nil {
			close(pr.respCh)
		}
	}
	s.pendingMu.Unlock()

	actor := "remote"
	profile := ""
	agent := ""
	if exists {
		actor = pr.info.Request.RequesterID
		profile = pr.matched.Name
		agent = pr.matched.Agent
	} else if rec, _ := s.store.GetShareRequest(req.RequestID); rec != nil {
		actor = rec.PeerID
		profile = rec.ProfileName
		agent = rec.Agent
	}

	cancelled, _ := s.store.CancelShareRequest(req.RequestID)
	if !cancelled && !exists {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "request was already decided or cancelled"})
		return
	}

	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_CANCELLED",
		Actor:   actor,
		Profile: profile,
		Agent:   agent,
		Details: fmt.Sprintf("Lease request %s was cancelled by %s", req.RequestID, actor),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLocalhost(r) {
		http.Error(w, "Forbidden: approve requires local host authorization", http.StatusForbidden)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}
	resp, err := s.ApproveRequest(req.RequestID)
	if err != nil {
		if strings.Contains(err.Error(), "has status") || strings.Contains(err.Error(), "has expired") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDeny(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLocalhost(r) {
		http.Error(w, "Forbidden: deny requires local host authorization", http.StatusForbidden)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}
	if err := s.DenyRequest(req.RequestID); err != nil {
		if strings.Contains(err.Error(), "has status") || strings.Contains(err.Error(), "has expired") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) GetPendingRequests() []PendingRequestInfo {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()

	seen := make(map[string]bool)
	var result []PendingRequestInfo
	now := time.Now()

	for id, pr := range s.pending {
		if now.After(pr.info.ExpiresAt) {
			delete(s.pending, id)
			continue
		}
		seen[id] = true
		result = append(result, pr.info)
	}

	// Also retrieve unexpired pending requests persisted in SQLite
	dbReqs := s.store.GetPendingIncomingRequests()
	for _, r := range dbReqs {
		if !seen[r.ID] && now.Before(r.ExpiresAt) {
			seen[r.ID] = true
			result = append(result, r.ToPendingInfo())
		}
	}

	return result
}

func (s *Server) ApproveRequest(reqID string) (*LeaseResponse, error) {
	s.pendingMu.Lock()
	if resp, ok := s.approvedResults[reqID]; ok {
		s.pendingMu.Unlock()
		return resp, nil
	}
	if s.deniedResults[reqID] {
		s.pendingMu.Unlock()
		return nil, fmt.Errorf("request %s was denied", reqID)
	}
	pr, exists := s.pending[reqID]
	if exists {
		delete(s.pending, reqID)
	}
	s.pendingMu.Unlock()

	var (
		profileName    string
		agent          string
		requesterID    string
		requesterIP    string
		requestedHours int
		respCh         chan *LeaseResponse
	)

	if exists {
		profileName = pr.matched.Name
		agent = pr.matched.Agent
		requesterID = pr.info.Request.RequesterID
		requesterIP = pr.info.Request.RequesterIP
		requestedHours = pr.info.Request.RequestedHours
		respCh = pr.respCh
	} else {
		// Look up in SQLite
		rec, err := s.store.GetShareRequest(reqID)
		if err != nil || rec == nil {
			return nil, fmt.Errorf("request %s not found", reqID)
		}
		if rec.Status == "approved" && rec.LeaseID != "" {
			leases := s.store.GetGrantedLeases()
			for _, gl := range leases {
				if gl.LeaseID == rec.LeaseID {
					payload, _ := s.packageCredentials(gl.ProfileName, gl.Agent)
					return &LeaseResponse{
						Status:      "approved",
						LeaseID:     gl.LeaseID,
						ProfileName: gl.ProfileName,
						Agent:       gl.Agent,
						ExpiresAt:   gl.ExpiresAt,
						Payload:     payload,
					}, nil
				}
			}
		}
		if rec.Status != "pending" {
			return nil, fmt.Errorf("request %s has status %s", reqID, rec.Status)
		}
		if time.Now().After(rec.ExpiresAt) {
			_ = s.store.UpdateRequestStatus(reqID, "expired", "", "Request expired")
			return nil, fmt.Errorf("request %s has expired", reqID)
		}

		profileName = rec.ProfileName
		agent = rec.Agent
		requesterID = rec.PeerID
		requesterIP = rec.PeerIP
		requestedHours = rec.RequestedHours
	}

	duration := time.Duration(requestedHours) * time.Hour
	if duration <= 0 {
		duration = 2 * time.Hour
	}
	for _, p := range s.store.GetAllSharedProfiles() {
		if p.Name == profileName && p.MaxLeaseTTL > 0 && duration > p.MaxLeaseTTL {
			duration = p.MaxLeaseTTL
			break
		}
	}
	expiresAt := time.Now().Add(duration)
	leaseID := generateID("ls")

	// Atomically transition status from pending to approved in SQLite
	// This prevents competing approvals/denials from creating multiple grants.
	transitioned, err := s.store.TransitionRequestStatus(reqID, "pending", "approved", leaseID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to transition request status: %w", err)
	}
	if !transitioned {
		// Check current status
		rec, _ := s.store.GetShareRequest(reqID)
		if rec != nil && rec.Status == "approved" && rec.LeaseID != "" {
			leases := s.store.GetGrantedLeases()
			for _, gl := range leases {
				if gl.LeaseID == rec.LeaseID {
					payload, _ := s.packageCredentials(gl.ProfileName, gl.Agent)
					return &LeaseResponse{
						Status:      "approved",
						LeaseID:     gl.LeaseID,
						ProfileName: gl.ProfileName,
						Agent:       gl.Agent,
						ExpiresAt:   gl.ExpiresAt,
						Payload:     payload,
					}, nil
				}
			}
		}
		if rec != nil {
			return nil, fmt.Errorf("cannot approve request %s: current status is %s", reqID, rec.Status)
		}
		return nil, fmt.Errorf("request %s was already decided or cancelled", reqID)
	}

	payload, err := s.packageCredentials(profileName, agent)
	if err != nil {
		_ = s.store.UpdateRequestStatus(reqID, "pending", "", "Failed to read credentials: "+err.Error())
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	resp := &LeaseResponse{
		Status:      "approved",
		LeaseID:     leaseID,
		ProfileName: profileName,
		Agent:       agent,
		ExpiresAt:   expiresAt,
		Payload:     payload,
	}

	if err := s.store.SaveGrantedLease(GrantedLease{
		LeaseID:      leaseID,
		ProfileName:  profileName,
		Agent:        agent,
		BorrowerID:   requesterID,
		BorrowerIP:   requesterIP,
		GrantedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		Revoked:      false,
		AutoApproved: false,
	}); err != nil {
		_ = s.store.UpdateRequestStatus(reqID, "pending", "", "Failed to record granted lease: "+err.Error())
		return nil, fmt.Errorf("failed to record granted lease: %w", err)
	}

	s.pendingMu.Lock()
	s.approvedResults[reqID] = resp
	s.pendingMu.Unlock()

	if respCh != nil {
		select {
		case respCh <- resp:
		default:
		}
	}

	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_GRANTED",
		Actor:   requesterID,
		Profile: profileName,
		Agent:   agent,
		Details: fmt.Sprintf("Manually approved lease %s (%s) for %v to %s", leaseID, profileName, duration, requesterID),
	})

	return resp, nil
}

func (s *Server) DenyRequest(reqID string) error {
	s.pendingMu.Lock()
	if s.deniedResults[reqID] {
		s.pendingMu.Unlock()
		return nil
	}
	if _, ok := s.approvedResults[reqID]; ok {
		s.pendingMu.Unlock()
		return fmt.Errorf("request %s was already approved", reqID)
	}
	pr, exists := s.pending[reqID]
	s.pendingMu.Unlock()

	var (
		actor   = "unknown"
		profile = ""
		agent   = ""
	)
	if exists {
		actor = pr.info.Request.RequesterID
		profile = pr.matched.Name
		agent = pr.matched.Agent
	} else {
		rec, err := s.store.GetShareRequest(reqID)
		if err != nil || rec == nil {
			return fmt.Errorf("request %s not found", reqID)
		}
		actor = rec.PeerID
		profile = rec.ProfileName
		agent = rec.Agent
	}

	transitioned, err := s.store.TransitionRequestStatus(reqID, "pending", "denied", "", "Host denied request")
	if err != nil {
		return fmt.Errorf("failed to transition request status: %w", err)
	}
	if !transitioned {
		rec, _ := s.store.GetShareRequest(reqID)
		if rec != nil {
			return fmt.Errorf("request %s cannot be denied: status is %s", reqID, rec.Status)
		}
		return fmt.Errorf("request %s was already decided", reqID)
	}

	// Update cached memory state ONLY after database transition successfully commits
	s.pendingMu.Lock()
	delete(s.pending, reqID)
	s.deniedResults[reqID] = true
	s.pendingMu.Unlock()

	if exists && pr.respCh != nil {
		close(pr.respCh)
	}

	_ = s.store.AppendAudit(AuditEntry{
		Action:  "LEASE_DENIED",
		Actor:   "local",
		Profile: profile,
		Agent:   agent,
		Details: fmt.Sprintf("Host denied request %s from %s", reqID, actor),
	})
	return nil
}

func (s *Server) packageCredentials(profileName, agent string) (*SharePayload, error) {
	return PackageProfileCredentials(s.profilesDir, profileName, agent)
}

// PackageProfileCredentials extracts relevant agent credentials for the specified profile.
func PackageProfileCredentials(profilesDir, profileName, agent string) (*SharePayload, error) {
	cleanProfilesRoot := filepath.Clean(profilesDir)
	cleanProfDir := filepath.Clean(filepath.Join(profilesDir, profileName))
	if !strings.HasPrefix(cleanProfDir, cleanProfilesRoot+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid profile name %q: path traversal", profileName)
	}
	profDir := cleanProfDir
	files := make(map[string]string)

	switch agent {
	case "claude":
		for _, rel := range []string{
			".claude/.credentials.json",
			".claude/auth.json",
			".claude.json",
			".claude/.claude.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
		// If credentials are not on disk, read exclusively from the selected profile's scoped Keychain service
		if _, hasCreds := files[".claude/.credentials.json"]; !hasCreds {
			claudeDir := filepath.Join(profDir, ".claude")
			if tokData := profile.ReadScopedClaudeKeychainCredentials(claudeDir); len(tokData) > 0 {
				files[".claude/.credentials.json"] = base64.StdEncoding.EncodeToString(tokData)
			}
		}
	case "codex":
		for _, rel := range []string{
			".codex/auth.json",
			".codex/config.toml",
			".config/codex/auth.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	case "agy":
		for _, rel := range []string{
			".gemini/antigravity-cli/antigravity-oauth-token",
			".gemini/antigravity-cli/config.json",
			".config/gcloud/application_default_credentials.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	case "gemini":
		for _, rel := range []string{
			".gemini/gemini-oauth-token",
			".gemini/oauth.json",
		} {
			p := filepath.Join(profDir, filepath.FromSlash(rel))
			if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
				files[rel] = base64.StdEncoding.EncodeToString(data)
			}
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no credentials found for profile %q (agent %s)", profileName, agent)
	}

	return &SharePayload{Files: files}, nil
}

// GenerateID produces a unique prefixed identifier with 128-bit cryptographic random bytes.
func GenerateID(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

func generateID(prefix string) string {
	return GenerateID(prefix)
}
