// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

// This should work as a default - even if it's in a different file
var Default = ReturnsError

// this should not be a target because it returns a string
func ReturnsString() string {
	fmt.Println("more stuff")
	return ""
}

func ReturnsVoid() {
	mg.Deps(f)
}

func f() {}
