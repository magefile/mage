package internal

import "fmt"

// wrappedError is an error that supports the Go 1.13+ mechanism of wrapping
// errors. It implements Unwrap to return the underlying error, but it still
// returns the string version of whatever "main" error it represents.
type wrappedError struct {
	underlyingError error
	stringError     error
}

// WrapErrorf returns an error that implements the Go 1.13 Unwrap interface.
// The Error function returns the message calculated from formatting the
// arguments, just like [fmt.Errorf], and the Unwrap function returns the
// "underlying" error. Use this wherever one might otherwise use the "%w"
// format string in [fmt.Errorf].
//
//	err := doSomething()
//	if err != nil {
//	    return WrapErrorf(err, "Could not do something: %v", err)
//	}
//
// Although the "%w" format string will be recognized in versions of Go that
// support it, any error wrapped by it will not be included in this error; only
// underlying is considered wrapped by this error.
func WrapErrorf(underlying error, format string, args ...interface{}) error {
	return &wrappedError{
		underlyingError: underlying,
		stringError:     fmt.Errorf(format, args...),
	}
}

func (e *wrappedError) Error() string {
	return e.stringError.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.underlyingError
}
