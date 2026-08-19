package iconlogger

import (
	"sync"
	"testing"

	"github.com/briandowns/spinner"
)

// TestPauseResumeRaceOnSpinnerField exercises PauseSpinner/ResumeSpinner
// concurrently with the same lock+reassign pattern StartSpinner/StopSpinner
// use on il.spinner. isSupported() gates the real entry points behind a TTY
// check that's false in test environments, so this drives the identical
// synchronization directly to prove the race independent of that gate.
func TestPauseResumeRaceOnSpinnerField(t *testing.T) {
	logger := &IconLogger{
		mutex:   sync.Mutex{},
		spinner: &spinner.Spinner{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.mutex.Lock()
			logger.spinner = &spinner.Spinner{}
			logger.mutex.Unlock()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.PauseSpinner()
			logger.ResumeSpinner()
		}()
	}
	wg.Wait()
}
