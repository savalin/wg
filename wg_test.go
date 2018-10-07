package wg

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

var count int64

func slowFunc(ctx context.Context) error {
	for i := 0; i < 10000000; i++ {
		count *= int64(i)
	}

	return nil
}

func fastFunc(context.Context) error {
	// do nothing
	return nil
}

func errorFunc(context.Context) error {
	for i := 0; i < 10000; i++ {
		count *= int64(i)
	}

	return errors.New("Test error")
}

func panicFunc(context.Context) error {
	panic("Test expected panic, it's ok ;)")
}

// TestWaitGroup_Timeout test for timeout
func Test_WaitGroup_Timeout(t *testing.T) {
	var wg = New()

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
		slowFunc,
	)
	wg.SetTimeout(time.Nanosecond * 1).Start()

	err := wg.GetLastError()
	if err == nil {
		t.Error("WaitGroup should stops with error by timeout!")
	}

	if _, ok := err.(ErrorTimeout); !ok {
		t.Errorf("Wrong error type. Got %[1]T: %[1]q", err)
	}
}

// Test_WaitGroup_Timeout_Context test for timeout
func Test_WaitGroup_Timeout_Context(t *testing.T) {
	var wg = &waitGroup{}

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
		slowFunc,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	wg.WithContext(ctx).Start()
	if wg.status != statusTimeout {
		t.Error("WaitGroup should stops by timeout!")
	}

	err := wg.GetLastError()
	if err == nil {
		t.Error("WaitGroup should stops with error by timeout!")
	}

	if _, ok := err.(ErrorTimeout); !ok {
		t.Errorf("Wrong error type. Got %[1]T: %[1]q", err)
	}
}

// TestWaitGroup_Error test for error
func Test_WaitGroup_Error(t *testing.T) {
	var wg waitGroup

	wg.Add(
		errorFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
		slowFunc,
	)

	wg.SetStopOnError(true).Start()

	if err := wg.GetLastError(); err == nil {
		t.Error("WaitGroup should stops by error!")
	}

}

// TestWaitGroup_Success test for success case
func Test_WaitGroup_Success(t *testing.T) {
	var wg waitGroup

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
	)

	if errs := wg.SetStopOnError(true).Start().GetAllErrors(); len(errs) != 0 {
		t.Errorf("WaitGroup result should be 'success'! But got errors %v", errs)
	}

	if wg.status != statusSuccess {
		t.Error("WaitGroup result should be 'success'!")
	}
}

// Test_WaitGroup_Success_WithCapacity test for success case with capacity
func Test_WaitGroup_Success_WithCapacity(t *testing.T) {
	var wg waitGroup

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
	)
	wg.SetCapacity(2)

	if errs := wg.SetStopOnError(true).Start().GetAllErrors(); len(errs) != 0 {
		t.Errorf("WaitGroup result should be 'success'! But got errors %v", errs)
	}

	if wg.status != statusSuccess {
		t.Error("WaitGroup result should be 'success'!")
	}
}

// Test_WaitGroup_Cancel_Success test for cancel case
func Test_WaitGroup_Cancel_Success(t *testing.T) {
	var wg waitGroup

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(5 * time.Microsecond)
		cancel()
	}()

	if errs := wg.WithContext(ctx).SetStopOnError(true).Start().GetAllErrors(); len(errs) != 0 {
		t.Errorf("WaitGroup result should be 'success'! But got errors %v", errs)
	}

	if wg.status != statusCancelled {
		t.Error("WaitGroup result should be 'canelled'!")
	}
}

// Test_WaitGroup_CancelWithCapacity_Success test for cancel case with capacity
func Test_WaitGroup_CancelWithCapacity_Success(t *testing.T) {
	var wg waitGroup

	wg.Add(
		fastFunc,
		fastFunc,
		fastFunc,
		fastFunc,
		slowFunc,
		slowFunc,
	)
	wg.SetCapacity(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(5 * time.Microsecond)
		cancel()
	}()

	if errs := wg.WithContext(ctx).SetStopOnError(true).Start().GetAllErrors(); len(errs) != 0 {
		t.Errorf("WaitGroup result should be 'success'! But got errors %v", errs)
	}

	if wg.status != statusCancelled {
		t.Error("WaitGroup result should be 'success'!")
	}
}

// TestWaitGroup_PanicError test for success case
func Test_WaitGroup_PanicError(t *testing.T) {
	var wg waitGroup

	wg.Add(slowFunc)
	wg.Add(panicFunc)

	wg.SetStopOnError(true).Start()

	if wg.status != statusError {
		t.Error("WaitGroup result should be 'error'!")
	}

	t.Log(wg.GetLastError())
}

// TestWaitGroup_PanicSuccess test for success case
func TestWaitGroup_PanicSuccess(t *testing.T) {
	var wg waitGroup

	wg.Add(slowFunc)
	wg.Add(panicFunc)

	err := wg.SetStopOnError(false).Start().GetLastError()

	if err == nil {
		t.Error("Panic should be an error")
	}
}

// Test_WaitGroup_StopOnErrorPanic test for success case
func Test_WaitGroup_StopOnErrorPanic(t *testing.T) {
	var wg waitGroup

	wg.Add(slowFunc)
	wg.Add(panicFunc)

	wg.SetStopOnError(true).
		Start()

	if wg.status != statusError {
		t.Error("WaitGroup result should be 'error'!")
	}
}

// TestWaitGroupStopOnError tests WaitGroup_ with StopOnError set to false
// and with failing task.
func TestWaitGroupStopOnError(t *testing.T) {
	var wg waitGroup
	wg.Add(fastFunc)
	wg.Add(errorFunc)
	wg.Add(fastFunc)
	wg.SetStopOnError(false).
		Start()
}

// Test_WaitGroup__NoLeak tests for goroutines leaks
func Test_WaitGroup__NoLeak(t *testing.T) {
	var wg waitGroup

	wg.Add(errorFunc)

	wg.SetStopOnError(true).
		Start()

	time.Sleep(2 * time.Second)

	numGoroutines := runtime.NumGoroutine()

	var wg2 waitGroup

	wg2.Add(errorFunc)
	wg2.Add(slowFunc)

	wg2.SetStopOnError(true).
		Start()

	time.Sleep(3 * time.Second)

	numGoroutines2 := runtime.NumGoroutine()

	if numGoroutines != numGoroutines2 {
		t.Fatalf("We leaked %d goroutine(s)", numGoroutines2-numGoroutines)
	}
}

// Test_WaitGroup_AddTimeout test for timeout
func Test_WaitGroup_AddTimeout(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, fastFunc, fastFunc, slowFunc, slowFunc, slowFunc)
	wg.SetTimeout(time.Nanosecond * 10).SetStopOnError(true).Start()

	if wg.status != statusTimeout {
		t.Error("WaitGroup should stops by timeout!", wg.status)
	}

	err := wg.GetLastError()
	if err == nil {
		t.Error("WaitGroup should stops with error by timeout!")
	}

	if _, ok := err.(ErrorTimeout); !ok {
		t.Errorf("Wrong error type. Got %[1]T: %[1]q", err)
	}
}

// Test_WaitGroup_AddTimeoutSuccess test for timeout that if too long
func Test_WaitGroup_AddTimeoutSuccess(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, fastFunc, fastFunc)
	wg.SetTimeout(time.Second * 10).SetStopOnError(true).Start()

	if wg.status != statusSuccess {
		t.Error("WaitGroup shouldn`t stops by timeout!", wg.status)
	}

	err := wg.GetLastError()
	if err != nil {
		t.Error("WaitGroup shouldn`t stops with error by timeout!")
	}
}

// Test_WaitGroup_GetLastError test
func Test_WaitGroup_GetLastErrorSuccess(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc)
	wg.Start()

	err := wg.GetLastError()
	if err != nil {
		t.Error("Shouldn`t get errors!")
	}
}

// Test_WaitGroup_GetLastErrorSingleError test
func Test_WaitGroup_GetLastErrorSingleError(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, errorFunc)
	wg.Start()

	err := wg.GetLastError()
	if err == nil {
		t.Error("Should get error!")
	}
}

// Test_WaitGroup_GetAllErrorsSuccess test
func Test_WaitGroup_GetAllErrorsSuccess(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc)
	wg.Start()

	errs := wg.GetAllErrors()
	if len(errs) != 0 {
		t.Error("Shouldn`t get errors!")
	}
}

// Test_WaitGroup_GetAllErrorsManyErrors test
func Test_WaitGroup_GetAllErrorsManyErrors(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, errorFunc, errorFunc)
	wg.Start()

	errs := wg.GetAllErrors()
	if len(errs) != 2 {
		t.Error("Should get errors")
	}
}

// Test_WaitGroup_GetAllErrorsSingleError test
func Test_WaitGroup_GetAllErrorsSingleError(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, errorFunc)
	wg.Start()

	errs := wg.GetAllErrors()
	if len(errs) != 1 {
		t.Error("Should get one error!")
	}
}

// Test_WaitGroup_Reset test
func Test_WaitGroup_Reset(t *testing.T) {
	var wg waitGroup

	wg.Add(fastFunc, errorFunc)
	wg.Start()

	wg.Reset()
	if wg.status != statusIdle {
		t.Error("Cleaned wg should have idle status")
	}

	wg.Add(errorFunc, errorFunc, errorFunc)

	if errs := wg.Start().GetAllErrors(); len(errs) != 3 {
		t.Error("Should get three errors on cleaned wg")
	}
}

// Test_WaitGroup_DoubleStart test
func Test_WaitGroup_DoubleStart(t *testing.T) {
	var wg1, wg2 waitGroup

	wg1.Add(slowFunc, errorFunc, fastFunc, errorFunc)
	chDone := make(chan struct{})
	go func() {
		wg1.Start()
		chDone <- struct{}{}
	}()

	<-chDone
	if errs := wg1.Start().GetAllErrors(); len(errs) != 2 {
		t.Error("Should get two errors on wg")
	}

	chDone2 := make(chan struct{})
	wg2.Add(slowFunc, errorFunc, fastFunc, errorFunc)
	go func() {
		<-chDone2
		wg2.Start()
	}()

	wg2.Start()
	chDone2 <- struct{}{}
	if errs := wg2.GetAllErrors(); len(errs) != 2 {
		t.Error("Should get two errors on wg")
	}
}

var results = make(chan bool, 100)

func fastFuncWithResult(context.Context) error {
	results <- true
	return nil
}

// TestWaitGroup_Timeout test for timeout
func Test_WaitGroup_Timeout_Execution(t *testing.T) {
	var wg waitGroup
	maxProcs := 8 * runtime.NumCPU()

	for i := 0; i < maxProcs; i++ {
		wg.Add(errorFunc)
	}

	for i := 0; i < maxProcs; i++ {
		wg.Add(fastFuncWithResult)
	}

	wg.SetTimeout(time.Microsecond).SetStopOnError(true).Start()

	time.Sleep(time.Second)
	close(results)

	err := wg.GetLastError()
	if err == nil {
		t.Error("WaitGroup should stops with error")
	}

	if errs := wg.GetAllErrors(); len(errs) > 0 {
		for _, err := range errs {
			t.Log(err)
		}
	}

	count := 0
	for range results {
		count++
	}

	if count >= maxProcs {
		t.Errorf("Some wg functions should be interrupted")
	}

	//Debug
	t.Logf("Done %v of %v", count, maxProcs)
}
