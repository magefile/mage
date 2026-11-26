//go:build mage
// +build mage

package main

import "fmt"

var Default = Build

// Build prints the version argument it is given.
func Build(version string) error {
	fmt.Println("got:", version)
	return nil
}
