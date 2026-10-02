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

// OptionalAndVariadic prints an optional prefix and pass-through arguments.
func OptionalAndVariadic(prefix *string, args ...string) {
	value := "<nil>"
	if prefix != nil {
		value = *prefix
	}
	fmt.Printf("optional:%s:%q\n", value, args)
}

// OptionalTypes prints fixed, optional, and pass-through arguments.
func OptionalTypes(
	ctx context.Context,
	name string,
	text *string, // text value
	count *int, // count value
	ratio *float64, // ratio value
	enabled *bool, // enabled value
	timeout *time.Duration, // timeout value
	args ...string,
) error {
	_ = ctx
	textValue := "<nil>"
	countValue := "<nil>"
	ratioValue := "<nil>"
	enabledValue := "<nil>"
	timeoutValue := "<nil>"
	if text != nil {
		textValue = *text
	}
	if count != nil {
		countValue = fmt.Sprint(*count)
	}
	if ratio != nil {
		ratioValue = fmt.Sprint(*ratio)
	}
	if enabled != nil {
		enabledValue = fmt.Sprint(*enabled)
	}
	if timeout != nil {
		timeoutValue = timeout.String()
	}
	fmt.Printf("optionaltypes:%s:%s:%s:%s:%s:%s:%q\n", name, textValue, countValue, ratioValue, enabledValue, timeoutValue, args)
	return nil
}

func VariadicInt(args ...int) {}
