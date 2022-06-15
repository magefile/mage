package sh

import (
	"bytes"
	"fmt"
	"os"
	"strings"
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

func Test_When_execution_fails_quoted_command_without_arguments_is_formatted_without_trailing_whitespace(t *testing.T) {
	t.Parallel()

	command := "false"

	ran, err := Exec(nil, nil, nil, command)
	if err == nil {
		t.Fatalf("Expected runner to return error")
	}

	if !ran {
		t.Fatalf("Expected command %q to run", command)
	}

	expectedErrorFragment := fmt.Sprintf("running %q", command)

	if !strings.Contains(err.Error(), expectedErrorFragment) {
		t.Fatalf("Expected error message to contain %q, got %q", expectedErrorFragment, err.Error())
	}
}

func Test_When_command_without_arguments_is_not_found_it_is_formatted_without_trailing_whitespace(t *testing.T) {
	t.Parallel()

	command := "nonexistentcommand"

	ran, err := Exec(nil, nil, nil, command)
	if err == nil {
		t.Fatalf("Expected runner to return error")
	}

	if ran {
		t.Fatalf("Expected command %q to not run", command)
	}

	expectedErrorFragment := fmt.Sprintf("run %q", command)

	if !strings.Contains(err.Error(), expectedErrorFragment) {
		t.Fatalf("Expected error message to contain %q, got %q", expectedErrorFragment, err.Error())
	}
}
