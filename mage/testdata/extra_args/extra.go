//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
)

func TargetOne(extra mg.ExtraArgs) {
	fmt.Printf("%#v\n", extra)
}

func TargetTwo(ctx context.Context, extra mg.ExtraArgs) {
	fmt.Printf("Context is nil: %t\n", ctx == nil)
	fmt.Printf("%#v\n", extra)
}

func TargetThree(ctx context.Context, extra mg.ExtraArgs, aString string) error {
	fmt.Printf("Context is nil: %t\n", ctx == nil)
	fmt.Printf("%#v\n", extra)
	fmt.Printf("%s\n", aString)
	return nil
}

func TargetFour(ctx context.Context, extra mg.ExtraArgs, aString string, anInt int) error {
	fmt.Printf("Context is nil: %t\n", ctx == nil)
	fmt.Printf("%#v\n", extra)
	fmt.Printf("%s\n", aString)
	fmt.Printf("%d\n", anInt)
	return nil
}

func TargetFive(extra mg.ExtraArgs, aString string, anInt int) error {
	fmt.Printf("%#v\n", extra)
	fmt.Printf("%s\n", aString)
	fmt.Printf("%d\n", anInt)
	return nil
}

func TargetSix(aString string, anInt int, extra mg.ExtraArgs) error {
	fmt.Printf("%#v\n", extra)
	fmt.Printf("%s\n", aString)
	fmt.Printf("%d\n", anInt)
	return nil
}
