package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/aim-cli/aim/internal/config"
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

// NewClient constructs a new telemetry client with privacy and opt-out checks applied.
func NewClient(baseDir, cacheDir, version string, cfg *config.Config) *DefaultClient {
	enabled := IsTelemetryEnabled(cfg)
	spoolPath := filepath.Join(cacheDir, "telemetry_spool.json")
	endpoint := os.Getenv("AIM_TELEMETRY_ENDPOINT")

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
		spooler:    NewSpooler(spoolPath, endpoint, nil),
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

// Close attempts a bounded 200ms flush on CLI shutdown.
func (c *DefaultClient) Close() error {
	if !c.enabled {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	return c.Flush(ctx)
}
