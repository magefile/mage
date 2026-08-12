//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/magefile/mage/mg"
	//mage:import shared
	_ "github.com/magefile/mage/mage/testdata/variadic/imported"
)

var Default = Variadic

var Aliases = map[string]interface{}{
	"v": Variadic,
}

// Variadic prints all remaining arguments.
func Variadic(args ...string) {
	fmt.Printf("variadic:%q\n", args)
}

// Fixed prints one fixed argument.
func Fixed(value string) {
	fmt.Printf("fixed:%s\n", value)
}

// Collect prints its fixed prefix and all remaining arguments.
func Collect(ctx context.Context, prefix string, args ...string) error {
	fmt.Printf("collect:%s:%q\n", prefix, args)
	return nil
}

// Types prints converted fixed arguments followed by variadic arguments.
func Types(count int, ratio float64, enabled bool, timeout time.Duration, args ...string) {
	fmt.Printf("types:%d:%.1f:%t:%s:%q\n", count, ratio, enabled, timeout, args)
}

// Tools groups variadic targets.
type Tools mg.Namespace

// Collect prints namespace arguments.
func (Tools) Collect(prefix string, args ...string) {
	fmt.Printf("tools:collect:%s:%q\n", prefix, args)
}

func OptionalAndVariadic(prefix *string, args ...string) {}

func VariadicInt(args ...int) {}
