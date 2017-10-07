// +build mage

package main

import (
	"context"
	"fmt"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"os"
	"os/exec"
	"time"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

func Timeout(ctx context.Context) {
	fmt.Printf("timeout start\n")
	time.Sleep(1000000000000)
	fmt.Printf("timeout end\n")
}

// A build step that requires additional params, or platform specific steps for example
func Build(ctx context.Context) error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "./MyApp2")
}

// Manage your deps, or running package managers.
func InstallDeps(ctx context.Context) error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
