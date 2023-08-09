package prettylogger

import (
	"os"
	"sync"
	"testing"

	"github.com/briandowns/spinner"
	"github.com/stretchr/testify/assert"
)

func TestPrettyLoggerStartSpinner(t *testing.T) {
	logger := &PrettyLogger{
		mutex:   sync.Mutex{},
		spinner: &spinner.Spinner{},
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start multiple goroutines to call StartSpinner concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.StartSpinner(os.Stdout, "Testing spinner")
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	assert.True(t, (isSupported() && logger.spinner.Active()) || (!isSupported() && !logger.spinner.Active()))
}

func TestPrettyLoggerStopSpinner(t *testing.T) {
	logger := &PrettyLogger{
		mutex:   sync.Mutex{},
		spinner: &spinner.Spinner{},
	}

	// Start the spinner first
	logger.StartSpinner(os.Stdout, "Starting spinner for test")

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start multiple goroutines to call StopSpinner concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.StopSpinner("Stopping spinner")
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	assert.False(t, isSupported() && !logger.spinner.Active())

}

func TestPrettyLoggerPauseAndResumeSpinner(t *testing.T) {
	logger := &PrettyLogger{
		mutex:   sync.Mutex{},
		spinner: &spinner.Spinner{},
	}

	// Start the spinner first
	logger.StartSpinner(os.Stdout, "Starting spinner for test")

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.PauseSpinner()
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	assert.False(t, isSupported() && !logger.spinner.Active())

	// Reset the WaitGroup for the next set of goroutines
	wg = sync.WaitGroup{}

	// Start multiple goroutines to call ResumeSpinner concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.ResumeSpinner()
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Assuming spinner has an Active() method, we can assert that it's active after all the goroutines have resumed it
	assert.True(t, (isSupported() && logger.spinner.Active()) || (!isSupported() && !logger.spinner.Active()))
}
