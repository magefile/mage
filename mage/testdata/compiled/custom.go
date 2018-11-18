// +build mage

// Compiled package description.
package main

import (
	"log"
	"time"

	"github.com/magefile/mage/mg"
)

var Default = Deploy

// This is very verbose.
func TestVerbose() {
	log.Println("hi!")
}

// This is the synopsis for Deploy. This part shouldn't show up.
func Deploy() {
	mg.Deps(f)
}

// Sleep sleeps 5 seconds.
func Sleep() {
	time.Sleep(5 * time.Second)
}

func f() {
	log.Println("i am independent -- not")
}
