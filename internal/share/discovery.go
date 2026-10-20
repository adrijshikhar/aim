package share

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type DiscoveryService struct {
	store      *ShareStore
	port       int
	httpClient *http.Client
}

func NewDiscovery(store *ShareStore, port int) *DiscoveryService {
	return &DiscoveryService{
		store: store,
		port:  port,
		httpClient: &http.Client{
			Timeout: 1500 * time.Millisecond,
		},
	}
}

type PingResponse struct {
	NodeID         string                `json:"node_id"`
	HostName       string                `json:"hostname"`
	VPNIP          string                `json:"vpn_ip"`
	Profiles       []string              `json:"profiles"`
	SharedProfiles []SharedProfileConfig `json:"shared_profiles,omitempty"`
}

func (d *DiscoveryService) PingPeerWithURL(ctx context.Context, rawURL string) (*PeerInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL+"/api/share/ping", nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var pr PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	peer := &PeerInfo{
		ID:             pr.NodeID,
		HostName:       pr.HostName,
		VPNIP:          pr.VPNIP,
		Port:           d.port,
		LastSeen:       time.Now(),
		Profiles:       pr.Profiles,
		SharedProfiles: pr.SharedProfiles,
	}
	_ = d.store.SavePeer(*peer)
	return peer, nil
}

func (d *DiscoveryService) PingPeer(ctx context.Context, ip string) (*PeerInfo, error) {
	url := fmt.Sprintf("http://%s:%d", ip, d.port)
	return d.PingPeerWithURL(ctx, url)
}

func (d *DiscoveryService) RefreshPeers(ctx context.Context) ([]PeerInfo, error) {
	known := d.store.GetPeers()
	if len(known) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	updated := make([]PeerInfo, 0, len(known))

	for _, p := range known {
		wg.Add(1)
		go func(peer PeerInfo) {
			defer wg.Done()
			target := peer.HostName
			if target == "" {
				target = peer.VPNIP
			}
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			res, err := d.PingPeer(pingCtx, target)
			if err != nil && !strings.Contains(target, ".") {
				res, err = d.PingPeer(pingCtx, target+".local")
			}
			mu.Lock()
			if err == nil && res != nil {
				updated = append(updated, *res)
			} else {
				updated = append(updated, peer)
			}
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return updated, nil
}
