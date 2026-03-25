//go:build mage

// This is a global comment for the mage output.
// It should retain line returns.
package main

//mage:multiline // Enable multiline description output.

import "github.com/magefile/mage/mg"

// DoIt is a dummy function with a multiline comment.
// That should show up with multiple lines.
func DoIt() {
}

// Sub is a namespace.
// It also has a line return.
type Sub mg.Namespace

// DoItToo is a dummy function with a multiline comment.
// Here's the second line.
func (Sub) DoItToo() {
}
