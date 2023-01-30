package logger

import (
	"context"
	"net/url"
	"strings"

	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/go-logger/nonelogger"
	"github.com/kubescape/go-logger/prettylogger"
	"github.com/kubescape/go-logger/zaplogger"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/otel/attribute"
)

var l helpers.ILogger

// Return initialized logger. If logger not initialized, will call InitializeLogger() with the default value
func L() helpers.ILogger {
	if l == nil {
		InitDefaultLogger()
	}
	return l
}

/*
	InitLogger initialize desired logger

Use:
InitLogger("<logger name>")

Supported logger names (call ListLoggersNames() for listing supported loggers)
- "zap": Logger from package "go.uber.org/zap"
- "pretty", "colorful": Human friendly colorful logger
- "none", "mock", "empty", "ignore": Logger will not print anything

Default:
- "pretty"

e.g.
InitLogger("none") -> will initialize the mock logger
*/
func InitLogger(loggerName string) {

	switch strings.ToLower(loggerName) {
	case zaplogger.LoggerName:
		l = zaplogger.NewZapLogger()
	case prettylogger.LoggerName, "colorful":
		l = prettylogger.NewPrettyLogger()
	case nonelogger.LoggerName, "mock", "empty", "ignore":
		l = nonelogger.NewNoneLogger()
	default:
		InitDefaultLogger()
	}
}

func InitDefaultLogger() {
	l = prettylogger.NewPrettyLogger()
}

func DisableColor(flag bool) {
	prettylogger.DisableColor(flag)
}

func EnableColor(flag bool) {
	prettylogger.EnableColor(flag)
}

func ListLoggersNames() []string {
	return []string{prettylogger.LoggerName, zaplogger.LoggerName, nonelogger.LoggerName}
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
func InitOtel(serviceName, version, accountId string, collectorUrl url.URL) context.Context {
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
	attrs := []attribute.KeyValue{attribute.String("account.id", accountId)}
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
