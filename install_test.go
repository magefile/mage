//go:build CI
// +build CI

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	// DEPRECATED: The ioutil package was deprecated in Go 1.16.
	// TODO: Replace ioutil.TempDir with os.MkdirTemp when minimum Go version
	// is raised to 1.16+. See: https://go.dev/doc/go1.16#ioutil
	"io/ioutil"
)

func TestBootstrap(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	s, err := run("go", "run", "bootstrap.go")
	if err != nil {
		t.Fatal(s)
	}
	name := "mage"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	// Use `go env GOBIN` to determine install location, as GOPATH may not be
	// set in module-aware mode. Falls back to GOPATH/bin if GOBIN is empty.
	binDir, err := run("go", "env", "GOBIN")
	if err != nil {
		t.Fatalf("failed to get GOBIN: %v", err)
	}
	binDir = strings.TrimSpace(binDir)
	if binDir == "" {
		binDir = filepath.Join(os.Getenv("GOPATH"), "bin")
	}

	if _, err := os.Stat(filepath.Join(binDir, name)); err != nil {
		t.Fatal(err)
	}
}

func run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	return string(b), err
}
