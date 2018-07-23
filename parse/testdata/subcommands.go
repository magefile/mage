// +build mage
package main

import "github.com/magefile/mage/mg"

// this causes a panic defined in issue #126
type Foo = map[string]string

type Build mg.Namespace

func (Build) Foobar() error {
	// do your foobar build
	return nil
}

func (Build) Baz() {
	// do your baz build
}
