package web

import (
	"net/url"
	"testing"

	"github.com/aim-cli/aim/internal/share"
)

func TestBuildPeerShareURL_IPv4IPv6AndHostnames(t *testing.T) {
	cases := []struct {
		name  string
		addr  string
		want  string
		valid bool
	}{
		{name: "ipv4", addr: "100.64.0.1", want: "http://100.64.0.1:4319/api/share/lease/revoke", valid: true},
		{name: "ipv4_with_port", addr: "100.64.0.1:9999", want: "http://100.64.0.1:4319/api/share/lease/revoke", valid: true},
		{name: "ipv6_bare", addr: "fd7a:115c:a1e0::1", want: "http://[fd7a:115c:a1e0::1]:4319/api/share/lease/revoke", valid: true},
		{name: "ipv6_bracketed_with_port", addr: "[fd7a:115c:a1e0::1]:4319", want: "http://[fd7a:115c:a1e0::1]:4319/api/share/lease/revoke", valid: true},
		{name: "ipv6_zoned", addr: "fe80::1%en0", want: "http://[fe80::1]:4319/api/share/lease/revoke", valid: true},
		{name: "hostname", addr: "alice-mac.local", want: "http://alice-mac.local:4319/api/share/lease/revoke", valid: true},
		{name: "empty", addr: "   ", valid: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := buildPeerShareURL(tc.addr, "4319", "/api/share/lease/revoke", "")
			if ok != tc.valid {
				t.Fatalf("ok=%v, want %v (got=%q)", ok, tc.valid, got)
			}
			if tc.valid && got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEntryMatchesPeer_CanonicalHostEquality(t *testing.T) {
	shortPeer := share.PeerInfo{HostName: "mac.local", VPNIP: "100.64.0.5"}
	if entryMatchesPeer("macbook-pro:work", "other-peer", shortPeer) {
		t.Fatal("substring host match collision: 'mac' must not match 'macbook-pro:work'")
	}

	fullPeer := share.PeerInfo{HostName: "MacBook-Pro.local", VPNIP: "100.64.0.9"}
	if !entryMatchesPeer("macbook-pro:work", "", fullPeer) {
		t.Fatal("expected canonical peer prefix match for 'macbook-pro:work'")
	}
	if !entryMatchesPeer("other:work", "100.64.0.9", fullPeer) {
		t.Fatal("expected exact VPNIP match for sourcePeer")
	}
}

func FuzzParsePeerHost(f *testing.F) {
	for _, seed := range []string{"", "100.64.0.1", "100.64.0.1:4319", "::1", "[::1]:4319", "fe80::1%en0", "host.local:4319"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, addr string) {
		got, ok := buildPeerShareURL(addr, "4319", "/api/share/lease/status", "lease_id=ls_1")
		if ok {
			if _, err := url.Parse(got); err != nil {
				t.Fatalf("buildPeerShareURL(%q) produced unparseable URL %q: %v", addr, got, err)
			}
		}
	})
}

func TestExtractPeerHost_RejectsMaliciousAndInvalidHosts(t *testing.T) {
	invalid := []string{
		"",
		"   ",
		".",
		"..",
		"alice@evil.com",
		"10.0.0.1/admin",
		"10.0.0.1?x=1",
		"10.0.0.1#frag",
		"bad host.local",
		"-bad.local",
		"bad-.local",
		"127.0.0.1%en0",
		"foo%2Fbar.local",
		"not_valid_underscore.local",
	}
	for _, raw := range invalid {
		if got, ok := extractPeerHost(raw); ok {
			t.Fatalf("expected extractPeerHost(%q) to reject, got (%q, true)", raw, got)
		}
	}
	if got, ok := extractPeerHost("[fe80::1%en0]:4319"); !ok || got != "fe80::1" {
		t.Fatalf("expected bracketed IPv6 zone host fe80::1, got (%q, %v)", got, ok)
	}
	if got, ok := extractPeerHost("bobs-mac.local:4319"); !ok || got != "bobs-mac.local" {
		t.Fatalf("expected bobs-mac.local, got (%q, %v)", got, ok)
	}
}
