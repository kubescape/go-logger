package otelsetup

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type receivedExport struct {
	request *collectorlog.ExportLogsServiceRequest
	headers metadata.MD
}

type recordingLogService struct {
	collectorlog.UnimplementedLogsServiceServer
	exports chan receivedExport
}

func (s *recordingLogService) Export(ctx context.Context, request *collectorlog.ExportLogsServiceRequest) (*collectorlog.ExportLogsServiceResponse, error) {
	headers, _ := metadata.FromIncomingContext(ctx)
	s.exports <- receivedExport{request: request, headers: headers.Copy()}
	return &collectorlog.ExportLogsServiceResponse{}, nil
}

func TestInitProviders_LogBatchExport(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	service := &recordingLogService{exports: make(chan receivedExport, 8)}
	collectorlog.RegisterLogsServiceServer(server, service)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	previousLogs := global.GetLoggerProvider()
	previousTraces := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		global.SetLoggerProvider(previousLogs)
		otel.SetTracerProvider(previousTraces)
		otel.SetTextMapPropagator(previousPropagator)
	})
	for _, key := range []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "OTEL_BLRP_MAX_EXPORT_BATCH_SIZE", "OTEL_BLRP_MAX_QUEUE_SIZE"} {
		t.Setenv(key, "")
	}
	t.Setenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", listener.Addr().String())
	t.Setenv("OTEL_BLRP_SCHEDULE_DELAY", "3600000")
	t.Setenv("KS_LOGGER_LEVEL", "info")
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	shutdown, err := InitProviders(ctx, ProviderConfig{
		ServiceName: "export-test", AccessKey: "test-key", AccountID: "test-account",
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = shutdown(context.Background()) })
	provider, ok := global.GetLoggerProvider().(*sdklog.LoggerProvider)
	require.True(t, ok)
	logger := provider.Logger("export-test")
	traceID, spanID := trace.TraceID{1}, trace.SpanID{2}
	emitCtx := trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled,
	}))
	emit := func(i int) {
		var record otellog.Record
		record.SetBody(attribute.StringValue(fmt.Sprintf("record-%d", i)))
		record.SetSeverity(otellog.SeverityInfo)
		record.AddAttributes(attribute.Int("index", i), attribute.StringSlice("labels", []string{"first", "second"}))
		logger.Emit(emitCtx, record)
	}
	// One full default-sized batch and one partial batch must survive ForceFlush.
	for i := range 513 {
		emit(i)
	}
	require.NoError(t, provider.ForceFlush(ctx))
	// Shutdown must also deliver the final partial batch with the long timer.
	emit(513)
	require.NoError(t, shutdown(ctx))
	require.Len(t, service.exports, 3)
	nextIndex := 0
	for _, size := range []int{512, 1, 1} {
		exported := <-service.exports
		assert.Equal(t, []string{"test-key"}, exported.headers.Get("x-api-key"))
		assert.Equal(t, []string{"test-account"}, exported.headers.Get("x-api-account"))
		require.Len(t, exported.request.ResourceLogs, 1)
		resource := exported.request.ResourceLogs[0]
		var serviceName string
		for _, kv := range resource.Resource.Attributes {
			if kv.Key == "service.name" {
				serviceName = kv.Value.GetStringValue()
			}
		}
		assert.Equal(t, "export-test", serviceName)
		require.Len(t, resource.ScopeLogs, 1)
		scope := resource.ScopeLogs[0]
		assert.Equal(t, "export-test", scope.Scope.Name)
		require.Len(t, scope.LogRecords, size)
		for _, record := range scope.LogRecords {
			assert.Equal(t, fmt.Sprintf("record-%d", nextIndex), record.Body.GetStringValue())
			assert.Equal(t, traceID[:], record.TraceId)
			assert.Equal(t, spanID[:], record.SpanId)
			require.Len(t, record.Attributes, 2)
			assert.Equal(t, "index", record.Attributes[0].Key)
			assert.Equal(t, int64(nextIndex), record.Attributes[0].Value.GetIntValue())
			assert.Equal(t, "labels", record.Attributes[1].Key)
			values := record.Attributes[1].Value.GetArrayValue().GetValues()
			require.Len(t, values, 2)
			assert.Equal(t, "first", values[0].GetStringValue())
			assert.Equal(t, "second", values[1].GetStringValue())
			nextIndex++
		}
	}
	assert.Equal(t, 514, nextIndex)
}

func TestSDKRecordToLogRecord_PreservesContents(t *testing.T) {
	var original otellog.Record
	timestamp := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	original.SetTimestamp(timestamp)
	original.SetObservedTimestamp(timestamp.Add(time.Second))
	original.SetSeverity(otellog.SeverityWarn)
	original.SetSeverityText("WARN")
	original.SetEventName("snapshot")
	original.SetBody(attribute.StringValue("original body"))
	attributes := []attribute.KeyValue{
		attribute.String("operation", "test"),
		attribute.Int64("attempt", 7),
		attribute.StringSlice("labels", []string{"first", "second"}),
	}
	original.AddAttributes(attributes...)
	buffer := &RingBufferLogProcessor{}
	provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(buffer))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	provider.Logger("conversion-test").Emit(t.Context(), original)
	require.Equal(t, 1, buffer.size)
	converted := sdkRecordToLogRecord(&buffer.buf[0])
	assert.Equal(t, original.Timestamp(), converted.Timestamp())
	assert.Equal(t, original.ObservedTimestamp(), converted.ObservedTimestamp())
	assert.Equal(t, original.Severity(), converted.Severity())
	assert.Equal(t, original.SeverityText(), converted.SeverityText())
	assert.Equal(t, original.EventName(), converted.EventName())
	assert.Equal(t, "original body", converted.Body().AsString())
	var got []attribute.KeyValue
	converted.WalkAttributes(func(kv attribute.KeyValue) bool {
		got = append(got, kv)
		return true
	})
	assert.Equal(t, attributes, got)
}
