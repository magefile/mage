//go:build mage
// +build mage

package main

import (
	"fmt"
)

// No default so we can check the list().

// Prints out 'FooBar'.
func FooBar() {
	fmt.Println("FooBar")
}
