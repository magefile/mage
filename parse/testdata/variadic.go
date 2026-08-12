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

func VariadicInt(args ...int) {}

func VariadicFloat64(args ...float64) {}

func VariadicBool(args ...bool) {}

func VariadicDuration(args ...time.Duration) {}
