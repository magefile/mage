//go:build mage
// +build mage

package sametarget

import (
	// mage:import samenamespace
	_ "github.com/magefile/mage/mage/testdata/mageimport/samenamespace/duptargets/package1"
	// mage:import samenamespace
	_ "github.com/magefile/mage/mage/testdata/mageimport/samenamespace/duptargets/package2"
)
