package iconlogger

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/kubescape/go-logger/helpers"
	"github.com/stretchr/testify/assert"
)

func TestIconLoggerPrint(t *testing.T) {
	tests := []struct {
		name        string
		loggerLevel helpers.Level
		printLevel  helpers.Level
		msg         string
		details     []helpers.IDetails
		expected    string
		shouldPrint bool
	}{
		{
			"Print Info",
			helpers.InfoLevel,
			helpers.InfoLevel,
			"Info Message",
			[]helpers.IDetails{
				helpers.String("key1", "value1"),
				helpers.Int("key2", 123),
			},
			"〜 Info Message\n",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &IconLogger{
				level: tt.loggerLevel,
			}

			// Use a WaitGroup to wait for all goroutines to finish
			var wg sync.WaitGroup

			// Start multiple goroutines to call the print method concurrently
			for i := 0; i < 1000; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					logger.print(tt.printLevel, tt.msg, tt.details...)
				}()
			}

			// Wait for all goroutines to finish
			wg.Wait()

		})
	}
}

type mockDetail struct {
	key   string
	value interface{}
}

func (m *mockDetail) Key() string {
	return m.key
}

func (m *mockDetail) Value() interface{} {
	return m.value
}

func TestDetailsToString(t *testing.T) {
	tests := []struct {
		name     string
		details  []helpers.IDetails
		expected string
	}{
		{
			"Single Detail",
			[]helpers.IDetails{&mockDetail{"key1", "value1"}},
			"key1: value1",
		},
		{
			"Multiple Details",
			[]helpers.IDetails{
				&mockDetail{"key1", "value1"},
				&mockDetail{"key2", 123},
			},
			"key1: value1; key2: 123",
		},
		{
			"No Details",
			[]helpers.IDetails{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detailsToString(tt.details)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSymbol(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		expect string
	}{
		{"Warning", "warning", " ❗ "},
		{"Success", "success", "✅  "},
		{"Fatal", "fatal", "❌  "},
		{"Error", "error", "❌  "},
		{"Debug", "debug", "🐞  "},
		{"Default", "info", "ℹ️ "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSymbol(tt.level)
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestIconLoggerGetLevel(t *testing.T) {
	logger := &IconLogger{
		level: helpers.InfoLevel,
	}
	assert.Equal(t, "info", logger.GetLevel())
}

func TestIconLoggerSetAndGetWriter(t *testing.T) {
	logger := &IconLogger{}
	writer := os.Stdout
	logger.SetWriter(writer)
	assert.Equal(t, writer, logger.GetWriter())
}

func TestIconLoggerCtx(t *testing.T) {
	logger := &IconLogger{}
	ctx := context.Background()
	assert.Equal(t, logger, logger.Ctx(ctx))
}

func TestIconLoggerLoggerName(t *testing.T) {
	logger := &IconLogger{}
	assert.Equal(t, LoggerName, logger.LoggerName())
}
