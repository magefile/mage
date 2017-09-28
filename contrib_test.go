package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestGoFormatted(t *testing.T) {
	s, err := run("gofmt", "-l", ".")
	if s != "" {
		t.Fatalf("the following files are not gofmt'ed:\n%s", s)
	}
	if err != nil {
		t.Error(err)
	}
}

func TestGoVet(t *testing.T) {
	s, err := run("go", "vet", "./...")
	if s != "" {
		t.Fatalf("go vet fails with:\n%s", s)
	}
	if err != nil {
		t.Error(err)
	}
}

func run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	return string(b), err
}
