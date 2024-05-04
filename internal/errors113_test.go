// The concept of "wrapping" errors was only introduced in Go 1.13, so we only
// want to test that our errors behave like wrapped errors on Go versions that
// support `errors.Is`.
//go:build go1.13
// +build go1.13

package internal

import (
	"errors"
	"testing"
)

// TestErrorUnwrap checks that [errors.Is] can detect the underlying error in a
// wrappedError.
func TestErrorUnwrap(t *testing.T) {
	strError := errors.New("main error")
	underlyingError := errors.New("underlying error")
	actual := WrapError(underlyingError, strError)

	if !errors.Is(actual, underlyingError) {
		t.Fatalf("Expected outer error %#v to include %#v", actual, underlyingError)
	}
}
