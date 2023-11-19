//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

func TargetOne(extra mg.ExtraArgs) {
	fmt.Printf("%#v\n", extra)
}
