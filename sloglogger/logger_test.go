package sloglogger

import (
	"context"
	"testing"

	"github.com/kubescape/go-logger/helpers"
)

func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger()
	if logger == nil {
		t.Fatal("NewSlogLogger returned nil")
	}
	if logger.LoggerName() != LoggerName {
		t.Errorf("LoggerName() = %v, want %v", logger.LoggerName(), LoggerName)
	}
}

func TestSlogLoggerSetLevel(t *testing.T) {
	logger := NewSlogLogger()

	tests := []struct {
		level string
		want  string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warning", "warning"},
		{"error", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			err := logger.SetLevel(tt.level)
			if err != nil {
				t.Errorf("SetLevel(%s) returned error: %v", tt.level, err)
			}
			if logger.GetLevel() != tt.want {
				t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), tt.want)
			}
		})
	}
}

func TestSlogLoggerLogging(t *testing.T) {
	logger := NewSlogLogger()

	// Test all log methods to ensure they don't panic
	logger.Debug("debug message", helpers.String("key", "value"))
	logger.Info("info message", helpers.Int("count", 42))
	logger.Warning("warning message", helpers.String("warning", "test"))
	logger.Error("error message", helpers.Error(nil))
	logger.Success("success message")
	logger.Start("start message")
	logger.StopSuccess("stop success message")
	logger.StopError("stop error message")
}

func TestSlogLoggerWithCtx(t *testing.T) {
	logger := NewSlogLogger()
	ctx := context.Background()
	ctxLogger := logger.Ctx(ctx)

	if ctxLogger == nil {
		t.Fatal("Ctx() returned nil")
	}

	if ctxLogger.LoggerName() != LoggerName {
		t.Errorf("LoggerName() = %v, want %v", ctxLogger.LoggerName(), LoggerName)
	}

	// Test logging with context
	ctxLogger.Info("info with context", helpers.String("context", "test"))
	ctxLogger.Error("error with context", helpers.String("context", "test"))
}
