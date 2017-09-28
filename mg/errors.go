package mg

import (
	"fmt"

	"github.com/pkg/errors"
)

type fatalErr struct {
	code int
	error
}

func (f fatalErr) ExitStatus() int {
	return f.code
}

type exitStatus interface {
	ExitStatus() int
}

// Fatal returns an error that will cause mage to print out the
// given args and exit with the given exit code.
func Fatal(code int, args ...interface{}) error {
	return fatalErr{
		code:  1,
		error: errors.New(fmt.Sprint(args...)),
	}
}

// Fatalf returns an error that will cause mage to print out the
// given message and exit with an exit code of 1.
func Fatalf(code int, format string, args ...interface{}) error {
	return fatalErr{
		code:  1,
		error: errors.Errorf(format, args...),
	}
}
