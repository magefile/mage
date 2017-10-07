// +build amage

package main

import (
	"context"
	"fmt"
	"github.com/magefile/mage/mg"
)

func Build(ctx context.Context {
	mg.Deps(Clean)
	fmt.Println("Building")
}

func Clean() {
	fmt.Println("Cleaning")
}
