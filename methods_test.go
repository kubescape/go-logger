package logger

import (
	"os"
	"sync"
	"testing"

	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/go-logger/nonelogger"
	"github.com/kubescape/go-logger/prettylogger"
	"github.com/kubescape/go-logger/structuredlogger"
	"github.com/kubescape/go-logger/zaplogger"
)

// resetLogger clears the global logger so the next L() or InitLogger call reinitializes it.
func resetLogger() {
	loggerPtr.Store(nil)
}

func TestInitLogger(t *testing.T) {
	type args struct {
		loggerName  string
		loggerLevel string
	}
	type envs struct {
		loggerName  string
		loggerLevel string
	}

	tests := []struct {
		name string
		want args
		args args
		envs envs
	}{
		{
			name: "TestInitLogger default",
			want: args{
				loggerName:  prettylogger.LoggerName,
				loggerLevel: "info",
			},
		},
		{
			name: "TestInitLogger slog info",
			want: args{
				loggerName:  structuredlogger.LoggerName,
				loggerLevel: "info",
			},
			args: args{
				loggerName: "slog",
			},
			envs: envs{},
		},
		{
			name: "TestInitLogger slog debug",
			want: args{
				loggerName:  structuredlogger.LoggerName,
				loggerLevel: "debug",
			},
			args: args{
				loggerName: "slog",
			},
			envs: envs{
				loggerLevel: "debug",
			},
		},
		{
			name: "TestInitLogger zap info",
			want: args{
				loggerName:  zaplogger.LoggerName,
				loggerLevel: "info",
			},
			args: args{
				loggerName: "zap",
			},
			envs: envs{},
		},
		{
			name: "TestInitLogger zap debug",
			want: args{
				loggerName:  zaplogger.LoggerName,
				loggerLevel: "debug",
			},
			args: args{
				loggerName: "zap",
			},
			envs: envs{
				loggerLevel: "debug",
			},
		},
		{
			name: "TestInitLogger zap debug",
			want: args{
				loggerName:  zaplogger.LoggerName,
				loggerLevel: "debug",
			},
			args: args{},
			envs: envs{
				loggerLevel: "debug",
				loggerName:  "zap",
			},
		},
		{
			name: "TestInitLogger",
			want: args{
				loggerName:  prettylogger.LoggerName,
				loggerLevel: "debug",
			},
			args: args{
				loggerName: "pretty",
			},
			envs: envs{
				loggerLevel: "debug",
			},
		},
		{
			name: "TestInitLogger colorful warning",
			want: args{
				loggerName:  prettylogger.LoggerName,
				loggerLevel: "warning",
			},
			args: args{
				loggerName: "colorful",
			},
			envs: envs{
				loggerLevel: "warning",
			},
		},
		{
			name: "TestInitLogger none",
			want: args{
				loggerName:  nonelogger.LoggerName,
				loggerLevel: "",
			},
			args: args{
				loggerName: "none",
			},
			envs: envs{
				loggerLevel: "error",
			},
		},
		{
			name: "TestInitLogger mock",
			want: args{
				loggerName:  nonelogger.LoggerName,
				loggerLevel: "",
			},
			args: args{},
			envs: envs{
				loggerName: "mock",
			},
		},
		{
			name: "TestInitLogger empty",
			want: args{
				loggerName: nonelogger.LoggerName,
			},
			args: args{},
			envs: envs{
				loggerName: "empty",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetLogger()

			os.Setenv(EnvLoggerName, tt.envs.loggerName)
			os.Setenv(EnvLoggerLevel, tt.envs.loggerLevel)

			InitLogger(tt.args.loggerName)

			logger := L()
			if logger.GetLevel() != tt.want.loggerLevel {
				t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), tt.want.loggerLevel)
			}
			if logger.LoggerName() != tt.want.loggerName {
				t.Errorf("LoggerName() = %v, want %v", logger.LoggerName(), tt.want.loggerName)
			}
		})
	}
}

func TestL_ConcurrentAccess(t *testing.T) {
	resetLogger()

	var wg sync.WaitGroup

	// Half the goroutines call InitLogger, half call L().
	// This exercises the InitLogger + L() interleaving that
	// previously caused a deadlock due to lock ordering inversion.
	const n = 100
	loggers := make([]helpers.ILogger, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				InitLogger("pretty")
			}
			loggers[idx] = L()
		}(i)
	}
	wg.Wait()

	for i, logger := range loggers {
		if logger == nil {
			t.Errorf("goroutine %d: L() returned nil", i)
		}
	}
}
