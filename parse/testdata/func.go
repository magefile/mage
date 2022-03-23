// +build mage

package main

import (
	"context"
	"fmt"
)

// Synopsis for "returns" error.
// And some more text.
func ReturnsNilError() error {
	fmt.Println("stuff")
	return nil
}

func Shutdown(context.Context) error {
	fmt.Println("shutting down")
	return nil
}

func nonexported() {}
