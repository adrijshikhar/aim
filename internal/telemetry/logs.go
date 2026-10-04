package telemetry

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otelLog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var posthogLogProvider *sdklog.LoggerProvider

// InitPostHogLogs creates a dedicated OpenTelemetry logger for purpose-written
// AIM lifecycle logs. It intentionally does not register a global provider or
// attach to AIM's existing logger, so no pre-existing log lines are exported.
func InitPostHogLogs(cfg *config.Config) {
	if posthogLogProvider != nil || !IsTelemetryEnabled(cfg) {
		return
	}

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

	if projectToken == "" || host == "" {
		return
	}

	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(host+"/i/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + projectToken,
		}),
	)
	if err != nil {
		log.Printf("failed to create PostHog log exporter: %v", err)
		return
	}

	posthogLogProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
}

// LogPostHogInfo emits an integration-owned log record. Callers must supply
// only non-PII, non-sensitive attributes because this record is exported.
func LogPostHogInfo(ctx context.Context, message string, attributes map[string]any) {
	if posthogLogProvider == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	record := otelLog.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(otelLog.SeverityInfo)
	record.SetSeverityText("INFO")
	record.SetBody(attribute.StringValue(message))
	for key, value := range attributes {
		switch value := value.(type) {
		case string:
			record.AddAttributes(attribute.String(key, value))
		case int:
			record.AddAttributes(attribute.Int(key, value))
		case bool:
			record.AddAttributes(attribute.Bool(key, value))
		}
	}

	posthogLogProvider.Logger("aim.posthog").Emit(ctx, record)
}

// ClosePostHogLogs flushes integration-owned log records at process shutdown.
func ClosePostHogLogs() error {
	if posthogLogProvider == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	err := posthogLogProvider.Shutdown(ctx)
	posthogLogProvider = nil
	return err
}
