package sh

import (
	"bytes"
	"os"
	"testing"
)

func TestOutCmd(t *testing.T) {
	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd("baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Errorf("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, os.Args[0], "-printVar", env)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("MAGE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}

func TestAutoExpandPrecedent(t *testing.T) {
	// Environment variables passed to OutputWith should take precedence
	// over any variables set in the actual environment.
	if err := os.Setenv("MAGE_FOO", "wrong"); err != nil {
		t.Fatal(err)
	}
	s, err := OutputWith(map[string]string{
		"MAGE_FOO": "right",
	}, "echo", "$MAGE_FOO")
	if err != nil {
		t.Fatal(err)
	}
	if s != "right" {
		t.Fatalf(`Expected "right" but got %q`, s)
	}
}

func TestEscapeExpand(t *testing.T) {
	s, err := OutputWith(map[string]string{
		"MAGE_BAR": "bar",
	}, os.Args[0], "-printArgs", "foo${MAGE_BAR}baz", Escape("foo${MAGE_BAR}baz"), `foo\$${MAGE_BAR}\\baz`)
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foobarbaz foo${MAGE_BAR}baz foo$bar\\baz]"
	if s != expected {
		t.Fatalf(`Expected %q but got %q`, expected, s)
	}
}
