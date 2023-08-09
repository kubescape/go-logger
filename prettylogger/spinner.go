package prettylogger

import (
	"os"
	"time"

	spinnerpkg "github.com/briandowns/spinner"
	"github.com/mattn/go-isatty"
)

func (pl *PrettyLogger) StartSpinner(w *os.File, message string) {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	if pl.spinner != nil && pl.spinner.Active() {
		return
	}
	if isatty.IsTerminal(os.Stdout.Fd()) {
		pl.spinner = spinnerpkg.New(spinnerpkg.CharSets[14], 100*time.Millisecond, spinnerpkg.WithWriterFile(w)) // Build our new spinner
		pl.spinner.Suffix = " " + message
		pl.spinner.Start()
	}
}

func (pl *PrettyLogger) StopSpinner(message string) {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()

	if pl.spinner == nil || !pl.spinner.Active() {
		return
	}
	pl.spinner.FinalMSG = message
	pl.spinner.Stop()
	pl.spinner = nil
}

func (pl *PrettyLogger) PauseSpinner() {
	if pl.spinner == nil || !pl.spinner.Active() {
		return
	}
	pl.spinner.Stop()
}

func (pl *PrettyLogger) ResumeSpinner() {
	if pl.spinner == nil || pl.spinner.Active() {
		return
	}
	pl.spinner.Start()
}
