//go:build mage
// +build mage

package main

import "github.com/magefile/mage/mage/testdata/mixed_lib_files/subdir"

func Build() {
	subdir.Build()
}
