package logger

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/go-logger/iconlogger"
	"github.com/kubescape/go-logger/nonelogger"
	"github.com/kubescape/go-logger/prettylogger"
	"github.com/kubescape/go-logger/structuredlogger"
	"github.com/kubescape/go-logger/zaplogger"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// Logger level environment name
	EnvLoggerLevel = "KS_LOGGER_LEVEL"
	// Logger name environment name
	EnvLoggerName = "KS_LOGGER_NAME"
)

var (
	// loggerPtr stores *helpers.ILogger; accessed atomically for lock-free reads.
	loggerPtr atomic.Pointer[helpers.ILogger]
	// initMu serializes initialization in InitLogger and the lazy path of L().
	initMu sync.Mutex
)

// L returns the initialized logger. If no logger has been set, it lazily initializes the default.
func L() helpers.ILogger {
	if p := loggerPtr.Load(); p != nil {
		return *p
	}
	initMu.Lock()
	defer initMu.Unlock()
	if p := loggerPtr.Load(); p != nil {
		return *p
	}
	initLoggerInternal("")
	return *loggerPtr.Load()
}

/*
	InitLogger initialize desired logger

Use:
InitLogger("<logger name>")

Supported logger names (call ListLoggersNames() for listing supported loggers)
- "pretty", "colorful": Human friendly colorful logger
- "slog": Structured logger from package "log/slog" with OpenTelemetry support
- "zap": Logger from package "go.uber.org/zap"
- "none", "mock", "empty", "ignore": Logger will not print anything
- "icon", "emoji": Human friendly logger with colors and icons/symbols

Default:
- "pretty"

If the logger name is empty, will try to get the logger name from the environment variable KS_LOGGER_NAME.
If the logger level environment variable is set, will set the logger level to the value of the environment variable.

InitLogger is not safe for concurrent use. It should be called once during
program startup, before any concurrent access to L().

e.g.
InitLogger("none") -> will initialize the mock logger
*/
func InitLogger(loggerName string) {
	initMu.Lock()
	defer initMu.Unlock()
	initLoggerInternal(loggerName)
}

func initLoggerInternal(loggerName string) {
	if loggerName == "" {
		loggerName = os.Getenv(EnvLoggerName)
	}

	var newLogger helpers.ILogger
	switch strings.ToLower(loggerName) {
	case structuredlogger.LoggerName:
		newLogger = structuredlogger.NewStructuredLogger()
	case zaplogger.LoggerName:
		newLogger = zaplogger.NewZapLogger()
	case prettylogger.LoggerName, "colorful":
		newLogger = prettylogger.NewPrettyLogger()
	case iconlogger.LoggerName, "emoji":
		newLogger = iconlogger.NewIconLogger()
	case nonelogger.LoggerName, "mock", "empty", "ignore":
		newLogger = nonelogger.NewNoneLogger()
	default:
		newLogger = prettylogger.NewPrettyLogger()
	}

	if lev := os.Getenv(EnvLoggerLevel); lev != "" {
		if err := newLogger.SetLevel(lev); err != nil {
			newLogger.Warning("failed to set logger level", helpers.String("environment", EnvLoggerLevel), helpers.Error(err))
		}
	}

	// Caller must hold initMu.
	iface := helpers.ILogger(newLogger)
	loggerPtr.Store(&iface)
}

func InitDefaultLogger() {
	InitLogger("")
}

func DisableColor(flag bool) {
	prettylogger.DisableColor(flag)
}

func EnableColor(flag bool) {
	prettylogger.EnableColor(flag)
}

func ListLoggersNames() []string {
	return []string{prettylogger.LoggerName, structuredlogger.LoggerName, iconlogger.LoggerName, zaplogger.LoggerName, nonelogger.LoggerName}
}

// InitOtel configures OpenTelemetry to export data to OTEL_COLLECTOR_SVC using uptrace collector.
// You have to set the env variable OTEL_COLLECTOR_SVC to enable otel.
// It is required to call ShutdownOtel on the context at the end of the main.
//
//	func main() {
//	  // configure otel
//	  ctx := logger.InitOtel(logger.L(), "<service>", "<version>")
//	  defer logger.ShutdownOtel(ctx)
//
//	  // create a span
//	  ctx, span := otel.Tracer("").Start(ctx, "<name of the span>")
//	  defer span.End()
//
//	  if err := cmd.Execute(ctx); err != nil {
//	      // attach log to the span
//	      logger.L().Ctx(ctx).Fatal(err.Error())
//	  }
//	}
func InitOtel(serviceName, version, accountId, clusterName string, collectorUrl url.URL) context.Context {
	ctx := context.Background()
	if collectorUrl.Scheme == "" {
		collectorUrl.Scheme = "http"
	}
	if collectorUrl.User == nil {
		collectorUrl.User = url.User("t")
	}
	if collectorUrl.Path == "" {
		collectorUrl.Path = "1"
	}
	attrs := []attribute.KeyValue{
		attribute.String("account.id", accountId),
		attribute.String("cluster.name", clusterName),
	}

	uptrace.ConfigureOpentelemetry(
		uptrace.WithDSN(collectorUrl.String()),
		uptrace.WithServiceName(serviceName),
		uptrace.WithServiceVersion(version),
		uptrace.WithResourceAttributes(attrs...),
	)

	return ctx
}

func ShutdownOtel(ctx context.Context) {
	uptrace.Shutdown(ctx)
}
