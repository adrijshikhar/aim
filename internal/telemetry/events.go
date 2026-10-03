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

// NewEvent constructs an anonymous event with standard system properties.
func NewEvent(name, distinctID, version, goos, arch string, customProps map[string]any) Event {
	props := make(map[string]any, len(customProps)+4)
	props["aim_version"] = version
	props["os"] = goos
	props["arch"] = arch
	for k, v := range customProps {
		props[k] = v
	}

	return Event{
		EventName:  name,
		DistinctID: distinctID,
		Timestamp:  time.Now().UTC(),
		Properties: props,
	}
}
