// +build mage

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Builds things.
// And you know, other stuff.
func Build() {
	log.Println("Build it!")
	os.Exit(111)
}

// Installs some stuff
func Install() error {
	fmt.Println("error soon!")
	return errors.New("boo!")
}

func helper() {

}
