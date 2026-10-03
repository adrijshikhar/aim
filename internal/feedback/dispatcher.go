package feedback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultEndpoint can be set at build time or via environment variable.
var DefaultEndpoint = ""

// HTTPDispatcher sends feedback submissions to an ingest endpoint via HTTPS POST,
// and can generate pre-filled GitHub issue URLs as a local fallback.
type HTTPDispatcher struct {
	Endpoint   string
	HTTPClient *http.Client
}

// NewDispatcher creates a new HTTPDispatcher using configured or default endpoints.
func NewDispatcher(endpoint string) *HTTPDispatcher {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	return &HTTPDispatcher{
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Submit posts the feedback submission payload to the configured endpoint.
func (d *HTTPDispatcher) Submit(ctx context.Context, sub Submission) error {
	if d.Endpoint == "" {
		return fmt.Errorf("no feedback endpoint configured")
	}

	payload, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("failed to encode feedback payload: %w", err)
	}

	client := d.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create feedback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("aim-feedback/%s (%s; %s)", sub.AIMVersion, sub.OS, sub.Arch))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit feedback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("feedback server responded with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// FallbackURL builds a pre-filled GitHub issue URL from the feedback submission.
func (d *HTTPDispatcher) FallbackURL(sub Submission) string {
	baseURL := "https://github.com/adrijshikhar/aim/issues/new"

	firstLine := sub.Message
	if idx := strings.Index(firstLine, "\n"); idx != -1 {
		firstLine = firstLine[:idx]
	}
	firstLine = strings.TrimSpace(firstLine)
	if len(firstLine) > 60 {
		firstLine = firstLine[:60] + "..."
	}

	title := fmt.Sprintf("[%s] %s", strings.Title(string(sub.Category)), firstLine)
	if firstLine == "" {
		title = fmt.Sprintf("[%s] Feedback", strings.Title(string(sub.Category)))
	}

	var body strings.Builder
	body.WriteString("### User Feedback\n")
	body.WriteString(fmt.Sprintf("- **Category:** %s\n", sub.Category))
	if sub.AIMVersion != "" {
		body.WriteString(fmt.Sprintf("- **AIM Version:** %s (%s/%s)\n", sub.AIMVersion, sub.OS, sub.Arch))
	}
	body.WriteString("\n**Message:**\n")
	body.WriteString(sub.Message)
	body.WriteString("\n")

	if sub.IncludeDoctor && sub.DoctorReport != "" {
		body.WriteString("\n<details>\n<summary>Diagnostic Report (aim doctor --report)</summary>\n\n")
		body.WriteString(sub.DoctorReport)
		body.WriteString("\n</details>\n")
	}

	bodyStr := body.String()
	if len(bodyStr) > 4000 {
		bodyStr = bodyStr[:4000] + "\n\n... (truncated for URL length)"
	}

	params := url.Values{}
	params.Set("title", title)
	params.Set("body", bodyStr)
	params.Set("labels", "feedback,"+string(sub.Category))

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}
