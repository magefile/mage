package sh

import (
	"bytes"
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

func TestAutoEnvExpand(t *testing.T) {
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

func TestAutoGlobExpand(t *testing.T) {
	// these deliberately specify a set of files that should be stable over the longer term
	// (i.e. avoiding the names of source code files)
	t.Run("* ? and [] glob", func(t *testing.T) {
		s, err := Output("ls", "../R*.md", "../go.??[dm]", "../*/config.toml")
		if err != nil {
			t.Fatal(err)
		}
		if s != "../go.mod\n../go.sum\n../README.md\n../site/config.toml" {
			t.Errorf(`Expected "../go.mod\n../go.sum\n../README.md\n../site/config.toml" but got %q`, s)
		}
	})
	t.Run("glob syntax error", func(t *testing.T) {
		_, err := Output("ls", "../go.\\")
		if err == nil {
			t.Fatalf("expected error, but got nil")
		} else if !strings.Contains(err.Error(), `failed to run "ls ../go.\:`) {
			t.Errorf("Actual error was %q", err.Error())
		}
	})
}
