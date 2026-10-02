//go:build mage
// +build mage

package main

import (
	"context"
	"time"

	"github.com/magefile/mage/mg"
)

func Variadic(args ...string) {}

func VariadicWithPrefix(ctx context.Context, name string, args ...string) error {
	return nil
}

type VariadicNamespace mg.Namespace

func (VariadicNamespace) Run(args ...string) {}

func OptionalAndVariadic(prefix *string, args ...string) {}

// OptionalTypes exercises every supported pointer-style optional argument type
// before a terminal variadic string argument.
func OptionalTypes(
	ctx context.Context,
	name string,
	text *string,
	count *int,
	ratio *float64,
	enabled *bool,
	timeout *time.Duration,
	args ...string,
) error {
	return nil
}

func VariadicInt(args ...int) {}

func VariadicFloat64(args ...float64) {}

func VariadicBool(args ...bool) {}

func VariadicDuration(args ...time.Duration) {}
