package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/share"
)

type ShareStatusResponse struct {
	VPN            *share.VPNStatus            `json:"vpn"`
	SharedProfiles []share.SharedProfileConfig `json:"shared_profiles"`
	Peers          []share.PeerInfo            `json:"peers"`
	ActiveLeases   []ActiveLeaseInfo           `json:"active_leases"`
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
	store := share.NewStore(config.BaseDir())
	shared := store.GetSharedProfiles()
	peers := store.GetPeers()

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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ShareStatusResponse{
		VPN:            vpn,
		SharedProfiles: shared,
		Peers:          peers,
		ActiveLeases:   activeLeases,
	})
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

	store := share.NewStore(config.BaseDir())
	if !req.Enabled {
		if err := store.DisableProfile(req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "enabled": true})
}

func (s *Server) handleGetSharePeers(w http.ResponseWriter, r *http.Request) {
	store := share.NewStore(config.BaseDir())
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

	store := share.NewStore(config.BaseDir())
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

	store := share.NewStore(config.BaseDir())
	targetHost := req.Peer
	for _, p := range store.GetPeers() {
		shortHost := strings.TrimSuffix(strings.ToLower(p.HostName), ".local")
		if strings.EqualFold(p.HostName, req.Peer) || strings.EqualFold(shortHost, strings.ToLower(req.Peer)) || p.VPNIP == req.Peer {
			if p.HostName != "" {
				targetHost = p.HostName
			} else {
				targetHost = p.VPNIP
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

	if resp.Status != "approved" {
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":            true,
		"local_profile": localName,
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
