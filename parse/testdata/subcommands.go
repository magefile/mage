//go:build mage
// +build mage

package main

import "github.com/magefile/mage/mg"

// Foo is a type alias to test that type aliases don't cause panics during
// parsing. See issue #126. Type aliases were introduced in Go 1.9, which is
// now well below the minimum supported version (1.12).
type Foo = map[string]string

type Build mg.Namespace

func (Build) Foobar() error {
	// do your foobar build
	return nil
}

func (Build) Baz() {
	// do your baz build
}

type Init mg.Namespace

func (Init) Foobar() error {
	// do your foobar defined in init namespace
	return nil
}
