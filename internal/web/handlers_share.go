package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/share"
)

type ShareStatusResponse struct {
	VPN                 *share.VPNStatus            `json:"vpn"`
	GlobalMode          string                      `json:"global_mode"`
	EffectiveGlobalMode string                      `json:"effective_global_mode"`
	SharedProfiles      []share.SharedProfileConfig `json:"shared_profiles"`
	Peers               []share.PeerInfo            `json:"peers"`
	ActiveLeases        []ActiveLeaseInfo           `json:"active_leases"`
	GrantedLeases       []share.GrantedLease        `json:"granted_leases"`
	PendingRequests     []share.PendingRequestInfo  `json:"pending_requests,omitempty"`
	OutgoingRequests    []share.ShareRequestRecord  `json:"outgoing_requests,omitempty"`
	AuditLogs           []share.AuditEntry          `json:"audit_logs,omitempty"`
	Schedules           []share.SchedulePolicy      `json:"schedules,omitempty"`
}

type ActiveLeaseInfo struct {
	LocalProfile  string    `json:"local_profile"`
	SourcePeer    string    `json:"source_peer"`
	SourceProfile string    `json:"source_profile"`
	Agent         string    `json:"agent"`
	GrantedAt     time.Time `json:"granted_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsExpired     bool      `json:"is_expired"`
}

func (s *Server) handleGetShareStatus(w http.ResponseWriter, r *http.Request) {
	vpn, _ := share.DetectVPN()
	store := s.getShareStore()

	// Live refresh of known peers if requested or on poll
	if r.URL.Query().Get("refresh") == "true" {
		disc := share.NewDiscovery(store, 4319)
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		_, _ = disc.RefreshPeers(ctx)
	}

	shared := store.GetSharedProfiles()
	peers := store.GetPeers()
	granted := store.GetGrantedLeases()
	audit := store.GetAuditLogs(50)

	// Read active leased profiles on disk
	var activeLeases []ActiveLeaseInfo
	profRoot := filepath.Join(config.BaseDir(), "profiles")
	entries, err := os.ReadDir(profRoot)
	if err == nil {
		now := time.Now()
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			metaPath := filepath.Join(profRoot, e.Name(), ".aim-lease.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var meta share.LeaseMeta
			if err := json.Unmarshal(data, &meta); err != nil {
				continue
			}
			activeLeases = append(activeLeases, ActiveLeaseInfo{
				LocalProfile:  e.Name(),
				SourcePeer:    meta.SourcePeer,
				SourceProfile: meta.SourceProfile,
				Agent:         meta.Agent,
				GrantedAt:     meta.GrantedAt,
				ExpiresAt:     meta.ExpiresAt,
				IsExpired:     now.After(meta.ExpiresAt),
			})
		}
	}

	// Query active pending approval requests from P2P daemon on :4319
	var pendingRequests []share.PendingRequestInfo
	ctxPending, cancelPending := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancelPending()
	pReq, err := http.NewRequestWithContext(ctxPending, http.MethodGet, "http://127.0.0.1:4319/api/share/pending", nil)
	if err == nil {
		if pResp, err := http.DefaultClient.Do(pReq); err == nil {
			if pResp.StatusCode == http.StatusOK {
				_ = json.NewDecoder(pResp.Body).Decode(&pendingRequests)
			}
			pResp.Body.Close()
		}
	}

	// If background daemon did not return pending requests, fallback to SQLite store
	if len(pendingRequests) == 0 {
		dbIncoming := store.GetPendingIncomingRequests()
		for _, req := range dbIncoming {
			pendingRequests = append(pendingRequests, req.ToPendingInfo())
		}
	}

	outgoingRequests := store.GetPendingOutgoingRequests()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ShareStatusResponse{
		VPN:                 vpn,
		GlobalMode:          store.GetGlobalMode(),
		EffectiveGlobalMode: store.GetEffectiveGlobalMode(),
		SharedProfiles:      shared,
		Peers:               peers,
		ActiveLeases:        activeLeases,
		GrantedLeases:       granted,
		PendingRequests:     pendingRequests,
		OutgoingRequests:    outgoingRequests,
		AuditLogs:           audit,
		Schedules:           store.GetSchedules(),
	})
}

type SetGlobalModeRequest struct {
	Mode string `json:"mode"`
}

func (s *Server) handleSetGlobalMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SetGlobalModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Mode == "" {
		http.Error(w, "mode is required", http.StatusBadRequest)
		return
	}
	store := s.getShareStore()
	if err := store.SetGlobalMode(req.Mode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = store.AppendAudit(share.AuditEntry{
		Action:  "GLOBAL_MODE_CHANGED",
		Actor:   "local",
		Details: fmt.Sprintf("Host switched global sharing mode to %s", req.Mode),
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "mode": req.Mode})
}

type ToggleShareRequest struct {
	Name        string `json:"name"`
	Agent       string `json:"agent"`
	AutoApprove bool   `json:"auto_approve"`
	TTL         string `json:"ttl"`
	Enabled     bool   `json:"enabled"`
}

func (s *Server) handleToggleShare(w http.ResponseWriter, r *http.Request) {
	var req ToggleShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	if !req.Enabled {
		if err := store.DisableProfile(req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = store.AppendAudit(share.AuditEntry{
			Action:  "PROFILE_SHARING_OFF",
			Actor:   "local",
			Profile: req.Name,
			Details: fmt.Sprintf("Disabled sharing for profile %s", req.Name),
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "enabled": false})
		return
	}

	ttl, err := time.ParseDuration(req.TTL)
	if err != nil || ttl <= 0 {
		ttl = 4 * time.Hour
	}

	if err := store.EnableProfile(req.Name, req.Agent, req.AutoApprove, ttl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = store.AppendAudit(share.AuditEntry{
		Action:  "PROFILE_SHARING_ON",
		Actor:   "local",
		Profile: req.Name,
		Agent:   req.Agent,
		Details: fmt.Sprintf("Enabled sharing for profile %s (%s, auto_approve=%v, ttl=%s)", req.Name, req.Agent, req.AutoApprove, ttl),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "enabled": true})
}

func (s *Server) handleGetSharePeers(w http.ResponseWriter, r *http.Request) {
	store := s.getShareStore()
	peers := store.GetPeers()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peers)
}

type AddPeerRequest struct {
	Host string `json:"host"`
}

func (s *Server) handleAddSharePeer(w http.ResponseWriter, r *http.Request) {
	var req AddPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Host == "" {
		http.Error(w, "host is required", http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	disc := share.NewDiscovery(store, 4319)
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()

	peer, err := disc.PingPeer(ctx, req.Host)
	if err != nil && !strings.Contains(req.Host, ".") {
		peer, err = disc.PingPeer(ctx, req.Host+".local")
	}
	if err != nil {
		http.Error(w, "could not reach peer: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peer)
}

type RequestLeasePayload struct {
	Peer    string `json:"peer"`
	Profile string `json:"profile"`
	Agent   string `json:"agent"`
	Hours   int    `json:"hours"`
}

func (s *Server) handleRequestShareLease(w http.ResponseWriter, r *http.Request) {
	var req RequestLeasePayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	targetHost := req.Peer
	for _, p := range store.GetPeers() {
		shortHost := strings.TrimSuffix(strings.ToLower(p.HostName), ".local")
		if strings.EqualFold(p.HostName, req.Peer) || strings.EqualFold(shortHost, strings.ToLower(req.Peer)) || p.VPNIP == req.Peer {
			if p.VPNIP != "" {
				targetHost = p.VPNIP
			} else if p.HostName != "" {
				targetHost = p.HostName
			}
			break
		}
	}

	client := share.NewClient(4319)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.RequestLease(ctx, targetHost, req.Profile, req.Agent, req.Hours)
	if err != nil && !strings.Contains(targetHost, ".") {
		resp, err = client.RequestLease(ctx, targetHost+".local", req.Profile, req.Agent, req.Hours)
	}
	if err != nil {
		http.Error(w, "lease request failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Always record the lease request action in SQLite audit log
	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_REQUESTED",
		Actor:   "local",
		Profile: req.Profile,
		Agent:   req.Agent,
		Details: fmt.Sprintf("Requested %d-hour lease for profile %s from %s", req.Hours, req.Profile, req.Peer),
	})

	if resp.Status != "approved" {
		if resp.RequestID != "" {
			_ = store.SaveShareRequest(share.ShareRequestRecord{
				ID:             resp.RequestID,
				Direction:      "outgoing",
				PeerID:         req.Peer,
				PeerIP:         targetHost,
				ProfileName:    req.Profile,
				Agent:          req.Agent,
				RequestedHours: req.Hours,
				Status:         "pending",
				CreatedAt:      time.Now(),
				ExpiresAt:      time.Now().Add(30 * time.Minute),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	mgr := share.NewLeaseManager(filepath.Join(config.BaseDir(), "profiles"))
	localName, err := mgr.InstallLease(req.Peer, req.Profile, resp)
	if err != nil {
		http.Error(w, "failed to install lease: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.RequestID != "" {
		_ = store.UpdateRequestStatus(resp.RequestID, "approved", resp.LeaseID, "")
	}

	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_BORROWED",
		Actor:   req.Peer,
		Profile: req.Profile,
		Agent:   resp.Agent,
		Details: fmt.Sprintf("Lease %s granted by %s for profile %s (installed as %s, expires in %d hours)", resp.LeaseID, req.Peer, req.Profile, localName, req.Hours),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":            true,
		"local_profile": localName,
		"agent":         resp.Agent,
		"expires_at":    resp.ExpiresAt,
	})
}

type ReleaseLeasePayload struct {
	Profile string `json:"profile"`
}

func (s *Server) handleReleaseShareLease(w http.ResponseWriter, r *http.Request) {
	var req ReleaseLeasePayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Profile == "" {
		http.Error(w, "profile is required", http.StatusBadRequest)
		return
	}

	mgr := share.NewLeaseManager(filepath.Join(config.BaseDir(), "profiles"))
	if err := mgr.ReleaseLease(req.Profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleCleanExpiredShareLeases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mgr := share.NewLeaseManager(filepath.Join(config.BaseDir(), "profiles"))
	cleaned, err := mgr.CleanExpiredLeases()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "cleaned": cleaned})
}

type RevokeLeasePayload struct {
	LeaseID string `json:"lease_id"`
}

func (s *Server) handleRevokeShareLease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RevokeLeasePayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LeaseID == "" {
		http.Error(w, "lease_id is required", http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	revoked, err := store.RevokeGrantedLease(req.LeaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_REVOKED",
		Actor:   "local",
		Profile: revoked.ProfileName,
		Agent:   revoked.Agent,
		Details: fmt.Sprintf("Host revoked lease %s granted to %s", req.LeaseID, revoked.BorrowerID),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "revoked": revoked})
}

func (s *Server) handleApproveShareRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(req)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1. First attempt to call the running P2P share daemon on port 4319
	pReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:4319/api/share/approve", bytes.NewReader(body))
	if err == nil {
		pReq.Header.Set("Content-Type", "application/json")
		pResp, err := http.DefaultClient.Do(pReq)
		if err == nil {
			defer pResp.Body.Close()
			if pResp.StatusCode == http.StatusOK {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(pResp.StatusCode)
				_, _ = io.Copy(w, pResp.Body)
				return
			}
		}
	}

	// 2. Direct fallback via Store if daemon on :4319 is unreachable or returned non-200
	store := s.getShareStore()
	rec, err := store.GetShareRequest(req.RequestID)
	if err != nil || rec == nil {
		http.Error(w, "request not found or expired", http.StatusNotFound)
		return
	}

	if rec.Status == "approved" && rec.LeaseID != "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "lease_id": rec.LeaseID, "status": "approved"})
		return
	}

	duration := time.Duration(rec.RequestedHours) * time.Hour
	if duration <= 0 {
		duration = 2 * time.Hour
	}
	for _, p := range store.GetSharedProfiles() {
		if p.Name == rec.ProfileName && p.MaxLeaseTTL > 0 && duration > p.MaxLeaseTTL {
			duration = p.MaxLeaseTTL
			break
		}
	}
	expiresAt := time.Now().Add(duration)
	leaseID := share.GenerateID("ls")

	payload, err := share.PackageProfileCredentials(filepath.Join(config.BaseDir(), "profiles"), rec.ProfileName, rec.Agent)
	if err != nil {
		http.Error(w, "failed to read credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := store.SaveGrantedLease(share.GrantedLease{
		LeaseID:      leaseID,
		ProfileName:  rec.ProfileName,
		Agent:        rec.Agent,
		BorrowerID:   rec.PeerID,
		BorrowerIP:   rec.PeerIP,
		GrantedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		Revoked:      false,
		AutoApproved: false,
	}); err != nil {
		http.Error(w, "failed to record granted lease: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = store.UpdateRequestStatus(req.RequestID, "approved", leaseID, "")
	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_APPROVED",
		Actor:   "local",
		Profile: rec.ProfileName,
		Agent:   rec.Agent,
		Details: fmt.Sprintf("Host approved lease %s for request %s (peer %s, duration %v)", leaseID, req.RequestID, rec.PeerID, duration),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(share.LeaseResponse{
		Status:      "approved",
		LeaseID:     leaseID,
		ProfileName: rec.ProfileName,
		Agent:       rec.Agent,
		ExpiresAt:   expiresAt,
		Payload:     payload,
	})
}

func (s *Server) handleDenyShareRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}
	body, _ := json.Marshal(req)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1. First attempt to call the running P2P share daemon on port 4319
	pReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:4319/api/share/deny", bytes.NewReader(body))
	if err == nil {
		pReq.Header.Set("Content-Type", "application/json")
		pResp, err := http.DefaultClient.Do(pReq)
		if err == nil {
			defer pResp.Body.Close()
			if pResp.StatusCode == http.StatusOK {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(pResp.StatusCode)
				_, _ = io.Copy(w, pResp.Body)
				return
			}
		}
	}

	// 2. Direct fallback via Store if daemon on :4319 is unreachable or returned non-200
	store := s.getShareStore()
	rec, err := store.GetShareRequest(req.RequestID)
	if err != nil || rec == nil {
		http.Error(w, "request not found", http.StatusNotFound)
		return
	}

	_ = store.UpdateRequestStatus(req.RequestID, "denied", "", "Host denied request")
	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_DENIED",
		Actor:   "local",
		Profile: rec.ProfileName,
		Agent:   rec.Agent,
		Details: fmt.Sprintf("Host denied request %s from %s", req.RequestID, rec.PeerID),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleCancelShareRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	rec, _ := store.GetShareRequest(req.RequestID)

	_ = store.CancelShareRequest(req.RequestID)

	actor := "local"
	profile := ""
	agent := ""
	peerHost := ""
	if rec != nil {
		profile = rec.ProfileName
		agent = rec.Agent
		peerHost = rec.PeerIP
	}

	_ = store.AppendAudit(share.AuditEntry{
		Action:  "LEASE_CANCELLED",
		Actor:   actor,
		Profile: profile,
		Agent:   agent,
		Details: fmt.Sprintf("Cancelled lease request %s for %s", req.RequestID, profile),
	})

	// Forward cancellation to remote peer if known
	if peerHost != "" {
		go func(host, reqID string) {
			client := share.NewClient(4319)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = client.CancelLeaseRequest(ctx, host, reqID)
		}(peerHost, req.RequestID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handlePollShareRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RequestID string `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	rec, err := store.GetShareRequest(req.RequestID)
	if err != nil || rec == nil {
		http.Error(w, "request not found", http.StatusNotFound)
		return
	}

	if rec.Status == "approved" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "approved",
			"lease_id": rec.LeaseID,
			"profile":  rec.ProfileName,
			"agent":    rec.Agent,
		})
		return
	}
	if rec.Status == "denied" || rec.Status == "cancelled" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  rec.Status,
			"message": rec.ErrorMessage,
		})
		return
	}

	// Poll remote host
	client := share.NewClient(4319)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.PollLease(ctx, rec.PeerIP, req.RequestID)
	if err != nil {
		http.Error(w, "failed to poll peer: "+err.Error(), http.StatusBadGateway)
		return
	}

	if resp.Status == "approved" {
		mgr := share.NewLeaseManager(filepath.Join(config.BaseDir(), "profiles"))
		localName, err := mgr.InstallLease(rec.PeerID, rec.ProfileName, resp)
		if err == nil {
			_ = store.UpdateRequestStatus(req.RequestID, "approved", resp.LeaseID, "")
			_ = store.AppendAudit(share.AuditEntry{
				Action:  "LEASE_BORROWED",
				Actor:   rec.PeerID,
				Profile: rec.ProfileName,
				Agent:   resp.Agent,
				Details: fmt.Sprintf("Lease %s approved and installed as %s", resp.LeaseID, localName),
			})
		}
	} else if resp.Status == "denied" {
		_ = store.UpdateRequestStatus(req.RequestID, "denied", "", resp.Message)
		_ = store.AppendAudit(share.AuditEntry{
			Action:  "LEASE_DENIED",
			Actor:   rec.PeerID,
			Profile: rec.ProfileName,
			Agent:   rec.Agent,
			Details: fmt.Sprintf("Peer %s denied lease request %s", rec.PeerID, req.RequestID),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetShareAudit(w http.ResponseWriter, r *http.Request) {
	store := s.getShareStore()
	logs := store.GetAuditLogs(100)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)
}

func (s *Server) handleGetSchedules(w http.ResponseWriter, r *http.Request) {
	store := s.getShareStore()
	schedules := store.GetSchedules()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(schedules)
}

func (s *Server) handleSaveSchedule(w http.ResponseWriter, r *http.Request) {
	var sched share.SchedulePolicy
	if err := json.NewDecoder(r.Body).Decode(&sched); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if sched.Scope == "" {
		sched.Scope = "global"
	}
	if sched.Mode == "" {
		sched.Mode = "auto"
	}

	store := s.getShareStore()
	if sched.Scope != "global" && sched.Scope != "" {
		// Ensure profile is registered in shared_profiles so teammates can discover & borrow it
		all := store.GetSharedProfiles()
		exists := false
		for _, p := range all {
			if p.Name == sched.Scope {
				exists = true
				break
			}
		}
		if !exists {
			agent := "claude"
			profRoot := filepath.Join(config.BaseDir(), "profiles", sched.Scope)
			if entries, err := os.ReadDir(profRoot); err == nil {
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), ".claude") {
						agent = "claude"
					} else if strings.HasPrefix(e.Name(), ".codex") {
						agent = "codex"
					} else if strings.HasPrefix(e.Name(), ".gemini") {
						agent = "gemini"
					}
				}
			}
			_ = store.EnableProfile(sched.Scope, agent, sched.Mode == "auto", 4*time.Hour)
		}
	}

	if err := store.SaveSchedule(sched); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = store.AppendAudit(share.AuditEntry{
		Action:  "SCHEDULE_SAVED",
		Actor:   "local",
		Profile: sched.Scope,
		Details: fmt.Sprintf("Schedule saved for %s (%s to %s, mode=%s, enabled=%v)", sched.Scope, sched.StartTime, sched.EndTime, sched.Mode, sched.Enabled),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "schedule": sched})
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	store := s.getShareStore()
	if err := store.DeleteSchedule(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = store.AppendAudit(share.AuditEntry{
		Action:  "SCHEDULE_DELETED",
		Actor:   "local",
		Details: fmt.Sprintf("Schedule %s removed", req.ID),
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
