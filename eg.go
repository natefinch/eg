// package eg implements improved error handling mechanisms.
//
// eg solves several common problems with Go's native error handling:
//
package eg

import (
	"fmt"
	"runtime"
	"strings"
)

// Annotatable is an interface that represents an error that can aggregate
// messages with associated locations in source code.
type Annotatable interface {
	Annotate(msg, function, file string, line int)
}

// Err is an error that fulfills the Error interface.
type Err struct {
	message     string
	location    location
	cause       error
	annotations []annotation
}

var _ error = (*Err)(nil)

// New returns a new Err object.
func New(msg string, args ...interface{}) *Err {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &Err{
		message:  msg,
		location: locate(1),
	}
}

// Error implements the error interface.
func (e *Err) Error() string {
	msgs := []string{}

	// LIFO the annotations
	for x := len(e.annotations) - 1; x >= 0; x-- {
		msgs = append(msgs, e.annotations[x].String())
	}

	msgs = append(msgs, e.message)

	if e.cause != nil {
		msgs = append(msgs, e.cause.Error())
	}
	return strings.Join(msgs, ": ")
}

// Cause returns the error object that caused this error.
func (e *Err) Cause() error {
	return e.cause
}

// Annotate adds the message to the list of annotations on the error.
func (e *Err) Annotate(msg, function, file string, line int) {
	e.annotations = append(e.annotations,
		annotation{
			Message:  msg,
			location: location{function, file, line},
		})
}

// Details returns a detailed list of annotations including files and line
// numbers.
func (e *Err) Details() string {
	msgs := []string{}

	// LIFO the annotations
	for x := len(e.annotations) - 1; x >= 0; x-- {
		msgs = append(msgs, e.annotations[x].Details())
	}

	msgs = append(msgs, fmt.Sprintf("%s %s", e.location, e.message))

	if e.cause != nil {
		msgs = append(msgs, e.cause.Error())
	}
	return strings.Join(msgs, "\n")

}

// Wrap wraps the given error in an Err object and sets the message on the Err
// to msg formatted by args (or unformatted if no args).  If err is
// Stacktraceable, it will copy the stacktrace from err.
func Wrap(err error, msg string, args ...interface{}) *Err {
	return wrap(err, 1, msg, args...)
}

func wrap(err error, depth int, msg string, args ...interface{}) *Err {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	return &Err{message: msg, cause: err, location: locate(depth + 1)}
}

// Note annotates the error if it is already an Annotable error, otherwise it
// wraps the error in an Err using msg as the error's message.
func Note(err error, msg string, args ...interface{}) error {
	return note(err, 1, msg, args...)
}

func note(err error, depth int, msg string, args ...interface{}) error {
	if a, ok := err.(Annotatable); ok {
		l := locate(depth + 1)
		if len(args) == 0 {
			a.Annotate(msg, l.Function, l.File, l.Line)
		} else {
			a.Annotate(fmt.Sprintf(msg, args), l.Function, l.File, l.Line)
		}
		return err
	}

	return wrap(err, depth+1, msg, args...)
}

// Pass will annotate any errors that match the conditions in iff, and any
// errors which do not match will be wrapped instead.
func Pass(err error, msg string, iff ...func(error) bool) error {
	for _, shouldPass := range iff {
		if shouldPass(err) {
			return note(err, 1, msg)
		}
	}
	return wrap(err, 1, msg)
}

// location is a line in source control
type location struct {
	Function string
	File     string
	Line     int
}

func (l location) String() string {
	return fmt.Sprintf("[%s@%s:%d]", l.Function, l.File, l.Line)
}

// locate returns info about thje line of sourcecode depth levels above the
// caller of locate.
func locate(depth int) location {
	pc, file, line, _ := runtime.Caller(depth + 1)
	function := runtime.FuncForPC(pc).Name()
	return location{function, file, line}
}

// annotation is a message associated with a location.
type annotation struct {
	Message string
	location
}

func (a annotation) String() string {
	return a.Message
}

func (a annotation) Details() string {
	return fmt.Sprintf("%s %s", a.location, a.Message)
}
