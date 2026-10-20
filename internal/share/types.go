package share

import "time"

type SharedProfileConfig struct {
	Name        string        `json:"name"`
	Agent       string        `json:"agent"`
	AutoApprove bool          `json:"auto_approve"`
	MaxLeaseTTL time.Duration `json:"max_lease_ttl"`
	Enabled     bool          `json:"enabled"`
}

type PeerInfo struct {
	ID             string                `json:"id"`
	HostName       string                `json:"hostname"`
	VPNIP          string                `json:"vpn_ip"`
	Port           int                   `json:"port"`
	LastSeen       time.Time             `json:"last_seen"`
	Profiles       []string              `json:"profiles"`
	SharedProfiles []SharedProfileConfig `json:"shared_profiles,omitempty"`
}

type LeaseRecord struct {
	ID          string    `json:"id"`
	RequesterID string    `json:"requester_id"`
	RequesterIP string    `json:"requester_ip"`
	ProfileName string    `json:"profile_name"`
	Agent       string    `json:"agent"`
	GrantedAt   time.Time `json:"granted_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Active      bool      `json:"active"`
}

type SharePayload struct {
	Files map[string]string `json:"files"` // relative_path -> base64_content
}

type LeaseRequest struct {
	RequesterID    string `json:"requester_id"`
	RequesterIP    string `json:"requester_ip"`
	ProfileName    string `json:"profile_name"`
	Agent          string `json:"agent"`
	RequestedHours int    `json:"requested_hours"`
}

type LeaseResponse struct {
	Status      string        `json:"status"` // "approved", "pending", "denied"
	RequestID   string        `json:"request_id,omitempty"`
	LeaseID     string        `json:"lease_id,omitempty"`
	ProfileName string        `json:"profile_name,omitempty"`
	Agent       string        `json:"agent,omitempty"`
	ExpiresAt   time.Time     `json:"expires_at,omitempty"`
	Payload     *SharePayload `json:"payload,omitempty"`
	Message     string        `json:"message,omitempty"`
}
