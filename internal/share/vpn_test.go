package share

import (
	"net"
	"testing"
)

func TestDetectVPNFromInterfaces(t *testing.T) {
	mockInterfaces := []net.Interface{
		{
			Index: 1,
			Name:  "lo0",
			Flags: net.FlagUp | net.FlagLoopback,
		},
		{
			Index: 2,
			Name:  "en0",
			Flags: net.FlagUp | net.FlagBroadcast,
		},
		{
			Index: 3,
			Name:  "utun3",
			Flags: net.FlagUp,
		},
	}

	mockAddrs := func(ifi net.Interface) ([]net.Addr, error) {
		if ifi.Name == "utun3" {
			ip, ipnet, _ := net.ParseCIDR("10.8.19.251/22")
			ipnet.IP = ip
			return []net.Addr{ipnet}, nil
		}
		if ifi.Name == "en0" {
			ip, ipnet, _ := net.ParseCIDR("192.168.1.50/24")
			ipnet.IP = ip
			return []net.Addr{ipnet}, nil
		}
		return nil, nil
	}

	status, err := DetectVPNFromInterfaces(mockInterfaces, mockAddrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Connected {
		t.Errorf("expected Connected=true, got false")
	}
	if status.InterfaceName != "utun3" {
		t.Errorf("expected utun3, got %s", status.InterfaceName)
	}
	if status.IP != "10.8.19.251" {
		t.Errorf("expected 10.8.19.251, got %s", status.IP)
	}
	if status.BroadcastIP != "10.8.19.255" {
		t.Errorf("expected 10.8.19.255, got %s", status.BroadcastIP)
	}
}

func TestDetectVPNDisconnected(t *testing.T) {
	mockInterfaces := []net.Interface{
		{
			Index: 1,
			Name:  "lo0",
			Flags: net.FlagUp | net.FlagLoopback,
		},
		{
			Index: 2,
			Name:  "en0",
			Flags: net.FlagUp | net.FlagBroadcast,
		},
	}

	mockAddrs := func(ifi net.Interface) ([]net.Addr, error) {
		return nil, nil
	}

	status, err := DetectVPNFromInterfaces(mockInterfaces, mockAddrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Connected {
		t.Errorf("expected Connected=false, got true")
	}
}
