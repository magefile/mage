//go:build mage
// +build mage

package main

import (
	"fmt"

	//mage:import
	_ "github.com/magefile/mage/parse/testdata/importself"
)

func Build() {
	fmt.Println("built")
}
