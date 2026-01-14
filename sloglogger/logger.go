package sloglogger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/kubescape/go-logger/helpers"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
)

const LoggerName string = "slog"

type SlogLogger struct {
	slogL *slog.Logger
	level *slog.LevelVar
}

var _ helpers.ILogger = (*SlogLogger)(nil) // ensure all interface methods are here

func NewSlogLogger() *SlogLogger {
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)

	// Create handler with OpenTelemetry bridge
	loggerProvider := global.GetLoggerProvider()
	otelHandler := otelslog.NewHandler(LoggerName, otelslog.WithLoggerProvider(loggerProvider))

	// Wrap with LevelHandler to support level filtering
	handler := &LevelHandler{
		Handler: otelHandler,
		level:   level,
	}

	logger := slog.New(handler)

	return &SlogLogger{
		slogL: logger,
		level: level,
	}
}

func (sl *SlogLogger) GetLevel() string {
	return levelToString(sl.level.Level())
}

func (sl *SlogLogger) SetWriter(w *os.File) {
	// slog writes to the handler's writer, which is configured at creation time
	// For simplicity, we'll skip dynamic writer changes for now
}

func (sl *SlogLogger) GetWriter() *os.File {
	return nil
}

func (sl *SlogLogger) Ctx(ctx context.Context) helpers.ILogger {
	return &SlogLoggerWithCtx{
		slogL: sl.slogL,
		level: sl.level,
		ctx:   ctx,
	}
}

func (sl *SlogLogger) LoggerName() string {
	return LoggerName
}

func (sl *SlogLogger) SetLevel(level string) error {
	l := stringToLevel(level)
	sl.level.Set(l)
	return nil
}

func (sl *SlogLogger) Fatal(msg string, details ...helpers.IDetails) {
	sl.slogL.Error(msg, detailsToAttrs(details)...)
	os.Exit(1)
}

func (sl *SlogLogger) Error(msg string, details ...helpers.IDetails) {
	sl.slogL.Error(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) Warning(msg string, details ...helpers.IDetails) {
	sl.slogL.Warn(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) Success(msg string, details ...helpers.IDetails) {
	// Success is logged as Info with a "success" attribute
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.Info(msg, attrs...)
}

func (sl *SlogLogger) Info(msg string, details ...helpers.IDetails) {
	sl.slogL.Info(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) Debug(msg string, details ...helpers.IDetails) {
	sl.slogL.Debug(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) Start(msg string, details ...helpers.IDetails) {
	sl.slogL.Info(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) StopSuccess(msg string, details ...helpers.IDetails) {
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.Info(msg, attrs...)
}

func (sl *SlogLogger) StopError(msg string, details ...helpers.IDetails) {
	sl.slogL.Error(msg, detailsToAttrs(details)...)
}

func (sl *SlogLogger) TimedWrapper(funcName string, timeout time.Duration, task func()) {
	helpers.TimedWrapperHelper(sl, funcName, timeout, task)
}

// detailsToAttrs converts helpers.IDetails to slog attributes
func detailsToAttrs(details []helpers.IDetails) []any {
	attrs := make([]any, 0, len(details))
	for _, d := range details {
		attrs = append(attrs, slog.Any(d.Key(), d.Value()))
	}
	return attrs
}

// stringToLevel converts string level to slog.Level
func stringToLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "success":
		return slog.LevelInfo
	case "warning", "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// levelToString converts slog.Level to string
func levelToString(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warning"
	case slog.LevelError:
		return "error"
	default:
		return "info"
	}
}

// LevelHandler wraps a handler to add level filtering
type LevelHandler struct {
	slog.Handler
	level *slog.LevelVar
}

func (h *LevelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *LevelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.Handler.Handle(ctx, r)
}

func (h *LevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelHandler{
		Handler: h.Handler.WithAttrs(attrs),
		level:   h.level,
	}
}

func (h *LevelHandler) WithGroup(name string) slog.Handler {
	return &LevelHandler{
		Handler: h.Handler.WithGroup(name),
		level:   h.level,
	}
}
