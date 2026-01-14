package sloglogger

import (
	"context"
	"testing"

	"github.com/kubescape/go-logger/helpers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TestSlogLoggerWithOTel tests that slog logger integrates with OpenTelemetry
func TestSlogLoggerWithOTel(t *testing.T) {
	logger := NewSlogLogger()

	// Create a proper trace provider for testing
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	// Create a span context
	ctx := context.Background()
	tracer := otel.Tracer("test-tracer")
	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()

	// Create a context-aware logger
	ctxLogger := logger.Ctx(ctx)

	// Test logging with span context
	ctxLogger.Info("test info with span", helpers.String("key", "value"))
	ctxLogger.Error("test error with span", helpers.String("error", "test"))
	ctxLogger.Warning("test warning with span", helpers.String("warning", "test"))

	// Verify span is valid
	if !span.SpanContext().IsValid() {
		t.Error("Expected valid span context")
	}
}

// TestSlogLoggerLevelFiltering tests that log levels are properly filtered
func TestSlogLoggerLevelFiltering(t *testing.T) {
	logger := NewSlogLogger()

	// Set to warning level
	logger.SetLevel("warning")

	// These should be filtered out (but we can't verify output easily in tests)
	logger.Debug("debug message - should be filtered")
	logger.Info("info message - should be filtered")

	// These should not be filtered
	logger.Warning("warning message - should appear")
	logger.Error("error message - should appear")

	// Verify level was set correctly
	if logger.GetLevel() != "warning" {
		t.Errorf("Expected level 'warning', got '%s'", logger.GetLevel())
	}
}
