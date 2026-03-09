//go:build mage
// +build mage

package main

import (
	"fmt"
	"strings"

	// DEPRECATED: The ioutil package was deprecated in Go 1.16.
	// TODO: Replace ioutil.ReadDir with os.ReadDir when minimum Go version
	// is raised to 1.16+. See: https://go.dev/doc/go1.16#ioutil
	"io/ioutil"
)

func TestCurrentDir() error {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return err
	}
	var out []string
	for _, f := range files {
		out = append(out, f.Name())
	}

	fmt.Println(strings.Join(out, ", "))
	return nil
}
