package helpers

import (
	"context"
	"os"
)

// ILogger interface moved here to prevent import cycles
type ILogger interface {
	Fatal(msg string, details ...IDetails) // print log and exit 1
	Error(msg string, details ...IDetails)
	Success(msg string, details ...IDetails)
	Warning(msg string, details ...IDetails)
	Info(msg string, details ...IDetails)
	Debug(msg string, details ...IDetails)

	SetLevel(level string) error
	GetLevel() string

	SetWriter(w *os.File)
	GetWriter() *os.File

	Ctx(ctx context.Context) ILogger
	LoggerName() string
}
