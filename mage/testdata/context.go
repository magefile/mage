// +build mage

package main

import "context"

// Returns a non-nil error.
func TakesContextNoError(ctx context.Context)  {
	fmt.Printf("Context timeout: %v\n", ctx.Timeout)
}

func TakesContextWithError(ctx context.Context) error {
  return errors.New("Something went sideways")
}

func DepWithContext(ctx context.Context) {
  mg.DepWithContext(ctx, TakesContextNoError)
}
