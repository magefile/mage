//+build mage

package main

import (
	"github.com/magefile/mage/mg"
)

var Default = FooBar

func WrongSignature(i int) {
}

func FooBar() {
	mg.Deps(WrongSignature)
}
