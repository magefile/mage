//go:build mage
// +build mage

package main

import (
	// mage:import samenamespace
	_ "github.com/magefile/mage/mage/testdata/mageimport/samenamespace/uniquetargets/package1"
	// mage:import samenamespace
	_ "github.com/magefile/mage/mage/testdata/mageimport/samenamespace/uniquetargets/package2"
)
