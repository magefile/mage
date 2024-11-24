//go:build mage && go1.9
// +build mage,go1.9

package main

// this causes a panic defined in issue #126
// Note this is only valid in go 1.9+, thus the additional build tag above.
type Foo = map[string]string
