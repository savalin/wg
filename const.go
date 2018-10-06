package wg

import "time"

const (
	// statusIdle means that WaitGroup did not run yet
	statusIdle int = iota
	// statusSuccess means successful execution of all tasks
	statusSuccess
	// statusTimeout means that job was broken by timeout
	statusTimeout
	// statusCaneled means that job was broken by context.CancelFunc call
	statusCaneled
	// statusError means that job was broken by error in one task (if stopOnError is true)
	statusError

	errTimeoutMessage = "Wait group timeout after %v"
	stackBufferSize   = 1000
	defaultMaxTimeout = time.Second * 15
)
