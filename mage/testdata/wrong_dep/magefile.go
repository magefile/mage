//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/mg"
)

var Default = FooBar

func WrongSignature(c complex128) {
}

func FooBar() {
	mg.Deps(mg.F(WrongSignature, 0))
}
