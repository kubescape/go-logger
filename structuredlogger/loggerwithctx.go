package structuredlogger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kubescape/go-logger/helpers"
)

var _ helpers.ILogger = (*StructuredLoggerWithCtx)(nil)

type StructuredLoggerWithCtx struct {
	slogL *slog.Logger
	level *slog.LevelVar
	ctx   context.Context
}

func (sl *StructuredLoggerWithCtx) GetLevel() string {
	return levelToString(sl.level.Level())
}

func (sl *StructuredLoggerWithCtx) SetWriter(w *os.File) {
	// slog writes to the handler's writer, which is configured at creation time
}

func (sl *StructuredLoggerWithCtx) GetWriter() *os.File {
	return nil
}

func (sl *StructuredLoggerWithCtx) Ctx(ctx context.Context) helpers.ILogger {
	return &StructuredLoggerWithCtx{
		slogL: sl.slogL,
		level: sl.level,
		ctx:   ctx,
	}
}

func (sl *StructuredLoggerWithCtx) LoggerName() string {
	return LoggerName
}

func (sl *StructuredLoggerWithCtx) SetLevel(level string) error {
	l := stringToLevel(level)
	sl.level.Set(l)
	return nil
}

func (sl *StructuredLoggerWithCtx) Fatal(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
	os.Exit(1)
}

func (sl *StructuredLoggerWithCtx) Error(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) Warning(msg string, details ...helpers.IDetails) {
	sl.slogL.WarnContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) Success(msg string, details ...helpers.IDetails) {
	// Success is logged as Info with a "success" attribute
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), attrs...)
}

func (sl *StructuredLoggerWithCtx) Info(msg string, details ...helpers.IDetails) {
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) Debug(msg string, details ...helpers.IDetails) {
	sl.slogL.DebugContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) Start(msg string, details ...helpers.IDetails) {
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) StopSuccess(msg string, details ...helpers.IDetails) {
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), attrs...)
}

func (sl *StructuredLoggerWithCtx) StopError(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *StructuredLoggerWithCtx) TimedWrapper(funcName string, timeout time.Duration, task func()) {
	helpers.TimedWrapperHelper(sl, funcName, timeout, task)
}
