// Package otelsetup initialises OpenTelemetry providers (Tracer + Logger +
// Meter) for a kubescape service. It owns endpoint resolution, per-signal
// exporter gating, credential header injection, TLS/plaintext detection, and
// the in-memory ring-buffer log processor used for retroactive log export.
//
// Auth header policy: X-API-Key and X-Customer-GUID are injected whenever
// cfg.AccessKey is non-empty, regardless of the endpoint hostname. This is the
// same credential-presence gate used by the SBOM scan-failure reporter and
// avoids fragile hostname matching that breaks on domain changes or
// self-hosted collector deployments.
//
// Ordering constraint: callers MUST invoke InitProviders before any code that
// captures global.GetLoggerProvider() at construction time (e.g. the
// kubescape/go-logger structuredlogger). Once InitProviders returns the global
// TracerProvider, LoggerProvider, and MeterProvider are set.
package otelsetup

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ProviderConfig carries the inputs InitProviders needs to construct OTEL
// providers and authenticate with the back-office when credentials are present.
type ProviderConfig struct {
	ServiceName    string
	ServiceVersion string
	NodeName       string
	PodName        string
	Namespace      string
	ClusterName    string
	AccountID      string // e.g. clusterData.AccountID
	AccessKey      string // e.g. from /etc/credentials
}

// InitProviders initialises the TracerProvider, LoggerProvider, and
// MeterProvider. It returns a combined shutdown func that flushes batches with
// a 5s timeout. When OTEL_EXPORTER_OTLP_ENDPOINT is unset, providers fall back
// to no-op (no panics, no log noise). When cfg.AccessKey is non-empty,
// X-API-Key and X-Customer-GUID headers are attached to every outbound RPC.
func InitProviders(ctx context.Context, cfg ProviderConfig) (shutdown func(context.Context) error, err error) {
	applyLegacyEnvAliases()

	baseEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	traceEndpoint := coalesce(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"), baseEndpoint)
	logEndpoint := coalesce(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"), baseEndpoint)
	metricEndpoint := coalesce(os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"), baseEndpoint)

	if traceEndpoint == "" && logEndpoint == "" && metricEndpoint == "" {
		slog.Debug("otelsetup: no endpoint configured, telemetry disabled")
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.TraceContext{})
		return func(context.Context) error { return nil }, nil
	}

	authHeaders := buildAuthHeaders(cfg.AccessKey, cfg.AccountID)

	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.K8SClusterName(cfg.ClusterName),
		semconv.K8SNodeName(cfg.NodeName),
		semconv.K8SPodName(cfg.PodName),
		semconv.K8SNamespaceName(cfg.Namespace),
	))
	if err != nil {
		return nil, err
	}

	// --- TracerProvider ---
	var tp *sdktrace.TracerProvider
	if traceEndpoint != "" {
		spanExporter, err := otlptracegrpc.New(ctx, grpcTraceOpts(traceEndpoint, authHeaders)...)
		if err != nil {
			return nil, err
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(spanExporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
	} else {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
	}
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// --- LoggerProvider ---
	ringBuf := &RingBufferLogProcessor{}
	var logProvider *sdklog.LoggerProvider
	if logEndpoint != "" {
		logExporter, err := otlploggrpc.New(ctx, grpcLogOpts(logEndpoint, authHeaders)...)
		if err != nil {
			if tp != nil {
				_ = tp.Shutdown(ctx)
			}
			return nil, err
		}
		logProvider = sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
			sdklog.WithProcessor(ringBuf),
		)
		global.SetLoggerProvider(logProvider)
	}

	// --- MeterProvider ---
	var mp *sdkmetric.MeterProvider
	if metricEndpoint != "" {
		metricExporter, err := otlpmetricgrpc.New(ctx, grpcMetricOpts(metricEndpoint, authHeaders)...)
		if err != nil {
			if tp != nil {
				_ = tp.Shutdown(ctx)
			}
			if logProvider != nil {
				_ = logProvider.Shutdown(ctx)
			}
			return nil, err
		}
		mp = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
			sdkmetric.WithResource(res),
		)
		otel.SetMeterProvider(mp)
	}

	// --- Debug HTTP listener (gated by ENABLE_DEBUG_LISTENER=true) ---
	var debugSrv *http.Server
	if os.Getenv("ENABLE_DEBUG_LISTENER") == "true" && logProvider != nil {
		port := coalesce(os.Getenv("OTEL_DEBUG_PORT"), "6062")
		l := logProvider.Logger(cfg.ServiceName + "/ringbuf")
		mux := http.NewServeMux()
		mux.HandleFunc("POST /debug/flush-ring-buffer", func(w http.ResponseWriter, r *http.Request) {
			ringBuf.FlushToBackend(r.Context(), l)
			w.WriteHeader(http.StatusNoContent)
		})
		debugSrv = &http.Server{
			Addr:              "localhost:" + port,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		go func() {
			if err := debugSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Warn("otelsetup: debug listener stopped", "error", err.Error())
			}
		}()
	}

	shutdown = func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		var tpErr, logErr, mpErr, debugErr error
		if tp != nil {
			tpErr = tp.Shutdown(shutdownCtx)
		}
		if logProvider != nil {
			logErr = logProvider.Shutdown(shutdownCtx)
		}
		if mp != nil {
			mpErr = mp.Shutdown(shutdownCtx)
		}
		if debugSrv != nil {
			debugErr = debugSrv.Shutdown(shutdownCtx)
		}
		return errors.Join(tpErr, logErr, mpErr, debugErr)
	}
	return shutdown, nil
}

func grpcTraceOpts(endpoint string, headers map[string]string) []otlptracegrpc.Option {
	var opts []otlptracegrpc.Option
	if strings.Contains(endpoint, "://") {
		opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
	}
	if !strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	return opts
}

func grpcLogOpts(endpoint string, headers map[string]string) []otlploggrpc.Option {
	var opts []otlploggrpc.Option
	if strings.Contains(endpoint, "://") {
		opts = append(opts, otlploggrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlploggrpc.WithEndpoint(endpoint))
	}
	if !strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(headers))
	}
	return opts
}

func grpcMetricOpts(endpoint string, headers map[string]string) []otlpmetricgrpc.Option {
	var opts []otlpmetricgrpc.Option
	if strings.Contains(endpoint, "://") {
		opts = append(opts, otlpmetricgrpc.WithEndpointURL(endpoint))
	} else {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}
	if !strings.HasPrefix(endpoint, "https://") {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}
	return opts
}

// applyLegacyEnvAliases maps the older OTEL_COLLECTOR_SVC env var onto the
// standard OTEL_EXPORTER_OTLP_ENDPOINT so existing deployments keep working.
// The legacy var wins only when the standard one is unset.
func applyLegacyEnvAliases() {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		if legacy := os.Getenv("OTEL_COLLECTOR_SVC"); legacy != "" {
			_ = os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", legacy)
		}
	}
}

// buildAuthHeaders returns credential headers when accessKey is non-empty.
// Headers are injected for any endpoint, consistent with the credential-presence
// gate used by the rest of the codebase (e.g. SBOM scan-failure reporter).
func buildAuthHeaders(accessKey, accountID string) map[string]string {
	if accessKey == "" {
		return nil
	}
	return map[string]string{
		"X-API-Key":       accessKey,
		"X-Customer-GUID": accountID,
	}
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
