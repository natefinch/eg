
# eg
    import "github.com/natefinch/eg"

Package eg implements improved error handling mechanisms.

This package solves several common problems with Go's native error handling:

Tracebacks with context to help understand where an error came from.

The ability to wrap an error with a new error without losing the context of
the original error.

A way to print out more detailed information about an error.

A way to mask some or all of the errors coming out of a function with
anonymous errors to prevent deeep coupling.

Examples:


	type NotFoundError {
		*eg.Err
	}
	
	func IsNotFound(err error) bool {
		_, ok := err.(NotFoundError)
		return ok
	}
	
	func GetConfig() []byte, error {
		data, err := ioutil.ReadFile("config_file")
		if os.IsNotExists(err) {
			// Return a new error with the original error as the cause.
			return nil, NotFoundError{eg.Wrap(err, "Couldn't find config file")}
		}
		if err != nil {
			// Return a generic error for other problems.
			return eg.Wrap(err, "Error reading config file")
		}
		return data, nil
	}
	
	func StartFoo() error {
		data, err := GetConfig()
		if err != nil {
			// Only let the IsNotFound error percolate up, so we don't let
			// callers depend on implementation-specific errors.
			return eg.Pass(err, "Can't start foo", IsNotFound)
		}
		return nil
	}
	
	func Bootstrap() error {
		err := StartFoo()
		if err != nil {
			// Add context to the error.
			return eg.Note(err, "Can't bootstrap")
		}
		return nil
	}
	
	func main() {
		err := Bootstrap()
		fmt.Printf("%v", err)
	}
	
	// Output:
	// Can't bootstrap: Can't start foo: Couldn't find config file: open config_file: file or directory not found


## func Cause
``` go
func Cause(err error) (cause error, ok bool)
```
Cause returns the cause of the error.  If the error has a cause, ok will be
true, and cause will contain the cause.  Otherwise the err will be returned
as the cause.


## func Details
``` go
func Details(err error) string
```
Details returns detailed information about the error, or the error's Error()
string if no detailed information is available.


## func Note
``` go
func Note(err error, msg string, args ...interface{}) error
```
Note annotates the error if it is already an Annotable error, otherwise it
wraps the error in an Err using msg as the error's message.


## func Pass
``` go
func Pass(err error, msg string, iff ...func(error) bool) error
```
Pass will Note any errors that match the conditions in iff, and Mask any
errors which do not match.



## type Annotatable
``` go
type Annotatable interface {
    Annotate(msg, function, file string, line int)
}
```
Annotatable is an interface that represents an error that can aggregate
messages with associated locations in source code.











## type Detailed
``` go
type Detailed interface {
    Details() string
}
```
Detailed is an interface that represents an error that can returned detailed
information.











## type Effect
``` go
type Effect interface {
    Cause() error
}
```
Effect is an interface that represents an error that can have a cause.











## type Err
``` go
type Err struct {
    // contains filtered or unexported fields
}
```
Err is an an error that implements Annotatable, Effect, and Detailed.









### func Error
``` go
func Error(msg string, args ...interface{}) *Err
```
Error returns a new Err object with the given message.


### func Mask
``` go
func Mask(err error, msg string, args ...interface{}) *Err
```
Mask returns a new Err object with a message based on the given error's
message but without listing the error as the Cause.


### func Wrap
``` go
func Wrap(err error, msg string, args ...interface{}) *Err
```
Wrap wraps the given error in an Err object, in effect obscuring the original
error.




### func (\*Err) Annotate
``` go
func (e *Err) Annotate(msg, function, file string, line int)
```
Annotate adds the message to the list of annotations on the error.  If msg is
empty, the annotation will only be displayed when printing the error's
details.



### func (\*Err) Cause
``` go
func (e *Err) Cause() error
```
Cause returns the error object that caused this error.



### func (\*Err) Details
``` go
func (e *Err) Details() string
```
Details returns a detailed list of annotations including files and line
numbers.



### func (\*Err) Error
``` go
func (e *Err) Error() string
```
Error implements the error interface.