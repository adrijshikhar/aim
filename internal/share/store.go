package share

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	mu           sync.RWMutex
}

func NewStore(aimDir string) *ShareStore {
	return &ShareStore{
		aimDir:       aimDir,
		shareFile:    filepath.Join(aimDir, "share.json"),
		peerFile:     filepath.Join(aimDir, "peers.json"),
		leaseFile:    filepath.Join(aimDir, "leases.json"),
		grantedFile:  filepath.Join(aimDir, "granted_leases.json"),
		auditFile:    filepath.Join(aimDir, "share_audit.json"),
		scheduleFile: filepath.Join(aimDir, "share_schedules.json"),
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
			configs[i].MaxLeaseTTL = ttl
			configs[i].Enabled = true
			found = true
			break
		}
	}
	if !found {
		configs = append(configs, SharedProfileConfig{
			Name:        profileName,
			Agent:       agent,
			AutoApprove: autoApprove,
			MaxLeaseTTL: ttl,
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
		return nil
	}
	var configs []SharedProfileConfig
	_ = json.Unmarshal(data, &configs)
	var active []SharedProfileConfig
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
		if peers[i].VPNIP == peer.VPNIP || (peer.ID != "" && peers[i].ID == peer.ID) {
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
		return nil
	}
	var peers []PeerInfo
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

	found := false
	for i := range leases {
		if leases[i].LeaseID == lease.LeaseID {
			leases[i] = lease
			found = true
			break
		}
	}
	if !found {
		leases = append(leases, lease)
	}

	_ = os.MkdirAll(s.aimDir, 0700)
	data, err := json.MarshalIndent(leases, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", s.grantedFile, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.grantedFile)
}

func (s *ShareStore) GetGrantedLeases() []GrantedLease {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.grantedFile)
	if err != nil {
		return nil
	}
	var leases []GrantedLease
	if err := json.Unmarshal(data, &leases); err != nil {
		return nil
	}
	return leases
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
