package logger

import (
	"os"
	"testing"

	"github.com/kubescape/go-logger/nonelogger"
	"github.com/kubescape/go-logger/prettylogger"
	"github.com/kubescape/go-logger/zaplogger"
)

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
			os.Setenv(EnvLoggerName, tt.envs.loggerName)
			os.Setenv(EnvLoggerLevel, tt.envs.loggerLevel)

			InitLogger(tt.args.loggerName)

			if l.GetLevel() != tt.want.loggerLevel {
				t.Errorf("GetLevel() = %v, want %v", l.GetLevel(), tt.want.loggerLevel)
			}
			if l.LoggerName() != tt.want.loggerName {
				t.Errorf("LoggerName() = %v, want %v", l.LoggerName(), tt.want.loggerName)
			}
		})
	}
}
