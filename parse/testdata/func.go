// +build mage

package main

import "fmt"

// Synopsis for returns error.
// And some more text.
func ReturnsError() error {
	fmt.Println("stuff")
	return nil
}

func nonexported() {}
