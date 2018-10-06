[![Build Status](https://api.travis-ci.com/savalin/waitgroup.svg)](http://travis-ci.com/savalin/waitgroup) [![GoDoc](https://godoc.org/github.com/savalin/waitgroup?status.svg)](http://godoc.org/github.com/savalin/waitgroup)

# WaitGroup
Enhanced go wait group implementation.

## Features:
- stop all goroutines on an error in one of them
- execute all goroutines and collect errors
- goroutines panic handling
- timeouts: context and global

*If you don't need at least one of them please use [sync.WaitGroup](https://golang.org/pkg/sync/#WaitGroup) from the standard library.

## Usage
### General
```go
wg := wg.New()

// optional 
wg.WithContext(ctx)

wg.Add(
    // Add first goroutine
    func(ctx context.Context) error {
        // some logic
        return nil
    },
    
    // Add second goroutine
    func(ctx context.Context) error {
        // some logic
        return nil
    }        
)

wg.Start()
```

### Error handling
#### stop on error
```go
wg := wg.New()

// add some gouroutines to be executed
// ...

var err error
err = wg.
        SetStopOnError(true).
        Start().
        GetLastError()

// handle the error from failed goroutine
// ...
```

#### execute all goroutines and collect errors
```go
wg := wg.New()

// add some gouroutines to be executed
// ...

var errs []error
errs = wg.
        Start().
        GetAllErrors()

// handle slice of errors from failed goroutines
// ...

```

### Timeout
#### if you have a context
```go
var ( 
    wg = wg.New()
    ctx, cancel = context.WithTimeout(context.TODO(), time.Second)
)

wg.WithContext(ctx)

// add some gouroutines to be executed
// ...

wg.Start()

```

#### if you don't have a context
```go
wg := wg.New()

// add some gouroutines to be executed
// ...

wg.
    SetTimeout(time.Second).
    Start()

```

### Reusage
```go
wg := wg.New()

// add some gouroutines to be executed
// ...

wg.Start()

// RESET wg state for further usage
wg.Reset()

// you can add goroutines and set new parameters again
// ...

```
