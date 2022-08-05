//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
)

const UseKebabTargets = true

type MultiWordNamespace mg.Namespace

func (MultiWordNamespace) MultiWordTask() error {
	fmt.Println("Running MultiWordNamespace.MultiWordTask")
	return nil
}

func MultiWordTask() error {
	fmt.Println("Running MultiWordTask")
	return nil
}

var Aliases = map[string]interface{}{
	"mwn:mwt": MultiWordNamespace.MultiWordTask,
	"mwt":     MultiWordTask,
}
