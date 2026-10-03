package feedback

import (
	"context"
	"time"
)

// Category classifies the type of direct user feedback.
type Category string

const (
	CategoryBug     Category = "bug"
	CategoryFeature Category = "feature"
	CategoryGeneral Category = "general"
)

// Submission represents a qualitative feedback message from a user.
type Submission struct {
	Category      Category          `json:"category"`
	Message       string            `json:"message"`
	IncludeDoctor bool              `json:"include_doctor"`
	DoctorReport  string            `json:"doctor_report,omitempty"`
	AIMVersion    string            `json:"aim_version"`
	OS            string            `json:"os"`
	Arch          string            `json:"arch"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Dispatcher defines the interface for submitting user feedback.
type Dispatcher interface {
	Submit(ctx context.Context, sub Submission) error
	FallbackURL(sub Submission) string
}
