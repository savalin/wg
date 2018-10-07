package wg

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// waitGroup enhanced wait group struct
type waitGroup struct {
	waitGroupStatus
	ctx      context.Context
	receiver chan WaitgroupFunc
	sender   chan WaitgroupFunc

	errors      []error
	stopOnError bool

	stackBuffer []WaitgroupFunc
	capacity    uint32
	length      int
	timeout     *time.Duration
}

type waitGroupStatus struct {
	status     int
	statusLock sync.RWMutex
}

// WithContext make wait group work with context timeout and Done
func (wg *waitGroup) WithContext(ctx context.Context) WaitGroup {
	wg.ctx = ctx
	return wg
}

// Add adds new task in waitgroup
func (wg *waitGroup) Add(f ...WaitgroupFunc) WaitGroup {
	wg.stackBuffer = append(wg.stackBuffer, f...)
	return wg
}

// SetTimeout defines timeout for all tasks
func (wg *waitGroup) SetTimeout(t time.Duration) WaitGroup {
	wg.timeout = &t
	return wg
}

// SetStopOnError make wait group stops if any task returns error
func (wg *waitGroup) SetStopOnError(b bool) WaitGroup {
	wg.stopOnError = b
	return wg
}

// SetCapacity defines tasks channel capacity
func (wg *waitGroup) SetCapacity(c int) WaitGroup {
	if c >= 0 {
		wg.capacity = uint32(c)
	}
	return wg
}

// Start runs tasks in separate goroutines
func (wg *waitGroup) Start() WaitGroup {
	if wg.checkStatus(statusSuccess) {
		return wg
	}

	wg.init()

	if wg.length > 0 {
		var (
			failed = make(chan error, wg.length)
			done   = make(chan struct{}, wg.length)
			wgDone = make(chan struct{})

			cancel    context.CancelFunc
			startTime = time.Now()
			timeout   = defaultMaxTimeout
		)

		if wg.timeout != nil && *wg.timeout != 0 {
			startTime = time.Now()
			timeout = *wg.timeout
		}

		wg.ctx, cancel = context.WithTimeout(wg.ctx, timeout)
		defer cancel()

		go func() {
			for f := range wg.sender {
				select {
				case wg.receiver <- f:
					// successfully sent a func to the execution queue
				case <-wgDone:
					return
				}
			}
		}()

	ForLoop:
		for wg.length > 0 {
			select {

			// If we have functions in queue to be ran
			case f := <-wg.receiver:
				go func(f WaitgroupFunc, failed chan<- error, done chan<- struct{}) {
					if wg.stopOnError {
						wg.do(f, failed, done, true)
						return
					}
					wg.do(f, failed, done, false)

				}(f, failed, done)

				// If we got en error returned from some goroutine
			case err := <-failed:
				wg.errors = append(wg.errors, err)
				wg.length--
				wg.setStatus(statusError)
				if wg.stopOnError {
					break ForLoop
				}

				// If all working goroutines are successfully finished
			case <-done:
				wg.length--

				// If context deadline exceeded
			case <-wg.ctx.Done():
				if wg.ctx.Err().Error() == context.Canceled.Error() {
					wg.setStatus(statusCancelled)
				} else if deadlineTime, ok := wg.ctx.Deadline(); ok {
					wg.errors = append(wg.errors, ErrorTimeout(deadlineTime.Sub(startTime)))
					wg.setStatus(statusTimeout)
				}
				break ForLoop
			}
		}

		close(wgDone)
		close(wg.sender)
	}

	return wg
}

// GetCapacity defines tasks channel capacity
func (wg *waitGroup) GetCapacity() int {
	return int(wg.capacity)
}

// GetLastError returns last error that caught by execution process
func (wg *waitGroup) GetLastError() error {
	if l := len(wg.errors); l > 0 {
		return wg.errors[l-1]
	}
	return nil
}

// GetAllErrors returns all errors that caught by execution process
func (wg *waitGroup) GetAllErrors() []error {
	return wg.errors
}

// Reset performs cleanup task queue and reset state
func (wg *waitGroup) Reset() WaitGroup {
	wg.stackBuffer = []WaitgroupFunc{}
	wg.receiver = nil
	wg.sender = nil
	wg.timeout = nil
	wg.stopOnError = false
	wg.setStatus(statusIdle)
	wg.errors = []error{}
	wg.ctx = nil

	return wg
}

func (wg *waitGroup) init() {
	wg.setStatus(statusSuccess)

	if wg.ctx == nil {
		wg.ctx = context.Background()
	}

	wg.length = len(wg.stackBuffer)
	cap := wg.length
	if c := wg.GetCapacity(); c > 0 {
		cap = c
	}

	wg.receiver = make(chan WaitgroupFunc, cap)
	wg.sender = make(chan WaitgroupFunc, wg.length)
	for _, f := range wg.stackBuffer {
		wg.sender <- f
	}
}

func (wg *waitGroup) do(f WaitgroupFunc, failed chan<- error, done chan<- struct{}, stopOnError bool) {
	// Handle panic and pack it into stdlib error
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, stackBufferSize)
			count := runtime.Stack(buf, false)
			failed <- fmt.Errorf("Panic handeled\n%v\n%s", r, buf[:count])
		}
	}()

	// Check stop on error
	if stopOnError && !wg.checkStatus(statusSuccess) {
		// If some other goroutine get an error
		done <- struct{}{}
		return
	}

	if err := f(wg.ctx); err != nil {
		failed <- err
		return
	}

	done <- struct{}{}
}

func (wg *waitGroup) setStatus(status int) {
	if status < statusIdle || status > statusError {
		return
	}

	wg.statusLock.Lock()
	wg.status = status
	wg.statusLock.Unlock()
}

func (wg *waitGroup) checkStatus(status int) bool {
	if status < statusIdle || status > statusError {
		return false
	}

	wg.statusLock.RLock()
	defer wg.statusLock.RUnlock()

	return wg.status == status
}
