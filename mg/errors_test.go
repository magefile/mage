package mg

import (
	"errors"
	"fmt"
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

func TestExitStatusNil(t *testing.T) {
	code := ExitStatus(nil)
	if code != 0 {
		t.Fatalf("expected 0 for nil error, got %d", code)
	}
}

func TestExitStatusGenericError(t *testing.T) {
	code := ExitStatus(errors.New("generic"))
	if code != 1 {
		t.Fatalf("expected 1 for generic error, got %d", code)
	}
}

func TestExitStatusFatal(t *testing.T) {
	code := ExitStatus(Fatal(42))
	if code != 42 {
		t.Fatalf("expected 42, got %d", code)
	}
}

func TestFatalErrorMessage(t *testing.T) {
	err := Fatal(1, "hello", " ", "world")
	if err.Error() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", err.Error())
	}
}

func TestFatalfErrorMessage(t *testing.T) {
	err := Fatalf(1, "code %d", 42)
	if err.Error() != "code 42" {
		t.Fatalf("expected 'code 42', got %q", err.Error())
	}
}

func TestFatalImplementsExitStatus(t *testing.T) {
	err := Fatal(7, "test")
	var es exitStatus
	if !errors.As(err, &es) {
		// Fatal error implements exitStatus directly via type assertion, not via errors.As,
		// since fatalError doesn't support unwrap. Verify it implements the interface.
		fe, ok := err.(exitStatus)
		if !ok {
			t.Fatal("Fatal error should implement exitStatus interface")
		}
		if fe.ExitStatus() != 7 {
			t.Fatalf("expected exit status 7, got %d", fe.ExitStatus())
		}
	}
}

func TestWrappedFatalExitStatus(t *testing.T) {
	inner := Fatal(42, "inner")
	wrapped := fmt.Errorf("outer: %w", inner)
	// With errors.As, wrapped errors now correctly propagate exit status
	code := ExitStatus(wrapped)
	if code != 42 {
		t.Fatalf("expected 42 for wrapped Fatal, got %d", code)
	}
}
