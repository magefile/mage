//go:build mage

// Package doc for help_no_compile.
package main

import "github.com/magefile/mage/mg"

// Build compiles the project.
func Build() {
	// This references a package that doesn't exist, so go build will fail,
	// but the AST parser can still extract target metadata.
	doesnotexist.Fail()
}

// Deploy pushes to production.
// This is the extended description.
func Deploy() {
	doesnotexist.Fail()
}

// NS is a namespace for grouped targets.
type NS mg.Namespace

// Run runs within the namespace.
func (NS) Run() {
	doesnotexist.Fail()
}
