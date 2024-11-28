// The concept of "wrapping" errors was only introduced in Go 1.13, so we only
// want to test that our errors behave like wrapped errors on Go versions that
// support `errors.Is`.
//go:build go1.13
// +build go1.13

package sh

import (
	"errors"
	"os"
	"testing"
)

func TestWrappedError(t *testing.T) {
	_, err := Exec(nil, nil, nil, os.Args[0]+"-doesnotexist", "-printArgs", "foo")
	if err == nil {
		t.Fatalf("Expected to fail running %s", os.Args[0]+"-doesnotexist")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expected error to be like ErrNotExist, got %#v", err)
	}
}
