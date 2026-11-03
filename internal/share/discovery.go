package share

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
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
	PublicKey      ed25519.PublicKey     `json:"public_key,omitempty"`
}

type pingResponseWire struct {
	NodeID         string                `json:"node_id"`
	HostName       string                `json:"hostname"`
	VPNIP          string                `json:"vpn_ip"`
	Profiles       []string              `json:"profiles"`
	SharedProfiles []SharedProfileConfig `json:"shared_profiles,omitempty"`
	PublicKey      string                `json:"public_key,omitempty"`
}

func (pr PingResponse) MarshalJSON() ([]byte, error) {
	var pubKeyHex string
	if len(pr.PublicKey) > 0 {
		pubKeyHex = hex.EncodeToString(pr.PublicKey)
	}
	return json.Marshal(pingResponseWire{
		NodeID:         pr.NodeID,
		HostName:       pr.HostName,
		VPNIP:          pr.VPNIP,
		Profiles:       pr.Profiles,
		SharedProfiles: pr.SharedProfiles,
		PublicKey:      pubKeyHex,
	})
}

func (pr *PingResponse) UnmarshalJSON(data []byte) error {
	type rawPing struct {
		NodeID         string                `json:"node_id"`
		HostName       string                `json:"hostname"`
		VPNIP          string                `json:"vpn_ip"`
		Profiles       []string              `json:"profiles"`
		SharedProfiles []SharedProfileConfig `json:"shared_profiles,omitempty"`
		RawPubKey      any                   `json:"public_key,omitempty"`
	}
	var raw rawPing
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	pr.NodeID = raw.NodeID
	pr.HostName = raw.HostName
	pr.VPNIP = raw.VPNIP
	pr.Profiles = raw.Profiles
	pr.SharedProfiles = raw.SharedProfiles
	pr.PublicKey = nil

	if raw.RawPubKey != nil {
		switch v := raw.RawPubKey.(type) {
		case string:
			if b, err := hex.DecodeString(v); err == nil && len(b) == ed25519.PublicKeySize {
				pr.PublicKey = ed25519.PublicKey(b)
			} else if b, err := base64.StdEncoding.DecodeString(v); err == nil && len(b) == ed25519.PublicKeySize {
				pr.PublicKey = ed25519.PublicKey(b)
			}
		case []any:
			if len(v) == ed25519.PublicKeySize {
				b := make([]byte, len(v))
				for i, num := range v {
					if n, ok := num.(float64); ok {
						b[i] = byte(n)
					}
				}
				pr.PublicKey = ed25519.PublicKey(b)
			}
		}
	}
	return nil
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
		PublicKey:      pr.PublicKey,
	}
	_ = d.store.SavePeer(*peer)
	return peer, nil
}

func (d *DiscoveryService) PingPeer(ctx context.Context, target string) (*PeerInfo, error) {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return d.PingPeerWithURL(ctx, target)
	}
	if host, port, err := net.SplitHostPort(target); err == nil {
		return d.PingPeerWithURL(ctx, fmt.Sprintf("http://%s", net.JoinHostPort(host, port)))
	}
	url := fmt.Sprintf("http://%s", net.JoinHostPort(target, strconv.Itoa(d.port)))
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
			hostOnly := target
			if h, _, err := net.SplitHostPort(target); err == nil {
				hostOnly = h
			}
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			res, err := d.PingPeer(pingCtx, hostOnly)
			if err != nil && peer.VPNIP != "" && peer.VPNIP != target {
				vpnHostOnly := peer.VPNIP
				if vh, _, err := net.SplitHostPort(peer.VPNIP); err == nil {
					vpnHostOnly = vh
				}
				res, err = d.PingPeer(pingCtx, vpnHostOnly)
			}
			if err != nil && !strings.Contains(hostOnly, ".") {
				res, err = d.PingPeer(pingCtx, hostOnly+".local")
			}
			mu.Lock()
			if err == nil && res != nil {
				updated = append(updated, *res)
				_ = d.store.SavePeer(*res)
			} else {
				updated = append(updated, peer)
			}
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return updated, nil
}
