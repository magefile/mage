// +build mage

package main

import (
	"fmt"
	"io"
	"os"
)

// Synopsis for "returns" error.
// And some more text.
func ReturnsError() error {
	fmt.Println("stuff")
	return nil
}

func CopyStdin() error {
	_, err := io.Copy(os.Stdout, os.Stdin)
	return err
}

func nonexported() {}
