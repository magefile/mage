//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Greet greets someone with an optional greeting.
func Greet(name string, greeting *string) {
	if greeting != nil {
		fmt.Printf("%s, %s!\n", *greeting, name)
	} else {
		fmt.Printf("Hello, %s!\n", name)
	}
}

// Add adds two numbers. The second number is optional and defaults to 0.
func Add(a int, b *int) {
	if b != nil {
		fmt.Println(a + *b)
	} else {
		fmt.Println(a)
	}
}

// Scale scales a value by an optional factor.
func Scale(value float64, factor *float64) {
	if factor != nil {
		fmt.Printf("%.1f\n", value*(*factor))
	} else {
		fmt.Printf("%.1f\n", value)
	}
}

// Run runs with an optional verbose flag.
func Run(verbose *bool) {
	if verbose != nil && *verbose {
		fmt.Println("running verbose")
	} else {
		fmt.Println("running quiet")
	}
}

// Delay delays with an optional extra duration.
func Delay(base time.Duration, extra *time.Duration) {
	if extra != nil {
		fmt.Printf("delay %s + %s\n", base, *extra)
	} else {
		fmt.Printf("delay %s\n", base)
	}
}

// AllOptional takes only optional args.
func AllOptional(a *string, b *int) {
	if a != nil {
		fmt.Printf("a=%s\n", *a)
	} else {
		fmt.Println("a=<nil>")
	}
	if b != nil {
		fmt.Printf("b=%d\n", *b)
	} else {
		fmt.Println("b=<nil>")
	}
}

// Say says the message with optional capitalization and repeat count.
func Say(msg string, cap *bool, count *int) {
	if cap != nil && *cap {
		msg = strings.ToUpper(msg)
	}
	repeat := 1
	if count != nil {
		repeat = *count
	}
	for i := 0; i < repeat; i++ {
		fmt.Println(msg)
	}
}

// Announce prints an announcement.
func Announce(msg string) {
	fmt.Printf("Announcement: %s\n", msg)
}

// Mixed tests interleaved required and optional args with context.
func Mixed(ctx context.Context, name string, greeting *string, count int) error {
	g := "Hello"
	if greeting != nil {
		g = *greeting
	}
	for i := 0; i < count; i++ {
		fmt.Printf("%s, %s!\n", g, name)
	}
	return nil
}
