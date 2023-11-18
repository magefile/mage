//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Exits after receiving SIGHUP
func ExitsAfterSighup(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGHUP)
	<-sigC
	fmt.Println("received sighup")
}

// Exits after SIGINT and wait
func ExitsAfterSigint(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	<-sigC
	fmt.Printf("exiting...")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

// Exits after ctx cancel and wait
func ExitsAfterCancel(ctx context.Context) {
	defer func() {
		fmt.Println("deferred cleanup")
	}()
	<-ctx.Done()
	fmt.Printf("exiting...")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

// Ignores all signals, requires killing via timeout or second SIGINT
func IgnoresSignals(ctx context.Context) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT)
	for {
		<-sigC
	}
}
