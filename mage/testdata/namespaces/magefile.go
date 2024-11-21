//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"

	"github.com/magefile/mage/mg"
)

var Default = NS.Error

func TestNamespaceDep() {
	mg.Deps(NS.Error, NS.Bare, NS.BareCtx, NS.CtxErr)
}

type NS mg.Namespace

func (NS) Error() error {
	fmt.Println("hi!")
	return nil
}

func (NS) Bare() {
}

func (NS) BareCtx(ctx context.Context) {
}
func (NS) CtxErr(ctx context.Context) error {
	return nil
}
