package telemetry

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/profile"
	"github.com/posthog/posthog-go"
)

var (
	posthogMu     sync.RWMutex
	posthogClient posthog.Client
)

func resolvePostHogCredentials() (string, string) {
	projectToken := strings.TrimSpace(os.Getenv("POSTHOG_PROJECT_TOKEN"))
	if projectToken == "" {
		projectToken = strings.TrimSpace(os.Getenv("AIM_TELEMETRY_API_KEY"))
	}
	if projectToken == "" && !isRunningInTest() {
		projectToken = DefaultPostHogAPIKey
	}

	host := strings.TrimRight(strings.TrimSpace(os.Getenv("POSTHOG_HOST")), "/")
	if host == "" && !isRunningInTest() {
		host = DefaultPostHogHost
	}
	return projectToken, host
}

// InitPostHog creates the single PostHog client for this CLI process. It is a
// production no-op when telemetry is disabled or configuration is absent so
// analytics never prevents a command from running.
func InitPostHog(cfg *config.Config) {
	if !IsTelemetryEnabled(cfg) {
		return
	}

	projectToken, host := resolvePostHogCredentials()
	if projectToken == "" || host == "" {
		if cfg != nil && cfg.Debug {
			if projectToken == "" {
				log.Println("PostHog project token is not configured; event emission skipped")
			}
			if host == "" {
				log.Println("PostHog host is not configured; event emission skipped")
			}
		}
		return
	}

	posthogMu.Lock()
	defer posthogMu.Unlock()
	if posthogClient != nil {
		return
	}

	client, err := posthog.NewWithConfig(projectToken, posthog.Config{
		Endpoint:        host,
		ShutdownTimeout: 1500 * time.Millisecond,
	})
	if err != nil {
		log.Printf("failed to create PostHog client: %v", err)
		return
	}
	posthogClient = client
}

// PostHogClient returns the process-wide client, or nil when PostHog is not configured.
func PostHogClient() posthog.Client {
	posthogMu.RLock()
	defer posthogMu.RUnlock()
	return posthogClient
}

// AccountDistinctID returns the provider-scoped, stable account identifier for
// PostHog when the provider exposes one, falling back to the anonymous machine ID.
func AccountDistinctID(agentName string, account profile.AccountInfo) string {
	agentName = strings.TrimSpace(agentName)
	if agentName != "" && account.UserID != "" {
		return agentName + ":" + account.UserID
	}
	return AnonymousMachineID(config.BaseDir())
}

// IdentifyAccount identifies a successfully authenticated account once at the
// login boundary. Email and name are person properties, never event properties.
func IdentifyAccount(agentName string, account profile.AccountInfo) {
	agentName = strings.TrimSpace(agentName)
	if agentName == "" || account.UserID == "" {
		return
	}
	distinctID := agentName + ":" + account.UserID

	personProperties := map[string]any{}
	if account.Email != "" {
		personProperties["email"] = account.Email
	}
	if account.Name != "" {
		personProperties["name"] = account.Name
	}

	posthogMu.RLock()
	defer posthogMu.RUnlock()
	if posthogClient == nil {
		return
	}
	_ = posthogClient.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      "user_identified",
		Properties: posthog.NewProperties().Set("$set", personProperties),
	})
}

// CaptureAccountEvent records a non-PII product event for an authenticated
// provider account. It is a no-op when analytics is disabled.
func CaptureAccountEvent(agentName string, account profile.AccountInfo, event string, eventProperties map[string]any) {
	if !IsAllowedEvent(event) {
		return
	}
	distinctID := AccountDistinctID(agentName, account)
	if distinctID == "" {
		return
	}

	properties := posthog.NewProperties()
	for key, value := range eventProperties {
		properties.Set(key, value)
	}

	posthogMu.RLock()
	defer posthogMu.RUnlock()
	if posthogClient == nil {
		return
	}
	_ = posthogClient.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: properties,
	})
}

// CaptureException records a sanitized exception event using synchronized client access.
func CaptureException(exceptionType, message string) {
	machineID := AnonymousMachineID(config.BaseDir())

	posthogMu.RLock()
	defer posthogMu.RUnlock()
	if posthogClient == nil {
		return
	}
	_ = posthogClient.Enqueue(posthog.NewDefaultException(
		// posthog.NewDefaultException uses current timestamp passed cleanly
		nowUTC(),
		machineID,
		exceptionType,
		message,
	))
}

// ClosePostHog flushes and closes the process-wide client during graceful shutdown.
func ClosePostHog() error {
	posthogMu.Lock()
	defer posthogMu.Unlock()
	if posthogClient == nil {
		return nil
	}
	err := posthogClient.Close()
	posthogClient = nil
	return err
}
