//go:build mage

// This is a global comment for the mage output.
// It should be retained with CRLF.
package main

import "github.com/magefile/mage/mg"

// DoIt is a dummy function with a crlf comment.
// That should show up with multiple lines.
func DoIt() {
}

// Sub is a namespace.
// It also has a crlf.
type Sub mg.Namespace

// DoItToo is a dummy function with a crlf comment.
// Here's the second line.
func (Sub) DoItToo() {
}
