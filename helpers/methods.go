package helpers

import (
	"runtime"
	"time"

	"github.com/google/uuid"
)

type IDetails interface {
	Key() string
	Value() interface{}
}

// ======================================================================================
// ============================== String ================================================
// ======================================================================================

// Key
func (s *StringObj) Key() string {
	return s.key
}

// Value
func (s *StringObj) Value() interface{} {
	return s.value
}

// ======================================================================================
// =============================== Error ================================================
// ======================================================================================

// Key
func (s *ErrorObj) Key() string {
	return s.key
}

// Value
func (s *ErrorObj) Value() interface{} {
	return s.value
}

// ======================================================================================
// ================================= Int ================================================
// ======================================================================================

// Key
func (s *IntObj) Key() string {
	return s.key
}

// Value
func (s *IntObj) Value() interface{} {
	return s.value
}

// ======================================================================================
// =========================== Interface ================================================
// ======================================================================================

// Key
func (s *InterfaceObj) Key() string {
	return s.key
}

// Value
func (s *InterfaceObj) Value() interface{} {
	return s.value
}

func TimedWrapperHelper(logger ILogger, funcName string, timeout time.Duration, task func()) {
	if logger.GetLevel() == DebugLevel.String() {
		done := make(chan struct{})
		go func() {
			task()
			close(done)
		}()
		var id string
		select {
		case <-done:
			return
		case <-time.After(timeout):
			// Print all goroutines
			id = uuid.New().String()
			buf := make([]byte, 1<<16)
			n := runtime.Stack(buf, true)
			logger.Debug("TimedWrapperHelper - function still running after timeout", String("id", id), String("function", funcName), Interface("timeout", timeout), Interface("stack", string(buf[:n])))
		}
		<-done // Wait for the task to finish
		if id != "" {
			logger.Debug("TimedWrapperHelper - function finished after timeout", String("id", id), String("function", funcName), Interface("timeout", timeout))
		}
	} else {
		task()
	}
}
