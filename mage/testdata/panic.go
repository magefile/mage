// +build mage

package main

// Function that panics.
func Panics() {
	panic("boom!")
}

// Error function that panics.
func PanicsErr() error {
	panic("kaboom!")
}
