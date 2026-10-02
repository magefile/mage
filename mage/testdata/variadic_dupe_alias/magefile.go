//go:build mage
// +build mage

package main

var Aliases = map[string]interface{}{
	"BUILD": Existing,
}

func Existing() {}

func Build(args ...string) {}
