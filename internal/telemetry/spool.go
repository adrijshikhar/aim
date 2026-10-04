package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BatchPayload represents the envelope sent to batch analytics ingest APIs like PostHog.
type BatchPayload struct {
	APIKey string  `json:"api_key,omitempty"`
	Batch  []Event `json:"batch"`
}

// Spooler manages appending events to a local JSONL file and batch flushing over HTTPS.
type Spooler struct {
	mu        sync.Mutex
	SpoolPath string
	Endpoint  string
	APIKey    string
	Client    *http.Client
}

// NewSpooler constructs a Spooler.
func NewSpooler(spoolPath, endpoint, apiKey string, httpClient *http.Client) *Spooler {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Second}
	}
	return &Spooler{
		SpoolPath: spoolPath,
		Endpoint:  endpoint,
		APIKey:    apiKey,
		Client:    httpClient,
	}
}

// Append writes an event as a line in the spool file.
func (s *Spooler) Append(ev Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = os.MkdirAll(filepath.Dir(s.SpoolPath), 0755)
	f, err := os.OpenFile(s.SpoolPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = f.Write(data)
	return err
}

// Read loads all un-flushed events from the spool file.
func (s *Spooler) Read() ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.SpoolPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var events []Event
	lines := bytes.Split(data, []byte("\n"))
	for _, l := range lines {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(l, &ev); err == nil {
			events = append(events, ev)
		}
	}
	return events, nil
}

// Clear removes the spool file.
func (s *Spooler) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.SpoolPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Flush submits all queued events in a single batch request and clears the spool on success.
func (s *Spooler) Flush(ctx context.Context) error {
	events, err := s.Read()
	if err != nil || len(events) == 0 {
		return nil
	}
	if s.Endpoint == "" {
		// If no remote endpoint is set, keep spool bounded
		if len(events) > 100 {
			_ = s.Clear()
		}
		return nil
	}

	var payload []byte
	if s.APIKey != "" {
		payload, err = json.Marshal(BatchPayload{
			APIKey: s.APIKey,
			Batch:  events,
		})
	} else {
		payload, err = json.Marshal(events)
	}
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "aim-telemetry")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_ = s.Clear()
		return nil
	}

	return fmt.Errorf("telemetry endpoint returned status: %d", resp.StatusCode)
}
