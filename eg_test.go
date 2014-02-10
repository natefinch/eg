package eg_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/natefinch/eg"
)

func ExampleNote() {
	err := errors.New("Original error string")
	err = eg.Note(err, "first annotation")
	err = eg.Note(err, "second annotation")

	fmt.Println(err.Error())

	// output:
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

	err := eg.Pass(e, "Error during bar:", isFoo)

	fmt.Printf("Errors are the same: %v", isFoo(err))

	// output
	// Errors are the same: true

}
