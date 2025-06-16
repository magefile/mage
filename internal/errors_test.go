package internal

import (
	"errors"
	"testing"
)

// TestBasicWrappedError confirms that a wrappedError returns its string
// message, not that of its "underlying" error.
func TestBasicWrappedError(t *testing.T) {
	message := "main message"
	underlyingError := errors.New("underlying error")
	actual := WrapErrorf(underlyingError, message)

	if actual.Error() != message {
		t.Fatalf("Expected outer error to have Error() = %q but got %q", message, actual.Error())
	}
}

// TestWrapFormat checks that the arguments get formatted, just like
// [fmt.Sprintf].
func TestWrapFormat(t *testing.T) {
	underlyingError := errors.New("underlying error")
	actual := WrapErrorf(underlyingError, "%s %s", "main", "message")

	if actual.Error() != "main message" {
		t.Fatalf("Expected outer error to have formatted message, but got %q", actual.Error())
	}
}
