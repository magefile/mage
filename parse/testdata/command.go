// +build mage

package main

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
)

// This should work as a default - even if it's in a different file
var Default = ReturnsNilError

// This should work as a deinit - even if it's in a different file
var Deinit = Shutdown

// this should not be a target because it returns a string
func ReturnsString() string {
	fmt.Println("more stuff")
	return ""
}

func ReturnsVoid() {
	mg.Deps(f)
}

func f() {}

func TakesContextReturnsVoid(ctx context.Context) {

}

func TakesContextReturnsError(ctx context.Context) error {
	return nil
}
