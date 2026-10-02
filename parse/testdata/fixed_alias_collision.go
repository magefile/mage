//go:build mage
// +build mage

package main

// Aliases preserves the historical behavior where an alias may shadow a
// fixed-arity target with the same case-insensitive name.
var Aliases = map[string]interface{}{
	"BUILD": Existing,
}

// Existing is the target selected through the BUILD alias.
func Existing() {}

// Build is the fixed-arity target shadowed by the BUILD alias.
func Build() {}
