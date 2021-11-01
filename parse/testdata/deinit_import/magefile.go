//go:build mage
// +build mage

package main

import (
	// mage:import
	"github.com/magefile/mage/parse/testdata/deinit_import/nested"
)

var Deinit = nested.Shutdown
