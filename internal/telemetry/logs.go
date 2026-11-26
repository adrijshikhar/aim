package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/aim-cli/aim/internal/config"
	"github.com/aim-cli/aim/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otelLog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var (
	posthogLogMu       sync.RWMutex
	posthogLogProvider *sdklog.LoggerProvider
	otelErrHandlerOnce sync.Once
)

func nowUTC() time.Time {
	return time.Now().UTC()
}

func hasPostHogLogProvider() bool {
	posthogLogMu.RLock()
	defer posthogLogMu.RUnlock()
	return posthogLogProvider != nil
}

// InitPostHogLogs creates a dedicated OpenTelemetry logger for purpose-written
// AIM lifecycle logs. It intentionally does not register a global provider or
// attach to AIM's existing logger, so no pre-existing log lines are exported.
func InitPostHogLogs(cfg *config.Config) {
	if !IsTelemetryEnabled(cfg) {
		return
	}

	projectToken, host := resolvePostHogCredentials()
	if projectToken == "" || host == "" {
		return
	}

	posthogLogMu.Lock()
	defer posthogLogMu.Unlock()
	if posthogLogProvider != nil {
		return
	}

	exporter, err := otlploghttp.New(context.Background(),
		otlploghttp.WithEndpointURL(host+"/i/v1/logs"),
		otlploghttp.WithHeaders(map[string]string{
			"Authorization": "Bearer " + projectToken,
		}),
	)
	if err != nil {
		logger.Debug("failed to create PostHog log exporter: %v", err)
		return
	}

	otelErrHandlerOnce.Do(func() {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
			logger.Debug("otel background error: %v", err)
		}))
	})

	posthogLogProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)
}

// LogPostHogInfo emits an integration-owned log record. Callers must supply
// only non-PII, non-sensitive attributes because this record is exported.
func LogPostHogInfo(ctx context.Context, message string, attributes map[string]any) {
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

	posthogLogMu.RLock()
	defer posthogLogMu.RUnlock()
	if posthogLogProvider == nil {
		return
	}
	posthogLogProvider.Logger("aim.posthog").Emit(ctx, record)
}

// ClosePostHogLogs flushes integration-owned log records at process shutdown.
func ClosePostHogLogs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	return ClosePostHogLogsWithContext(ctx)
}

// ClosePostHogLogsWithContext flushes integration-owned log records using ctx.
func ClosePostHogLogsWithContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	posthogLogMu.Lock()
	defer posthogLogMu.Unlock()
	if posthogLogProvider == nil {
		return nil
	}
	err := posthogLogProvider.Shutdown(ctx)
	posthogLogProvider = nil
	return err
}
