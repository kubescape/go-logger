package iconlogger

import (
	"os"
	"time"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/mattn/go-isatty"
)

func (il *IconLogger) StartSpinner(w *os.File, message string) {
	il.mutex.Lock()
	defer il.mutex.Unlock()

	if il.spinner != nil && il.spinner.Active() {
		return
	}
	if isSupported() {
		il.spinner = spinnerpkg.New(spinnerpkg.CharSets[14], 100*time.Millisecond, spinnerpkg.WithWriterFile(w)) // Build our new spinner
		il.spinner.Suffix = " " + message
		il.spinner.Start()
	}
}

func (il *IconLogger) StopSpinner(message string) {
	il.mutex.Lock()
	defer il.mutex.Unlock()

	if il.spinner == nil || !il.spinner.Active() {
		return
	}
	il.spinner.FinalMSG = message
	il.spinner.Stop()
	il.spinner = nil
}

func (il *IconLogger) PauseSpinner() {
	if il.spinner == nil || !il.spinner.Active() {
		return
	}

	il.spinner.Stop()
}

func (il *IconLogger) ResumeSpinner() {
	if il.spinner == nil || il.spinner.Active() {
		return
	}
	if !isSupported() {
		return
	}
	il.spinner.Start()
}

func isSupported() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
