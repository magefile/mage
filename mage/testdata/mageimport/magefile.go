//+build mage

package main

import (
	// mage:import
	_ "github.com/magefile/mage/mage/testdata/mageimport/subdir1"
	// mage:import zz
	"github.com/magefile/mage/mage/testdata/mageimport/subdir2"
)

// just something to keep the import from being unused.
var _ = mage.BuildSubdir2

func Root() {

}
