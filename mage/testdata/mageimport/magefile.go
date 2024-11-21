//go:build mage
// +build mage

package main

// important things to note:
// * these two packages have the same package name, so they'll conflict
// when imported.
// * one is imported with underscore and one is imported normally.
//
// they should still work normally as mageimports

import (
	"fmt"

	//mage:import
	_ "github.com/magefile/mage/mage/testdata/mageimport/subdir1"
	//mage:import zz
	"github.com/magefile/mage/mage/testdata/mageimport/subdir2"
)

var Aliases = map[string]interface{}{
	"nsd2": mage.NS.Deploy2,
}

var Default = mage.NS.Deploy2

func Root() {
	fmt.Println("root")
}
