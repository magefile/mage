//go:build mage
// +build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

// mage:import
import _ "github.com/magefile/mage/magefiles/targets"
