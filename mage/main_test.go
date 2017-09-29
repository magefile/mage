package mage

import (
	"os/exec"
	"testing"
)

func TestGoRun(t *testing.T) {
	c := exec.Command("go", "run", "main.go")
	c.Dir = "./testdata"
	b, err := c.CombinedOutput()
	if err != nil {
		t.Error("error:", err)
	}
	actual := string(b)
	expected := "stuff\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}
