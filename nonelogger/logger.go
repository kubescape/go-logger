package nonelogger

import (
	"context"
	"os"

	"github.com/kubescape/go-logger/helpers"
)

const LoggerName string = "none"

type NoneLogger struct {
}

var _ helpers.ILogger = (*NoneLogger)(nil) // ensure all interface methods are here

func NewNoneLogger() *NoneLogger {
	return &NoneLogger{}
}

func (nl *NoneLogger) GetLevel() string                                    { return "" }
func (nl *NoneLogger) Ctx(_ context.Context) helpers.ILogger               { return nl }
func (nl *NoneLogger) LoggerName() string                                  { return LoggerName }
func (nl *NoneLogger) SetWriter(w *os.File)                                {}
func (nl *NoneLogger) GetWriter() *os.File                                 { return nil }
func (nl *NoneLogger) SetLevel(level string) error                         { return nil }
func (nl *NoneLogger) Fatal(msg string, details ...helpers.IDetails)       {}
func (nl *NoneLogger) Error(msg string, details ...helpers.IDetails)       {}
func (nl *NoneLogger) Warning(msg string, details ...helpers.IDetails)     {}
func (nl *NoneLogger) Success(msg string, details ...helpers.IDetails)     {}
func (nl *NoneLogger) Info(msg string, details ...helpers.IDetails)        {}
func (nl *NoneLogger) Debug(msg string, details ...helpers.IDetails)       {}
func (nl *NoneLogger) Start(msg string, details ...helpers.IDetails)       {}
func (nl *NoneLogger) StopSuccess(msg string, details ...helpers.IDetails) {}
func (nl *NoneLogger) StopError(msg string, details ...helpers.IDetails)   {}
