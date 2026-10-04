package telemetry

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
)

const (
	// DefaultPostHogHost is the default production PostHog host URL for AIM.
	DefaultPostHogHost = "https://us.i.posthog.com"
	// DefaultPostHogEndpoint is the default production PostHog batch ingest endpoint.
	DefaultPostHogEndpoint = DefaultPostHogHost + "/batch/"
	// DefaultPostHogAPIKey is the default production PostHog project API key for AIM.
	DefaultPostHogAPIKey = "phc_tTNz7hfe9r6HMVMR5fjrJVEBoaGvZ4V9onEhZrF74Ggs"
)

// Client defines the interface for emitting anonymous telemetry events.
type Client interface {
	Track(event string, properties map[string]any)
	Flush(ctx context.Context) error
	Close() error
}

// DefaultClient is the production telemetry client.
type DefaultClient struct {
	baseDir    string
	cacheDir   string
	version    string
	distinctID string
	enabled    bool
	spooler    *Spooler
}

func isRunningInTest() bool {
	return strings.HasSuffix(os.Args[0], ".test") || flag.Lookup("test.v") != nil
}

// NewClient constructs a new telemetry client with privacy and opt-out checks applied.
func NewClient(baseDir, cacheDir, version string, cfg *config.Config) *DefaultClient {
	enabled := IsTelemetryEnabled(cfg)
	spoolPath := filepath.Join(cacheDir, "telemetry_spool.json")

	endpoint := os.Getenv("AIM_TELEMETRY_ENDPOINT")
	if endpoint == "" && !isRunningInTest() {
		endpoint = DefaultPostHogEndpoint
	}

	apiKey := os.Getenv("AIM_TELEMETRY_API_KEY")
	if apiKey == "" && !isRunningInTest() {
		apiKey = DefaultPostHogAPIKey
	}

	var machineID string
	if enabled {
		machineID = AnonymousMachineID(baseDir)
	}

	return &DefaultClient{
		baseDir:    baseDir,
		cacheDir:   cacheDir,
		version:    version,
		distinctID: machineID,
		enabled:    enabled,
		spooler:    NewSpooler(spoolPath, endpoint, apiKey, nil),
	}
}

// Track records an anonymous event to the local spool.
func (c *DefaultClient) Track(event string, properties map[string]any) {
	if !c.enabled {
		return
	}
	ev := NewEvent(event, c.distinctID, c.version, runtime.GOOS, runtime.GOARCH, properties)
	_ = c.spooler.Append(ev)
}

// Flush dispatches spooled events over HTTPS.
func (c *DefaultClient) Flush(ctx context.Context) error {
	if !c.enabled {
		return nil
	}
	return c.spooler.Flush(ctx)
}

// Close attempts a bounded 1500ms flush on CLI shutdown.
func (c *DefaultClient) Close() error {
	if !c.enabled {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	return c.Flush(ctx)
}
