//go:build mage
// +build mage

package main

import "fmt"

// Synopsis for "returns" error.
// And some more text.
func ReturnsNilError() error {
	fmt.Println("stuff")
	return nil
}

func nonexported() {}

// WithBackticks has a synopsis that includes `backticks` which were a problem once.
func WithBackticks() {}
