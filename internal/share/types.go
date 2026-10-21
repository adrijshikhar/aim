package share

import (
	"fmt"
	"time"
)

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

type PendingRequestInfo struct {
	RequestID string       `json:"request_id"`
	Request   LeaseRequest `json:"request"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type ShareRequestRecord struct {
	ID             string    `json:"id"`
	Direction      string    `json:"direction"` // "incoming" or "outgoing"
	PeerID         string    `json:"peer_id"`
	PeerIP         string    `json:"peer_ip"`
	ProfileName    string    `json:"profile_name"`
	Agent          string    `json:"agent"`
	RequestedHours int       `json:"requested_hours"`
	Status         string    `json:"status"` // "pending", "approved", "denied", "cancelled", "expired"
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	LeaseID        string    `json:"lease_id,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

func (r *ShareRequestRecord) ToPendingInfo() PendingRequestInfo {
	return PendingRequestInfo{
		RequestID: r.ID,
		Request: LeaseRequest{
			RequesterID:    r.PeerID,
			RequesterIP:    r.PeerIP,
			ProfileName:    r.ProfileName,
			Agent:          r.Agent,
			RequestedHours: r.RequestedHours,
		},
		CreatedAt: r.CreatedAt,
		ExpiresAt: r.ExpiresAt,
	}
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

type GrantedLease struct {
	LeaseID      string    `json:"lease_id"`
	ProfileName  string    `json:"profile_name"`
	Agent        string    `json:"agent"`
	BorrowerID   string    `json:"borrower_id"`
	BorrowerIP   string    `json:"borrower_ip"`
	GrantedAt    time.Time `json:"granted_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Revoked      bool      `json:"revoked"`
	AutoApproved bool      `json:"auto_approved"`
}

type SchedulePolicy struct {
	ID        string    `json:"id"`
	Enabled   bool      `json:"enabled"`
	Scope     string    `json:"scope"`                // "global" or profile name e.g. "office"
	Mode      string    `json:"mode"`                 // "auto" or "request"
	StartTime string    `json:"start_time"`           // "HH:MM" e.g. "00:00"
	EndTime   string    `json:"end_time"`             // "HH:MM" e.g. "08:00"
	StartDate string    `json:"start_date,omitempty"` // "YYYY-MM-DD"
	EndDate   string    `json:"end_date,omitempty"`   // "YYYY-MM-DD"
	CreatedAt time.Time `json:"created_at"`
}

func (p *SchedulePolicy) IsActive(t time.Time) bool {
	if !p.Enabled {
		return false
	}
	// Case 1: Specific Date Range if StartDate and EndDate provided
	if p.StartDate != "" && p.EndDate != "" {
		startStr := p.StartDate + "T" + p.StartTime
		endStr := p.EndDate + "T" + p.EndTime
		if len(p.StartTime) == 5 {
			startStr += ":00"
		}
		if len(p.EndTime) == 5 {
			endStr += ":00"
		}
		start, err1 := time.ParseInLocation("2006-01-02T15:04:05", startStr, t.Location())
		end, err2 := time.ParseInLocation("2006-01-02T15:04:05", endStr, t.Location())
		if err1 == nil && err2 == nil {
			return (t.Equal(start) || t.After(start)) && t.Before(end)
		}
	}

	// Case 2: Daily time-of-day window (e.g. 00:00 to 08:00 or 12am to 8am)
	if p.StartTime != "" && p.EndTime != "" {
		var sh, sm, eh, em int
		_, errS := fmt.Sscanf(p.StartTime, "%d:%d", &sh, &sm)
		_, errE := fmt.Sscanf(p.EndTime, "%d:%d", &eh, &em)
		if errS == nil && errE == nil {
			nowMinutes := t.Hour()*60 + t.Minute()
			startMinutes := sh*60 + sm
			endMinutes := eh*60 + em

			if startMinutes <= endMinutes {
				return nowMinutes >= startMinutes && nowMinutes < endMinutes
			}
			// Overnight window, e.g. 22:00 to 06:00
			return nowMinutes >= startMinutes || nowMinutes < endMinutes
		}
	}

	return false
}

type AuditEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"` // "LEASE_REQUESTED", "LEASE_GRANTED", "LEASE_DENIED", "LEASE_REVOKED", "GLOBAL_MODE"
	Actor     string    `json:"actor"`  // peer hostname or "local"
	Profile   string    `json:"profile,omitempty"`
	Agent     string    `json:"agent,omitempty"`
	Details   string    `json:"details"`
}
