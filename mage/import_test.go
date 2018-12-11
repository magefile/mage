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
Targets:
  root               
  zz:nS:deploy2*     deploys stuff.
  zz:buildSubdir2    Builds stuff.
  nS:deploy          deploys stuff.
  buildSubdir        Builds stuff.

* default target
`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestMageImportsRoot(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"root"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "root\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestMageImportsNamedNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"zz:nS:deploy2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestMageImportsNamedRoot(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"zz:buildSubdir2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "buildsubdir2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
	if stderr := stderr.String(); stderr != "" {
		t.Fatal("unexpected output to stderr: ", stderr)
	}
}

func TestMageImportsRootImportNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"nS:deploy"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestMageImportsRootImport(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"buildSubdir"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "buildsubdir\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestMageImportsAliasToNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"nsd2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestMageImportsOneLine(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mageimport/oneline",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"build"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "build\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}
