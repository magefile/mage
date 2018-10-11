//+build mage

package main

// important things to note:
// * these two packages have the same package name, so they'll conflict
// when imported.
// * one is imported with underscore and one is imported normally.
//
// they should still work normally as mageimports

import (
	"fmt"

	// mage:import
	_ "github.com/magefile/mage/mage/testdata/mageimport/subdir1"
	// mage:import zz
	"github.com/magefile/mage/mage/testdata/mageimport/subdir2"
)

// just something to keep the import from being unused.
var _ = mage.BuildSubdir2

func Root() {
	fmt.Println("root")
}
