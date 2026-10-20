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
	aimDir    string
	shareFile string
	peerFile  string
	leaseFile string
	mu        sync.RWMutex
}

func NewStore(aimDir string) *ShareStore {
	return &ShareStore{
		aimDir:    aimDir,
		shareFile: filepath.Join(aimDir, "share.json"),
		peerFile:  filepath.Join(aimDir, "peers.json"),
		leaseFile: filepath.Join(aimDir, "leases.json"),
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

func (s *ShareStore) loadPeersLocked() []PeerInfo {
	data, err := os.ReadFile(s.peerFile)
	if err != nil {
		return nil
	}
	var peers []PeerInfo
	_ = json.Unmarshal(data, &peers)
	return peers
}
