//go:build mage
// +build mage

package main

import (
	"fmt"
	"log"

	"github.com/magefile/mage/mg"
)

var Default = SomePig

// this should not be a target because it returns a string
func ReturnsString() string {
	fmt.Println("more stuff")
	return ""
}

func TestVerbose() {
	log.Println("hi!")
}

// This is the synopsis for SomePig.  There's more data that won't show up.
func SomePig() {
	mg.Deps(f)
}

func f() {}
