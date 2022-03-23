// +build mage

package main

import "fmt"

var Deinit = Shutdown

func TargetWithCustomShutdown() error {
	return nil
}

func Shutdown() {
	fmt.Println("Shutting down")
}
