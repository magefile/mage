//go:build mage
// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
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
