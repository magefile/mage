// +build mage

package main

import (
	"fmt"
)

// This should work as a default - even if it's in a different file
var Default = Noop

// this should not be a target because it returns a string
func Noop() {
	fmt.Println("noop")
}
