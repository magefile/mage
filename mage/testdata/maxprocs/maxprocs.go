//+build mage

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/magefile/mage/mg"
)

func Sleep1(ctx context.Context) {
	time.Sleep(200 * time.Millisecond)
}

func Sleep2(ctx context.Context) {
	time.Sleep(200 * time.Millisecond)
}

func Sleep3(ctx context.Context) {
	time.Sleep(200 * time.Millisecond)
}

func Wait(ctx context.Context) {
	start := time.Now()
	mg.CtxDeps(ctx, Sleep1, Sleep2, Sleep3)
	fmt.Println(time.Since(start))
}
