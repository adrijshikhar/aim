package share

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type VPNStatus struct {
	Connected     bool       `json:"connected"`
	InterfaceName string     `json:"interface_name"`
	IP            string     `json:"ip"`
	Subnet        *net.IPNet `json:"-"`
	BroadcastIP   string     `json:"broadcast_ip"`
}

func DetectVPN() (*VPNStatus, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	return DetectVPNFromInterfaces(ifaces, func(ifi net.Interface) ([]net.Addr, error) {
		return ifi.Addrs()
	})
}

func DetectVPNFromInterfaces(ifaces []net.Interface, addrsFunc func(net.Interface) ([]net.Addr, error)) (*VPNStatus, error) {
	for _, ifi := range ifaces {
		name := strings.ToLower(ifi.Name)
		if (strings.HasPrefix(name, "utun") || strings.HasPrefix(name, "tun")) && (ifi.Flags&net.FlagUp != 0) {
			addrs, err := addrsFunc(ifi)
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				var ipnet *net.IPNet
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
					ipnet = v
				case *net.IPAddr:
					ip = v.IP
				}
				if ip == nil || ip.IsLoopback() || ip.To4() == nil {
					continue
				}
				ipv4 := ip.To4()
				broadcast := ""
				if ipnet != nil && ipnet.Mask != nil && len(ipnet.Mask) == 4 {
					maskVal := binary.BigEndian.Uint32(ipnet.Mask)
					ipVal := binary.BigEndian.Uint32(ipv4)
					bcastVal := ipVal | ^maskVal
					bcastBytes := make([]byte, 4)
					binary.BigEndian.PutUint32(bcastBytes, bcastVal)
					broadcast = net.IP(bcastBytes).String()
				}
				return &VPNStatus{
					Connected:     true,
					InterfaceName: ifi.Name,
					IP:            ipv4.String(),
					Subnet:        ipnet,
					BroadcastIP:   broadcast,
				}, nil
			}
		}
	}
	return &VPNStatus{Connected: false}, nil
}
