//go:build mage
// +build mage

package main

import (
	//mage:import
	_ "github.com/magefile/mage/mage/testdata/variadic_dupe_import/imported"
)

func Build(args ...string) {}
