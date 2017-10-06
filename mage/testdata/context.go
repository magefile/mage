// +build mage

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/magefile/mage/mg"
)

// Returns a non-nil error.
func TakesContextNoError(ctx context.Context) {
	deadline, _ := ctx.Deadline()
	fmt.Printf("Context timeout: %v\n", deadline)
}

func TakesContextWithError(ctx context.Context) error {
	return errors.New("Something went sideways")
}

func DepsWithContext(ctx context.Context) {
	mg.DepsWithContext(ctx, TakesContextNoError)
}
