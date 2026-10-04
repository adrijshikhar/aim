package telemetry

import "time"

const (
	EventCommandExecuted       = "command_executed"
	EventTUIOpened             = "tui_opened"
	EventDoctorReportGenerated = "doctor_report_generated"
	EventFeedbackSubmitted     = "feedback_submitted"
)

// Event represents an individual telemetry event payload.
type Event struct {
	EventName  string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Properties map[string]any `json:"properties"`
}

// IsAllowedEvent validates whether an event name is within the sanctioned AIM telemetry schema.
func IsAllowedEvent(name string) bool {
	switch name {
	case EventCommandExecuted, EventTUIOpened, EventDoctorReportGenerated, EventFeedbackSubmitted, "agent_run_completed", "user_identified":
		return true
	default:
		return false
	}
}

// NewEvent constructs an anonymous event with standard system properties.
func NewEvent(name, distinctID, version, goos, arch string, customProps map[string]any) Event {
	props := make(map[string]any, len(customProps)+5)
	for k, v := range customProps {
		props[k] = v
	}
	// System properties are immutable and cannot be overridden by custom properties (anti-spoofing)
	props["version"] = version
	props["aim_version"] = version
	props["os"] = goos
	props["arch"] = arch

	return Event{
		EventName:  name,
		DistinctID: distinctID,
		Timestamp:  time.Now().UTC(),
		Properties: props,
	}
}
