package prettylogger

import (
	"context"
	"fmt"
	"os"
	"sync"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/kubescape/go-logger/helpers"
)

const LoggerName string = "pretty"

type PrettyLogger struct {
	writer  *os.File
	level   helpers.Level
	spinner *spinnerpkg.Spinner
	mutex   sync.Mutex
}

var _ helpers.ILogger = (*PrettyLogger)(nil) // ensure all interface methods are here

func NewPrettyLogger() *PrettyLogger {

	return &PrettyLogger{
		writer:  os.Stderr, // default to stderr
		level:   helpers.InfoLevel,
		spinner: nil,
		mutex:   sync.Mutex{},
	}
}

func (pl *PrettyLogger) GetLevel() string                      { return pl.level.String() }
func (pl *PrettyLogger) SetWriter(w *os.File)                  { pl.writer = w }
func (pl *PrettyLogger) GetWriter() *os.File                   { return pl.writer }
func (pl *PrettyLogger) Ctx(_ context.Context) helpers.ILogger { return pl }
func (pl *PrettyLogger) LoggerName() string                    { return LoggerName }

func (pl *PrettyLogger) SetLevel(level string) error {
	pl.level = helpers.ToLevel(level)
	if pl.level == helpers.UnknownLevel {
		return fmt.Errorf("level '%s' unknown", level)
	}
	return nil
}
func (pl *PrettyLogger) Fatal(msg string, details ...helpers.IDetails) {
	pl.print(helpers.FatalLevel, msg, details...)
	os.Exit(1)
}
func (pl *PrettyLogger) Error(msg string, details ...helpers.IDetails) {
	pl.print(helpers.ErrorLevel, msg, details...)
}
func (pl *PrettyLogger) Warning(msg string, details ...helpers.IDetails) {
	pl.print(helpers.WarningLevel, msg, details...)
}
func (pl *PrettyLogger) Info(msg string, details ...helpers.IDetails) {
	pl.print(helpers.InfoLevel, msg, details...)
}
func (pl *PrettyLogger) Debug(msg string, details ...helpers.IDetails) {
	pl.print(helpers.DebugLevel, msg, details...)
}
func (pl *PrettyLogger) Success(msg string, details ...helpers.IDetails) {
	pl.print(helpers.SuccessLevel, msg, details...)
}
func (pl *PrettyLogger) Start(msg string, details ...helpers.IDetails) {
	pl.StartSpinner(pl.writer, msg)
}
func (pl *PrettyLogger) StopSuccess(msg string, details ...helpers.IDetails) {
	pl.StopSpinner(getSymbol("success") + msg + "\n")
}
func (pl *PrettyLogger) StopError(msg string, details ...helpers.IDetails) {
	pl.StopSpinner(getSymbol("error") + msg + "\n")
}

func (pl *PrettyLogger) print(level helpers.Level, msg string, details ...helpers.IDetails) {
	pl.PauseSpinner()
	if !level.Skip(pl.level) {
		pl.mutex.Lock()
		prefix(level)(pl.writer, "%s", getSymbol(level.String()))
		if d := detailsToString(details); d != "" {
			msg = fmt.Sprintf("%s. %s", msg, d)
		}
		message(pl.writer, fmt.Sprintf("%s\n", msg))
		pl.mutex.Unlock()
	}
	pl.ResumeSpinner()
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
		return "❗ "
	case "success":
		return "✔️  "
	case "fatal", "error":
		return "❌  "
	case "debug":
		return "—  "
	default:
		return "〜 "
	}
}
