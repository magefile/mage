package mg

import (
	"errors"
	"testing"
)

func TestFatalExit(t *testing.T) {
	expected := 99
	code := ExitStatus(Fatal(expected))
	if code != expected {
		t.Fatalf("Expected code %v but got %v", expected, code)
	}
}

func TestFatalfExit(t *testing.T) {
	expected := 99
	code := ExitStatus(Fatalf(expected, "boo!"))
	if code != expected {
		t.Fatalf("Expected code %v but got %v", expected, code)
	}
}

// TestBasicWrappedError confirms that a wrappedError returns the same string
// as its "str" error (not its "underlying" error).
func TestBasicWrappedError(t *testing.T) {
	strError := errors.New("main error")
	underlyingError := errors.New("underlying error")
	actual := WrapError(underlyingError, strError)

	if actual.Error() != strError.Error() {
		t.Fatalf("Expected outer error to have Error() = %q but got %q", strError.Error(), actual.Error())
	}
}
