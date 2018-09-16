//+build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

func TestNamespaceDep() {
	mg.Deps(Namespace.SayHi)
}

type Namespace mg.Namespace

func (Namespace) SayHi() {
	fmt.Println("hi!")
}
