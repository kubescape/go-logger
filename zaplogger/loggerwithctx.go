package zaplogger

import (
	"context"
	"os"

	"github.com/kubescape/go-logger/helpers"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ helpers.ILogger = (*ZapLoggerWithCtx)(nil)

type ZapLoggerWithCtx struct {
	zapL *otelzap.LoggerWithCtx
	cfg  zap.Config
}

func (zl *ZapLoggerWithCtx) GetLevel() string                      { return zl.cfg.Level.Level().String() }
func (zl *ZapLoggerWithCtx) SetWriter(w *os.File)                  {}
func (zl *ZapLoggerWithCtx) GetWriter() *os.File                   { return nil }
func (zl *ZapLoggerWithCtx) Ctx(_ context.Context) helpers.ILogger { return zl }
func (zl *ZapLoggerWithCtx) LoggerName() string                    { return LoggerName }
func (zl *ZapLoggerWithCtx) SetLevel(level string) error {
	l := zapcore.Level(1)
	err := l.Set(level)
	if err == nil {
		zl.cfg.Level.SetLevel(l)
	}
	return err
}
func (zl *ZapLoggerWithCtx) Fatal(msg string, details ...helpers.IDetails) {
	zl.zapL.Fatal(msg, detailsToZapFields(details)...)
}

func (zl *ZapLoggerWithCtx) Error(msg string, details ...helpers.IDetails) {
	zl.zapL.Error(msg, detailsToZapFields(details)...)
}

func (zl *ZapLoggerWithCtx) Warning(msg string, details ...helpers.IDetails) {
	zl.zapL.Warn(msg, detailsToZapFields(details)...)
}

func (zl *ZapLoggerWithCtx) Success(msg string, details ...helpers.IDetails) {
	zl.zapL.Info(msg, detailsToZapFields(details)...)
}

func (zl *ZapLoggerWithCtx) Info(msg string, details ...helpers.IDetails) {
	zl.zapL.Info(msg, detailsToZapFields(details)...)
}

func (zl *ZapLoggerWithCtx) Debug(msg string, details ...helpers.IDetails) {
	zl.zapL.Debug(msg, detailsToZapFields(details)...)
}
