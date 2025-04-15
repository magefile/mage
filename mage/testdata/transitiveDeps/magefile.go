//go:build mage
// +build mage

package main

import "github.com/magefile/mage/mage/testdata/transitiveDeps/dep"

func Run() {
	dep.Speak()
}
