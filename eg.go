// package eg implements improved error handling.
package eg

import (
	"fmt"
	"runtime"
	"strings"
)

// Err is an error that fulfills the Error interface.
type Err struct {
	message     string
	cause       error
	stack       string
	annotations []string
}

// Error is an interface that describes an error which can have a Cause,
// a StackTrace, and Annotations.
type Error interface {
	Cause() Error
	StackTrace() string
	Annotate(msg string)
	Message() string
	error
}

// Annotatable represent an error that can aggregate annotations.
type Annotatable interface {
	Annotate(msg string)
}

// StackTraceable represents an error that can return a stack trace.
type StackTraceable interface {
	StackTrace() string
}

var _ error = (*Err)(nil)

// New returns a new Err object.
func New(msg string, args ...interface{}) *Err {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &Err{
		message: msg,
		stack:   stacktrace(),
	}
}

// Error implements the error interface.
func (e *Err) Error() string {
	msgs := []string{}

	// LIFO the annotations
	for x := len(e.annotations) - 1; x >= 0; x-- {
		msgs = append(msgs, e.annotations[x])
	}
	if e.message != "" {
		msgs = append(msgs, e.message)
	}
	if e.cause != nil {
		msgs = append(msgs, e.cause.Error())
	}
	return strings.Join(msgs, ": ")
}

// Cause returns the error object that caused this error.
func (e *Err) Cause() error {
	return e.cause
}

// StackTrace returns a stack trace at the point where the error was wrapped.
func (e *Err) StackTrace() string {
	return e.stack
}

// Wrap wraps the given error in an Err object and sets the message on the Err
// to msg formatted by args (or unformatted if no args).  If err is
// Stacktraceable, it will copy the stacktrace from err.
func Wrap(err error, msg string, args ...interface{}) *Err {
	stack := ""

	// don't double up stack traces on already-wrapped errors
	if e, ok := err.(StackTraceable); ok {
		stack = e.StackTrace()
	} else {
		stack = stacktrace()
	}

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	return &Err{message: msg, stack: stack, cause: err}
}

// Note annotates the error if it is already an Annotable error, otherwise it
// wraps the error in an Err using msg as the error's message.
func Note(err error, msg string, args ...interface{}) error {
	if a, ok := err.(Annotatable); ok {
		if len(args) == 0 {
			a.Annotate(msg)
		} else {
			a.Annotate(fmt.Sprintf(msg, args))
		}
		return err
	}

	return Wrap(err, msg, args...)
}

// Pass will annotate any errors that match the conditions in iff, and any
// errors which do not match will be wrapped instead.
func Pass(err error, msg string, iff ...func(error) bool) error {
	for _, shouldPass := range iff {
		if shouldPass(err) {
			return Note(err, msg)
		}
	}
	return Wrap(err, msg)
}

// StackTrace returns the stacktrace for the error, or an empty string if the
// error is not Stacktraceable.
func StackTrace(err error) string {
	if e, ok := err.(StackTraceable); ok {
		return e.StackTrace()
	}
	return ""
}

// stacktrace returns a stacktrace for the current goroutine
func stacktrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	for n == len(buf) {
		buf = make([]byte, n+1024)
		n = runtime.Stack(buf, false)
	}
	return string(buf)
}
