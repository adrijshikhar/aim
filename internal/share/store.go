package share

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ShareStore struct {
	aimDir       string
	shareFile    string
	peerFile     string
	leaseFile    string
	grantedFile  string
	auditFile    string
	scheduleFile string
	db           *ShareDB
	mu           sync.RWMutex
}

func NewStore(aimDir string) *ShareStore {
	db, _ := NewShareDB(aimDir)
	return &ShareStore{
		aimDir:       aimDir,
		shareFile:    filepath.Join(aimDir, "share.json"),
		peerFile:     filepath.Join(aimDir, "peers.json"),
		leaseFile:    filepath.Join(aimDir, "leases.json"),
		grantedFile:  filepath.Join(aimDir, "granted_leases.json"),
		auditFile:    filepath.Join(aimDir, "share_audit.json"),
		scheduleFile: filepath.Join(aimDir, "share_schedules.json"),
		db:           db,
	}
}

func (s *ShareStore) EnableProfile(profileName, agent string, autoApprove bool, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs := s.loadSharesLocked()
	found := false
	for i := range configs {
		if configs[i].Name == profileName {
			configs[i].Agent = agent
			configs[i].AutoApprove = autoApprove
			if ttl > 0 {
				configs[i].MaxLeaseTTL = ttl
			} else if configs[i].MaxLeaseTTL <= 0 {
				configs[i].MaxLeaseTTL = 4 * time.Hour
			}
			configs[i].Enabled = true
			found = true
			break
		}
	}
	if !found {
		effTTL := ttl
		if effTTL <= 0 {
			effTTL = 4 * time.Hour
		}
		configs = append(configs, SharedProfileConfig{
			Name:        profileName,
			Agent:       agent,
			AutoApprove: autoApprove,
			MaxLeaseTTL: effTTL,
			Enabled:     true,
		})
	}
	return s.saveSharesLocked(configs)
}

func (s *ShareStore) DisableProfile(profileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	configs := s.loadSharesLocked()
	filtered := make([]SharedProfileConfig, 0, len(configs))
	for _, c := range configs {
		if c.Name != profileName {
			filtered = append(filtered, c)
		}
	}
	return s.saveSharesLocked(filtered)
}

func (s *ShareStore) GetSharedProfiles() []SharedProfileConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadSharesLocked()
}

func (s *ShareStore) loadSharesLocked() []SharedProfileConfig {
	data, err := os.ReadFile(s.shareFile)
	if err != nil {
		return []SharedProfileConfig{}
	}
	var configs []SharedProfileConfig
	_ = json.Unmarshal(data, &configs)
	active := make([]SharedProfileConfig, 0)
	for _, c := range configs {
		if c.Enabled {
			active = append(active, c)
		}
	}
	return active
}

func (s *ShareStore) saveSharesLocked(configs []SharedProfileConfig) error {
	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.shareFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.shareFile)
}

func (s *ShareStore) SavePeer(peer PeerInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	peers := s.loadPeersLocked()
	found := false
	for i := range peers {
		match := false
		if peer.ID != "" && peers[i].ID == peer.ID {
			match = true
		} else if peer.HostName != "" && strings.EqualFold(peers[i].HostName, peer.HostName) {
			match = true
		} else if peer.VPNIP != "" && (peers[i].VPNIP == peer.VPNIP || strings.Split(peers[i].VPNIP, ":")[0] == strings.Split(peer.VPNIP, ":")[0]) {
			match = true
		}
		if match {
			peers[i] = peer
			found = true
			break
		}
	}
	if !found {
		peers = append(peers, peer)
	}

	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.peerFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.peerFile)
}

func (s *ShareStore) GetPeers() []PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadPeersLocked()
}

func (s *ShareStore) GetGlobalMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(filepath.Join(s.aimDir, "share_global.json"))
	if err != nil {
		return "off"
	}
	var m struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(data, &m); err != nil || m.Mode == "" {
		return "off"
	}
	return m.Mode
}

func (s *ShareStore) SetGlobalMode(mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if mode != "off" && mode != "request" && mode != "auto" {
		mode = "off"
	}
	_ = os.MkdirAll(s.aimDir, 0700)
	target := filepath.Join(s.aimDir, "share_global.json")
	data, err := json.MarshalIndent(map[string]string{"mode": mode}, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", target, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func (s *ShareStore) loadPeersLocked() []PeerInfo {
	data, err := os.ReadFile(s.peerFile)
	if err != nil {
		return []PeerInfo{}
	}
	peers := make([]PeerInfo, 0)
	_ = json.Unmarshal(data, &peers)
	return peers
}

func (s *ShareStore) SaveGrantedLease(lease GrantedLease) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var leases []GrantedLease
	if data, err := os.ReadFile(s.grantedFile); err == nil {
		_ = json.Unmarshal(data, &leases)
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var retained []GrantedLease
	found := false
	for _, l := range leases {
		if l.LeaseID == lease.LeaseID {
			retained = append(retained, lease)
			found = true
		} else if l.GrantedAt.After(cutoff) || l.ExpiresAt.After(cutoff) {
			retained = append(retained, l)
		}
	}
	if !found {
		retained = append(retained, lease)
	}
	if len(retained) > 100 {
		retained = retained[len(retained)-100:]
	}

	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(retained, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.grantedFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.grantedFile)
}

func (s *ShareStore) pruneExpiredGrantedLeasesLocked() ([]GrantedLease, int) {
	data, err := os.ReadFile(s.grantedFile)
	if err != nil {
		return nil, 0
	}
	var leases []GrantedLease
	if err := json.Unmarshal(data, &leases); err != nil {
		return nil, 0
	}
	now := time.Now()
	var active []GrantedLease
	pruned := 0
	for _, l := range leases {
		if !l.Revoked && now.Before(l.ExpiresAt) {
			active = append(active, l)
		} else {
			pruned++
		}
	}
	return active, pruned
}

// GetGrantedLeases returns all retained granted leases (active, revoked, and recent expired history within 7 days).
func (s *ShareStore) GetGrantedLeases() []GrantedLease {
	s.mu.Lock()
	defer s.mu.Unlock()

	leases := make([]GrantedLease, 0)
	if data, err := os.ReadFile(s.grantedFile); err == nil {
		_ = json.Unmarshal(data, &leases)
	}
	return leases
}

// GetActiveGrantedLeases returns only currently active, non-revoked leases.
func (s *ShareStore) GetActiveGrantedLeases() []GrantedLease {
	s.mu.Lock()
	defer s.mu.Unlock()

	active, _ := s.pruneExpiredGrantedLeasesLocked()
	return active
}

func (s *ShareStore) PruneExpiredGrantedLeases() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, pruned := s.pruneExpiredGrantedLeasesLocked()
	return pruned
}

func (s *ShareStore) RevokeGrantedLease(leaseID string) (*GrantedLease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var leases []GrantedLease
	if data, err := os.ReadFile(s.grantedFile); err == nil {
		_ = json.Unmarshal(data, &leases)
	}

	var revoked *GrantedLease
	for i := range leases {
		if leases[i].LeaseID == leaseID {
			leases[i].Revoked = true
			revoked = &leases[i]
			break
		}
	}
	if revoked == nil {
		return nil, fmt.Errorf("lease %s not found", leaseID)
	}

	data, err := json.MarshalIndent(leases, "", "  ")
	if err != nil {
		return nil, err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.grantedFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, s.grantedFile); err != nil {
		return nil, err
	}
	return revoked, nil
}

func (s *ShareStore) AppendAudit(entry AuditEntry) error {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.db.AppendAudit(ctx, entry); err == nil {
			return nil
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("aud_%d", time.Now().UnixNano())
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	var logs []AuditEntry
	if data, err := os.ReadFile(s.auditFile); err == nil {
		_ = json.Unmarshal(data, &logs)
	}

	// Prepend so newest is first, bounded to 500 entries
	logs = append([]AuditEntry{entry}, logs...)
	if len(logs) > 500 {
		logs = logs[:500]
	}

	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.auditFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.auditFile)
}

func (s *ShareStore) GetAuditLogs(limit int) []AuditEntry {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if logs, err := s.db.GetAuditLogs(ctx, limit); err == nil && len(logs) > 0 {
			return logs
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.auditFile)
	if err != nil {
		return nil
	}
	var logs []AuditEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil
	}
	if limit > 0 && len(logs) > limit {
		return logs[:limit]
	}
	return logs
}

// SaveShareRequest persists a lease request (incoming or outgoing) in SQLite.
func (s *ShareStore) SaveShareRequest(req ShareRequestRecord) error {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.db.SaveRequest(ctx, req)
	}
	return nil
}

// GetShareRequest retrieves a request by ID.
func (s *ShareStore) GetShareRequest(id string) (*ShareRequestRecord, error) {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.db.GetRequest(ctx, id)
	}
	return nil, nil
}

// GetPendingIncomingRequests returns all active incoming requests awaiting manual host approval.
func (s *ShareStore) GetPendingIncomingRequests() []ShareRequestRecord {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if reqs, err := s.db.GetPendingIncomingRequests(ctx); err == nil {
			return reqs
		}
	}
	return nil
}

// GetPendingOutgoingRequests returns all active outgoing requests sent by this node.
func (s *ShareStore) GetPendingOutgoingRequests() []ShareRequestRecord {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if reqs, err := s.db.GetPendingOutgoingRequests(ctx); err == nil {
			return reqs
		}
	}
	return nil
}

// UpdateRequestStatus updates status, lease ID, and error message for a request.
func (s *ShareStore) UpdateRequestStatus(id, status, leaseID, errMsg string) error {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.db.UpdateRequestStatus(ctx, id, status, leaseID, errMsg)
	}
	return nil
}

// TransitionRequestStatus atomically updates status from fromStatus to toStatus.
// Returns (true, nil) if successfully transitioned, or (false, nil) if the request was not in fromStatus.
func (s *ShareStore) TransitionRequestStatus(id, fromStatus, toStatus, leaseID, errMsg string) (bool, error) {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.db.TransitionRequestStatus(ctx, id, fromStatus, toStatus, leaseID, errMsg)
	}
	return true, nil
}

// CancelShareRequest marks a request as cancelled if it is currently pending.
func (s *ShareStore) CancelShareRequest(id string) error {
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return s.db.CancelRequest(ctx, id)
	}
	return nil
}

// Close releases the SQLite connection pool if open.
func (s *ShareStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *ShareStore) GetSchedules() []SchedulePolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadSchedulesLocked()
}

func (s *ShareStore) loadSchedulesLocked() []SchedulePolicy {
	data, err := os.ReadFile(s.scheduleFile)
	if err != nil {
		return nil
	}
	var list []SchedulePolicy
	_ = json.Unmarshal(data, &list)
	return list
}

func (s *ShareStore) SaveSchedule(sched SchedulePolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.loadSchedulesLocked()
	found := false
	if sched.ID == "" {
		sched.ID = fmt.Sprintf("sch_%d", time.Now().UnixNano())
	}
	if sched.CreatedAt.IsZero() {
		sched.CreatedAt = time.Now()
	}
	for i := range list {
		if list[i].ID == sched.ID {
			list[i] = sched
			found = true
			break
		}
	}
	if !found {
		list = append(list, sched)
	}

	return s.saveSchedulesLocked(list)
}

func (s *ShareStore) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.loadSchedulesLocked()
	var updated []SchedulePolicy
	for _, sc := range list {
		if sc.ID != id {
			updated = append(updated, sc)
		}
	}
	return s.saveSchedulesLocked(updated)
}

func (s *ShareStore) saveSchedulesLocked(list []SchedulePolicy) error {
	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.scheduleFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.scheduleFile)
}

// GetEffectiveGlobalMode returns the active global mode considering scheduled policies.
func (s *ShareStore) GetEffectiveGlobalMode() string {
	now := time.Now()
	// Check global schedules first
	schedules := s.GetSchedules()
	for _, sc := range schedules {
		if sc.Scope == "global" && sc.IsActive(now) {
			return sc.Mode
		}
	}
	return s.GetGlobalMode()
}

// GetEffectiveProfileMode returns whether a profile is shareable and if it is auto-approved.
func (s *ShareStore) GetEffectiveProfileMode(profileName string) (enabled bool, autoApprove bool) {
	now := time.Now()
	gMode := s.GetEffectiveGlobalMode()

	// Check if this specific profile has an active schedule
	schedules := s.GetSchedules()
	for _, sc := range schedules {
		if sc.Scope == profileName && sc.IsActive(now) {
			return true, sc.Mode == "auto"
		}
	}

	// If global mode is off and no active schedule overrides it:
	if gMode == "off" {
		return false, false
	}

	// Look up profile configuration in shared_profiles
	profiles := s.GetSharedProfiles()
	for _, p := range profiles {
		if p.Name == profileName {
			if !p.Enabled {
				return false, false
			}
			if gMode == "auto" {
				return true, true
			}
			return true, p.AutoApprove
		}
	}

	return false, false
}
