package share

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestDiscoveryPingPeer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/share/ping" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"node_id":"alice","hostname":"alices-mac","vpn_ip":"10.8.16.9","profiles":["work"]}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	port, _ := strconv.Atoi(u.Port())

	store := NewStore(t.TempDir())
	discovery := NewDiscovery(store, port)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	peer, err := discovery.PingPeerWithURL(ctx, ts.URL)
	if err != nil {
		t.Fatalf("failed to ping peer: %v", err)
	}
	if peer.ID != "alice" || len(peer.Profiles) != 1 || peer.Profiles[0] != "work" {
		t.Errorf("unexpected peer info: %+v", peer)
	}

	peers := store.GetPeers()
	if len(peers) != 1 || peers[0].ID != "alice" {
		t.Errorf("expected peer saved in store, got %+v", peers)
	}
}
