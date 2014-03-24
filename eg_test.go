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
