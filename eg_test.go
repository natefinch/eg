package eg_test

import (
	"errors"
	"fmt"

	"github.com/natefinch/eg"
)

func ExampleNote() {
	err := errors.New("Original error string")
	err = eg.Note(err, "first annotation")
	err = eg.Note(err, "second annotation")

	fmt.Println(err.Error())

	// Output:
	// second annotation: first annotation: Original error string
}

func ExamplePass() {
	type foo struct {
		*eg.Err
	}

	isFoo := func(err error) bool {
		_, ok := err.(foo)
		return ok
	}

	e := foo{eg.New("foo error")}
	e2 := errors.New("not foo error")

	err := eg.Pass(e, "Error during bar", isFoo)
	err2 := eg.Pass(e2, "Error during bar", isFoo)

	fmt.Printf("Passed error type: %T\n", err)
	fmt.Printf("Passed error string: %v\n", err)
	fmt.Printf("Masked error type: %T\n", err2)
	fmt.Printf("Masked error string: %v\n", err2)

	// Output:
	// Passed error type: eg_test.foo
	// Passed error string: Error during bar: foo error
	// Masked error type: *eg.Err
	// Masked error string: Error during bar: not foo error
}
