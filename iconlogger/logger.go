package iconlogger

import (
	"context"
	"fmt"
	"os"
	"sync"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/kubescape/go-logger/helpers"
)

const LoggerName string = "icon"

type IconLogger struct {
	writer  *os.File
	level   helpers.Level
	spinner *spinnerpkg.Spinner
	mutex   sync.Mutex
}

var _ helpers.ILogger = (*IconLogger)(nil) // ensure all interface methods are here

func NewIconLogger() *IconLogger {

	return &IconLogger{
		writer:  os.Stderr, // default to stderr
		level:   helpers.InfoLevel,
		spinner: nil,
		mutex:   sync.Mutex{},
	}
}

func (il *IconLogger) GetLevel() string                      { return il.level.String() }
func (il *IconLogger) SetWriter(w *os.File)                  { il.writer = w }
func (il *IconLogger) GetWriter() *os.File                   { return il.writer }
func (il *IconLogger) Ctx(_ context.Context) helpers.ILogger { return il }
func (il *IconLogger) LoggerName() string                    { return LoggerName }

func (il *IconLogger) SetLevel(level string) error {
	il.level = helpers.ToLevel(level)
	if il.level == helpers.UnknownLevel {
		return fmt.Errorf("level '%s' unknown", level)
	}
	return nil
}
func (il *IconLogger) Fatal(msg string, details ...helpers.IDetails) {
	il.print(helpers.FatalLevel, msg, details...)
	os.Exit(1)
}
func (il *IconLogger) Error(msg string, details ...helpers.IDetails) {
	il.print(helpers.ErrorLevel, msg, details...)
}
func (il *IconLogger) Warning(msg string, details ...helpers.IDetails) {
	il.print(helpers.WarningLevel, msg, details...)
}
func (il *IconLogger) Info(msg string, details ...helpers.IDetails) {
	il.print(helpers.InfoLevel, msg, details...)
}
func (il *IconLogger) Debug(msg string, details ...helpers.IDetails) {
	il.print(helpers.DebugLevel, msg, details...)
}
func (il *IconLogger) Success(msg string, details ...helpers.IDetails) {
	il.print(helpers.SuccessLevel, msg, details...)
}
func (il *IconLogger) Start(msg string, details ...helpers.IDetails) {
	il.StartSpinner(il.writer, generateMessage(msg, details))
}
func (il *IconLogger) StopSuccess(msg string, details ...helpers.IDetails) {
	il.StopSpinner(getSymbol("success") + generateMessage(msg, details) + "\n")
}
func (il *IconLogger) StopError(msg string, details ...helpers.IDetails) {
	il.StopSpinner(getSymbol("error") + generateMessage(msg, details) + "\n")
}

func (il *IconLogger) print(level helpers.Level, msg string, details ...helpers.IDetails) {
	il.PauseSpinner()
	if !level.Skip(il.level) {
		il.mutex.Lock()
		fmt.Fprintf(il.writer, "%s", getSymbol(level.String()))
		fmt.Fprintf(il.writer, fmt.Sprintf("%s\n", generateMessage(msg, details)))
		il.mutex.Unlock()
	}
	il.ResumeSpinner()
}

func detailsToString(details []helpers.IDetails) string {
	s := ""
	for i := range details {
		s += fmt.Sprintf("%s: %v", details[i].Key(), details[i].Value())
		if i < len(details)-1 {
			s += "; "
		}
	}
	return s
}

func getSymbol(level string) string {
	switch level {
	case "warning":
		return " ❗ "
	case "success":
		return "✅  "
	case "fatal", "error":
		return "❌  "
	case "debug":
		return "🐞  "
	default:
		return "ℹ️ "
	}
}

func generateMessage(msg string, details []helpers.IDetails) string {
	if d := detailsToString(details); d != "" {
		msg = fmt.Sprintf("%s. %s", msg, d)
	}
	return msg
}
