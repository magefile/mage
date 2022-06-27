//go:build mage
// +build mage

package main

import "github.com/magefile/mage/mg"

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

type Nested Init

func (Nested) Foo() error {
	// do your foo defined in init:nested namespace
	return nil
}

func (Nested) Bar() {
	// do your bar defined in init:nested namespace
}

type DoubleNested Nested

func (DoubleNested) Foobar() error {
	// do your foobar defined in init:nested:doublenested namespace
	return nil
}
