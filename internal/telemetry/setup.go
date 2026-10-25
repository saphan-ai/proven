package telemetry

// This file reads the standard OTEL_* variables (ADR-0045 point 2, ADR-0013)
// and builds the OTLP exporters. Proven adds no setting of its own: where
// OpenTelemetry names one, that name is the one used.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/saphan-ai/proven/internal/envdoc"
)

// The standard variables Setup reads. Exporter details it does not read here
// (headers, timeout, certificates, batch delay) are read by the OpenTelemetry
// SDK itself from the process environment.
const (
	EnvDisabled        = "OTEL_SDK_DISABLED"
	EnvEndpoint        = "OTEL_EXPORTER_OTLP_ENDPOINT"
	EnvTracesEndpoint  = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	EnvMetricsEndpoint = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"
	EnvProtocol        = "OTEL_EXPORTER_OTLP_PROTOCOL"
	EnvTracesProtocol  = "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"
	EnvMetricsProtocol = "OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"
	EnvMetricInterval  = "OTEL_METRIC_EXPORT_INTERVAL"
	EnvServiceName     = "OTEL_SERVICE_NAME"
)

// DefaultServiceName names Proven in a backend when OTEL_SERVICE_NAME is unset.
const DefaultServiceName = "proven"

// protocolHTTP is the one OTLP transport built in. gRPC is not.
const protocolHTTP = "http/protobuf"

// ConfigError names a refused OTEL_* value and what is allowed. The reviewer
// turns it into its own *ConfigError, so the shell exits 7 before any review.
type ConfigError struct {
	Field   string
	Value   string
	Reason  string
	Allowed []string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("telemetry: %s=%q: %s", e.Field, e.Value, e.Reason)
}

// Environment lists the OTEL_* variables Proven reads, for docs/CONFIGURATION.md. Each is the
// OpenTelemetry specification's own; Proven adds no setting of its own (ADR-0045 point 2).
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{Name: EnvEndpoint, Default: "none (telemetry is off)", Description: "the base URL of an OTLP/HTTP collector, for example `http://127.0.0.1:4318`. Setting it turns telemetry on; the trace goes to `<url>/v1/traces` and the metrics to `<url>/v1/metrics`. Sample mode reads this variable too"},
		{Name: EnvTracesEndpoint, Default: "none", Description: "the full URL the trace is sent to. It wins over `" + EnvEndpoint + "` for the trace and, set alone, turns the trace on"},
		{Name: EnvMetricsEndpoint, Default: "none", Description: "the full URL the metrics are sent to. It wins over `" + EnvEndpoint + "` for the metrics and, set alone, turns the metrics on"},
		{Name: EnvDisabled, Default: "none", Description: "`true` turns telemetry off even when an endpoint is set"},
		{Name: EnvProtocol, Default: envdoc.Tick(protocolHTTP), Description: "the OTLP transport. Only `" + protocolHTTP + "` is built in; any other value is refused by name. `" + EnvTracesProtocol + "` and `" + EnvMetricsProtocol + "` are read the same way"},
		{Name: EnvServiceName, Default: envdoc.Tick(DefaultServiceName), Description: "the service name a backend shows"},
		{Name: EnvMetricInterval, Default: "`60000`", Description: "how often the metrics are pushed, in milliseconds. The lab sets it low so a dashboard moves while a review runs"},
	}
}

// Setup returns the telemetry the environment asks for. It returns nil, which
// is off, when no endpoint is set or OTEL_SDK_DISABLED is true: nothing is
// built and no connection is opened. env nil reads nothing, so it is off.
func Setup(ctx context.Context, env func(string) string) (*T, error) {
	if env == nil {
		return nil, nil
	}
	if strings.EqualFold(strings.TrimSpace(env(EnvDisabled)), "true") {
		return nil, nil
	}
	traceURL, err := signalURL(env, EnvTracesEndpoint, "/v1/traces")
	if err != nil {
		return nil, err
	}
	metricURL, err := signalURL(env, EnvMetricsEndpoint, "/v1/metrics")
	if err != nil {
		return nil, err
	}
	if traceURL == "" && metricURL == "" {
		return nil, nil
	}
	for _, name := range []string{EnvProtocol, EnvTracesProtocol, EnvMetricsProtocol} {
		if v := strings.TrimSpace(env(name)); v != "" && v != protocolHTTP {
			return nil, &ConfigError{
				Field: name, Value: v, Allowed: []string{protocolHTTP},
				Reason: "only the HTTP transport is built in",
			}
		}
	}
	interval, err := metricInterval(env)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(env(EnvServiceName))
	if name == "" {
		name = DefaultServiceName
	}
	// One process is one review, so the instance id tells two reviews apart in a backend: a counter
	// that starts at zero in each process must not be read as one series that went back to zero.
	// The name and the id come first, so OTEL_RESOURCE_ATTRIBUTES, read after them, can still say more.
	var id [8]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, fmt.Errorf("telemetry: instance id: %w", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", name),
			attribute.String("service.instance.id", hex.EncodeToString(id[:])),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}
	var tp *sdktrace.TracerProvider
	if traceURL != "" {
		exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(traceURL))
		if err != nil {
			return nil, fmt.Errorf("telemetry: trace exporter: %w", err)
		}
		tp = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp), sdktrace.WithResource(res))
	} else {
		tp = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	}
	var mp *sdkmetric.MeterProvider
	if metricURL != "" {
		exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(metricURL))
		if err != nil {
			return nil, fmt.Errorf("telemetry: metric exporter: %w", err)
		}
		mp = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(interval))),
			sdkmetric.WithResource(res))
	} else {
		mp = sdkmetric.NewMeterProvider(sdkmetric.WithResource(res))
	}
	return New(tp, mp)
}

// signalURL returns the URL one signal is sent to, "" when none is named. The
// signal-specific variable is a full URL and wins. The general one is a base
// URL and gets the signal's path.
func signalURL(env func(string) string, specific, path string) (string, error) {
	if v := strings.TrimSpace(env(specific)); v != "" {
		return checkURL(specific, v)
	}
	v := strings.TrimSpace(env(EnvEndpoint))
	if v == "" {
		return "", nil
	}
	u, err := checkURL(EnvEndpoint, v)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(u, "/") + path, nil
}

func checkURL(field, v string) (string, error) {
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", &ConfigError{
			Field: field, Value: v, Allowed: []string{"an http:// or https:// URL"},
			Reason: "not a URL the OTLP/HTTP exporter can use",
		}
	}
	return v, nil
}

// metricInterval reads OTEL_METRIC_EXPORT_INTERVAL: whole milliseconds, as the
// OpenTelemetry specification has it. Unset is the specification's 60 seconds.
func metricInterval(env func(string) string) (time.Duration, error) {
	v := strings.TrimSpace(env(EnvMetricInterval))
	if v == "" {
		return 60 * time.Second, nil
	}
	ms, err := strconv.Atoi(v)
	if err != nil || ms <= 0 {
		return 0, &ConfigError{
			Field: EnvMetricInterval, Value: v, Allowed: []string{"a whole number of milliseconds above zero"},
			Reason: "not a number of milliseconds",
		}
	}
	return time.Duration(ms) * time.Millisecond, nil
}
