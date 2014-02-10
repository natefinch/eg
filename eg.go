// Package eg implements improved error handling mechanisms.
//
// This package solves several common problems with Go's native error handling:
//
// Tracebacks with context to help understand where an error came from.
//
// The ability to wrap an erro with a new error without losing the context of
// the original error.
//
// A way to print out more detailed information about an error.
//
// Examples:
//  type NotFoundError {
//		*eg.Err
//	}
//
//	func IsNotFound(err error) bool {
//		_, ok := err.(NotFoundError)
//		return ok
//	}
//
//	func GetConfig() []byte, error {
//		data, err := ioutil.ReadFile("config_file")
//		if os.IsNotExists(err) {
//			// return a new error with the original error as the cause
//			return nil, NotFoundError{eg.Wrap(err, "Couldn't find config file")}
//		}
//		if err != nil {
//			// return a generic error for other problems
//			return eg.Wrap(err, "Error reading config file")
//		}
//		return data, nil
//	}
//
//	func StartFoo() error {
//		data, err := GetConfig()
//		if err != nil {
//			// only let the IsNotFound error percolate up, so we don't let callers
//			// depend on implementation-specific errors.
//			return eg.Pass(err, "Can't start foo", IsNotFound)
//		}
//		// <start foo>
//		return nil
//	}
//
//	func Bootstrap() error {
//		err := StartFoo()
// 		if err != nil {
//			// add context to the error
//			return eg.Note(err, "Can't bootstrap")
//		}
//		// <bootstrap stuff>
//		return nil
//	}
//
//	func main() {
//		fmt.Printf("%v", Bootstrap())
//	}
//
//	// Output:
// 	// Can't bootstrap: Can't start foo: Couldn't find config file: open config_file: file or directory not found
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

// Effect is an interface that represents an error that can have a cause.
type Effect interface {
	Cause() error
}

// Detailed is an interface that represents an error that can returned detailed
// information.
type Detailed interface {
	Details() string
}

// Err is an an error that implements Annotatable, Effect, and Detailed.
type Err struct {
	message     string
	location    location
	cause       error
	annotations []annotation
}

var _ error = (*Err)(nil)

// Mask returns a new Err object with a message based on the given error's
// message but without listing the error as the Cause.
func Mask(err error, msg string, args ...interface{}) *Err {
	return mask(err, 1, msg, args...)
}

func mask(err error, depth int, msg string, args ...interface{}) *Err {
	ret := newErr(depth+1, msg, args...)
	if err != nil {
		ret.message = ret.message + ": " + err.Error()
	}
	return ret
}

// New returns a new Err object with the given message.
func New(msg string, args ...interface{}) *Err {
	return newErr(1, msg, args...)
}

func newErr(depth int, msg string, args ...interface{}) *Err {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &Err{
		message:  msg,
		location: locate(depth + 1),
	}

}

// Error implements the error interface.
func (e *Err) Error() string {
	msgs := []string{}

	// LIFO the annotations
	for x := len(e.annotations) - 1; x >= 0; x-- {
		msg := e.annotations[x].String()
		if msg != "" {
			msgs = append(msgs, e.annotations[x].String())
		}
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

// Annotate adds the message to the list of annotations on the error.  If msg is
// empty, the annotation will only be displayed when printing the error's
// details.
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
		msgs = append(msgs, Details(e.cause))
	}
	return strings.Join(msgs, "\n")
}

// Wrap wraps the given error in an Err object, in effect obscuring the original
// error.
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

// Pass will Note any errors that match the conditions in iff, and Mask any
// errors which do not match.
func Pass(err error, msg string, iff ...func(error) bool) error {
	for _, shouldPass := range iff {
		if shouldPass(err) {
			return note(err, 1, msg)
		}
	}
	return mask(err, 1, msg)
}

// Cause returns the cause of the error.  If the error has a cause, ok will be
// true, and cause will contain the cause.  Otherwise the err will be returned
// as the cause.
func Cause(err error) (cause error, ok bool) {
	if err == nil {
		return nil, false
	}
	e, ok := err.(Effect)
	if !ok {
		return err, false
	}
	return e.Cause(), true
}

// Details returns detailed information about the error, or the error's Error()
// string if no detailed information is available.
func Details(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(Detailed); ok {
		return e.Details()
	}
	return err.Error()
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
