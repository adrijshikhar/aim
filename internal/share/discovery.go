package share

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
