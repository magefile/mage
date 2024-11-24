//go:build mage
// +build mage

package main

import "fmt"

var Aliases = map[string]interface{}{
	"st":   Status,
	"stat": Status,
	"co":   Checkout,
}

// Prints status.
func Status() {
	fmt.Println("alias!")
}

func Checkout() {}
