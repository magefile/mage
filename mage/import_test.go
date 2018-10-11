package mage

import (
	"bytes"
	"testing"
)

func TestMageImportsList(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
This is a comment on the package which should get turned into output with the list of targets.

Targets:
  somePig*       This is the synopsis for SomePig.
  testVerbose    

* default target
`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}
