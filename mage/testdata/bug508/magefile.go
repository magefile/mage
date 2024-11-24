//go:build mage
// +build mage

package main

import (
	//mage:import
	"github.com/magefile/mage/mage/testdata/bug508/deps"
)

var Default = deps.Test
