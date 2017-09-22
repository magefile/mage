// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
)

// Builds things
func Build() { // This one?
	fmt.Println("Build it!")
	os.Exit(111)
}

// Installs some stuff
func Install() error {
	fmt.Println("error soon!")
	return errors.New("boo!")
}

func helper() {

}
