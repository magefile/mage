//go:build mage
// +build mage

package main

var Aliases = map[string]interface{}{
	"void": ReturnsVoid,
	"baz":  Build.Baz,
}
