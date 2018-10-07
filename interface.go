package wg

import (
	"context"
	"fmt"
	"time"
)

// New returns new empty WaitGroup
func New() WaitGroup {
	return &waitGroup{}
}

// WaitGroup represents interface
type WaitGroup interface {
	// WithContext make wait group work with context timeout and Done
	// *must be called before Start()
	WithContext(ctx context.Context) WaitGroup

	// Add adds new task in waitgroup
	// *must be called before Start()
	Add(f ...WaitgroupFunc) WaitGroup

	// SetTimeout defines timeout for all tasks
	// *must be called before Start()
	SetTimeout(timeout time.Duration) WaitGroup

	// SetStopOnError make wait group stops if any task returns error
	// *must be called before Start()
	SetStopOnError(flag bool) WaitGroup

	// Start runs tasks in separate goroutines
	Start() WaitGroup

	// GetLastError returns last error that caught by execution process
	GetLastError() error

	// GetAllErrors returns all errors that caught by execution process
	GetAllErrors() []error

	// Reset performs cleanup task queue and reset state
	Reset() WaitGroup
}

// WaitgroupFunc goroutine func to be added in queue
type WaitgroupFunc func(context.Context) error

// ErrorTimeout error on timeout
type ErrorTimeout time.Duration

// Error interface implementation
func (e ErrorTimeout) Error() string {
	return fmt.Sprintf(errTimeoutMessage, time.Duration(e).String())
}
