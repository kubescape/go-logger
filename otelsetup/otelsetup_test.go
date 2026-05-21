package otelsetup

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// --- isARMOEndpoint ---

func TestIsARMOEndpoint_ExactMatch(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "")
	assert.True(t, isARMOEndpoint("otel.armosec.io:4317"))
}

func TestIsARMOEndpoint_SubdomainNotMatched(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "")
	assert.False(t, isARMOEndpoint("evil.otel.armosec.io:4317"))
}

func TestIsARMOEndpoint_CustomerCollector(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "")
	assert.False(t, isARMOEndpoint("customer-collector:4317"))
}

func TestIsARMOEndpoint_EmptyEndpoint(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "")
	assert.False(t, isARMOEndpoint(""))
}

func TestIsARMOEndpoint_ForceAuthEnvVar(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "true")
	assert.True(t, isARMOEndpoint("customer-collector:4317"))
}

func TestIsARMOEndpoint_WithScheme(t *testing.T) {
	t.Setenv("ARMO_OTEL_AUTH", "")
	assert.True(t, isARMOEndpoint("https://otel.armosec.io:4317"))
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

func TestInitProviders_ARMOWithoutCreds_ReturnsNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel.armosec.io:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("ARMO_OTEL_AUTH", "")

	shutdown, err := InitProviders(context.Background(), ProviderConfig{
		ServiceName: "test",
		AccessKey:   "", // no credentials
	})
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

func (c *recordCounter) OnEmit(_ context.Context, _ *sdklog.Record) error          { c.n.Add(1); return nil }
func (c *recordCounter) Enabled(_ context.Context, _ sdklog.EnabledParameters) bool { return true }
func (c *recordCounter) Shutdown(_ context.Context) error                           { return nil }
func (c *recordCounter) ForceFlush(_ context.Context) error                         { return nil }
