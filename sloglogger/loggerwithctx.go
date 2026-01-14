package sloglogger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kubescape/go-logger/helpers"
)

var _ helpers.ILogger = (*SlogLoggerWithCtx)(nil)

type SlogLoggerWithCtx struct {
	slogL *slog.Logger
	level *slog.LevelVar
	ctx   context.Context
}

func (sl *SlogLoggerWithCtx) GetLevel() string {
	return levelToString(sl.level.Level())
}

func (sl *SlogLoggerWithCtx) SetWriter(w *os.File) {
	// slog writes to the handler's writer, which is configured at creation time
}

func (sl *SlogLoggerWithCtx) GetWriter() *os.File {
	return nil
}

func (sl *SlogLoggerWithCtx) Ctx(ctx context.Context) helpers.ILogger {
	return &SlogLoggerWithCtx{
		slogL: sl.slogL,
		level: sl.level,
		ctx:   ctx,
	}
}

func (sl *SlogLoggerWithCtx) LoggerName() string {
	return LoggerName
}

func (sl *SlogLoggerWithCtx) SetLevel(level string) error {
	l := stringToLevel(level)
	sl.level.Set(l)
	return nil
}

func (sl *SlogLoggerWithCtx) Fatal(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
	os.Exit(1)
}

func (sl *SlogLoggerWithCtx) Error(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) Warning(msg string, details ...helpers.IDetails) {
	sl.slogL.WarnContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) Success(msg string, details ...helpers.IDetails) {
	// Success is logged as Info with a "success" attribute
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), attrs...)
}

func (sl *SlogLoggerWithCtx) Info(msg string, details ...helpers.IDetails) {
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) Debug(msg string, details ...helpers.IDetails) {
	sl.slogL.DebugContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) Start(msg string, details ...helpers.IDetails) {
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) StopSuccess(msg string, details ...helpers.IDetails) {
	attrs := append([]any{slog.Bool("success", true)}, detailsToAttrs(details)...)
	sl.slogL.InfoContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), attrs...)
}

func (sl *SlogLoggerWithCtx) StopError(msg string, details ...helpers.IDetails) {
	sl.slogL.ErrorContext(sl.ctx, strings.ToValidUTF8(msg, helpers.InvalidUtf8ReplacementString), detailsToAttrs(details)...)
}

func (sl *SlogLoggerWithCtx) TimedWrapper(funcName string, timeout time.Duration, task func()) {
	helpers.TimedWrapperHelper(sl, funcName, timeout, task)
}
