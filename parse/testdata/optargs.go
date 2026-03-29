//go:build mage
// +build mage

package main

import "time"

func OptionalString(name string, greeting *string) {}

func OptionalInt(a int, b *int) {}

func OptionalFloat64(value float64, factor *float64) {}

func OptionalBool(verbose *bool) {}

func OptionalDuration(base time.Duration, extra *time.Duration) {}

func AllOptional(a *string, b *int) {}

func FlagDocFunc(name string,
	greeting *string, // the greeting message
	count *int, // how many times
) {
}
