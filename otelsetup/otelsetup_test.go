package otelsetup

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// --- buildAuthHeaders (AC8 / AC9) ---

func TestBuildAuthHeaders_WithCredentials(t *testing.T) {
	h := buildAuthHeaders("my-key", "my-guid")
	assert.Equal(t, "my-key", h["X-API-KEY"], "X-API-KEY must be set when accessKey is non-empty")
	assert.Equal(t, "my-guid", h["X-API-ACCOUNT"], "X-API-ACCOUNT must be set when accessKey is non-empty")
}

func TestBuildAuthHeaders_NoCredentials_ReturnsNil(t *testing.T) {
	assert.Nil(t, buildAuthHeaders("", "my-guid"), "no headers when accessKey is empty")
}

// TestGrpcTraceOpts_HeadersInjectedWhenCredsPresent verifies the option slice
// includes WithHeaders when credentials are supplied (AC8).
func TestGrpcTraceOpts_HeadersInjectedWhenCredsPresent(t *testing.T) {
	opts := grpcTraceOpts("collector:4317", buildAuthHeaders("key", "guid"))
	// WithEndpoint + WithInsecure + WithHeaders = 3
	assert.Len(t, opts, 3, "must include WithHeaders when credentials present")
}

// TestGrpcTraceOpts_NoHeadersWhenNoCreds verifies no WithHeaders option is
// added when no credentials are configured (AC9).
func TestGrpcTraceOpts_NoHeadersWhenNoCreds(t *testing.T) {
	opts := grpcTraceOpts("collector:4317", nil)
	// WithEndpoint + WithInsecure = 2
	assert.Len(t, opts, 2, "must not include WithHeaders when no credentials")
}

func TestGrpcLogOpts_HeadersInjectedWhenCredsPresent(t *testing.T) {
	opts := grpcLogOpts("collector:4317", buildAuthHeaders("key", "guid"))
	assert.Len(t, opts, 3)
}

func TestGrpcMetricOpts_HeadersInjectedWhenCredsPresent(t *testing.T) {
	opts := grpcMetricOpts("collector:4317", buildAuthHeaders("key", "guid"))
	assert.Len(t, opts, 3)
}

// TestGrpcTraceOpts_HTTPSEndpoint verifies https:// uses WithEndpointURL and
// skips WithInsecure.
func TestGrpcTraceOpts_HTTPSEndpoint(t *testing.T) {
	opts := grpcTraceOpts("https://collector:4317", nil)
	// WithEndpointURL only (WithInsecure skipped for https) = 1
	assert.Len(t, opts, 1)
}

// --- InitProviders no-op path ---

func TestInitProviders_NoEndpoint_ReturnsNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")

	shutdown, err := InitProviders(context.Background(), ProviderConfig{ServiceName: "test"})
	assert.NoError(t, err)
	assert.NoError(t, shutdown(context.Background()))
}

// --- RingBufferLogProcessor ---

func TestRingBufferLogProcessor_WrapsCorrectly(t *testing.T) {
	p := &RingBufferLogProcessor{}
	ctx := context.Background()
	r := new(sdklog.Record)
	for range len(p.buf) + 100 {
		_ = p.OnEmit(ctx, r)
	}
	p.mu.Lock()
	sz := p.size
	p.mu.Unlock()
	assert.Equal(t, len(p.buf), sz, "buffer size must be capped at ring capacity")
}

func TestRingBufferLogProcessor_FlushClearsBuffer(t *testing.T) {
	p := &RingBufferLogProcessor{}
	ctx := context.Background()
	r := new(sdklog.Record)
	for range 10 {
		_ = p.OnEmit(ctx, r)
	}

	counter := &recordCounter{}
	provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(counter))
	l := provider.Logger("test")

	p.FlushToBackend(ctx, l)
	assert.Equal(t, int32(10), counter.n.Load(), "flush must re-emit all buffered records")

	p.mu.Lock()
	sz := p.size
	p.mu.Unlock()
	assert.Equal(t, 0, sz, "buffer must be empty after flush")

	// Second flush must emit nothing
	p.FlushToBackend(ctx, l)
	assert.Equal(t, int32(10), counter.n.Load(), "second flush must not re-export already-flushed records")
}

// recordCounter counts OnEmit calls via a real sdklog.Processor so we can
// use provider.Logger() — otellog.Logger uses the embedded interface pattern
// and cannot be implemented externally.
type recordCounter struct {
	n atomic.Int32
}

func (c *recordCounter) OnEmit(_ context.Context, _ *sdklog.Record) error           { c.n.Add(1); return nil }
func (c *recordCounter) Enabled(_ context.Context, _ sdklog.EnabledParameters) bool { return true }
func (c *recordCounter) Shutdown(_ context.Context) error                            { return nil }
func (c *recordCounter) ForceFlush(_ context.Context) error                          { return nil }
