package internal

// wrappedError is an error that supports the Go 1.13+ mechanism of wrapping
// errors. It implements Unwrap to return the underlying error, but it still
// returns the string version of whatever "main" error it represents.
type wrappedError struct {
	underlyingError error
	stringError     error
}

// WrapError returns an error that implements the Go 1.13 Unwrap interface. The
// Error function returns the value of the "string" error, and the Unwrap
// function returns the "underlying" error. Use this wherever one might
// otherwise use the "%w" format string in [fmt.Errorf].
//     err := doSomething()
//     if err != nil {
//         return WrapError(err, fmt.Errorf("Could not do something: %v", err))
//     }
// The premise is that the "string" error is not an interesting one to try to
// inspect with [errors.Is] or [errors.As] because it has no other wrapped
// errors of its own, and it will usually be of whatever error type
// `fmt.Errorf` returns.
func WrapError(underlying, str error) error {
	return &wrappedError{
		underlyingError: underlying,
		stringError:     str,
	}
}

func (e *wrappedError) Error() string {
	return e.stringError.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.underlyingError
}
