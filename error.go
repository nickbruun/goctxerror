package ctxerror

import (
	"fmt"
	"log"
	"sync"

	"golang.org/x/net/context"
)

// Error handler function.
type ErrorHandlerFunc func(ctx context.Context, err error, msg string)

// Error handler.
type errorHandler struct {
	lock        sync.Mutex
	handlerFunc ErrorHandlerFunc
	handled     map[error]struct{}
}

// Unexported context key type.
type key int

// Error handler context key type.
const errHandlerContextKey key = 0

// New context with error handler assigned.
func NewContext(ctx context.Context, errHandler ErrorHandlerFunc) context.Context {
	return context.WithValue(ctx, errHandlerContextKey, &errorHandler{
		handlerFunc: errHandler,
		handled:     make(map[error]struct{}),
	})
}

// Default error handler.
func defaultErrHandler(ctx context.Context, err error, msg string) {
	log.Printf("Error: %s: %s", msg, err.Error())
}

// Capture error in context with message.
//
// Private to maintain the same number of frames from caller to handler
// function.
func captureMessage(ctx context.Context, err error, msg string) {
	errHandler, ok := ctx.Value(errHandlerContextKey).(*errorHandler)
	if !ok {
		return
	}

	errHandler.lock.Lock()
	_, handled := errHandler.handled[err]
	if !handled {
		errHandler.handled[err] = struct{}{}
	}
	errHandler.lock.Unlock()

	if !handled {
		handlerFunc := errHandler.handlerFunc
		if handlerFunc == nil {
			handlerFunc = defaultErrHandler
		}
		handlerFunc(ctx, err, msg)
	}
}

// Capture error in context.
func Capture(ctx context.Context, err error) {
	captureMessage(ctx, err, err.Error())
}

// Capture error in context with message.
func CaptureMessage(ctx context.Context, err error, msg string) {
	captureMessage(ctx, err, msg)
}

// Capture error in context with message.
func CaptureMessagef(ctx context.Context, err error, msgFormat string, msgArgs ...interface{}) {
	captureMessage(ctx, err, fmt.Sprintf(msgFormat, msgArgs...))
}
