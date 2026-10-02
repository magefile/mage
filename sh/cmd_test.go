package sh

import (
	"bytes"
	"errors"
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
		t.Error("ran returned as false, but should have been true")
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
		t.Error("expected ran to be true but was false.")
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
	t.Setenv("MAGE_FOOBAR", "baz")
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}

func TestCmdRanNilErr(t *testing.T) {
	if !CmdRan(nil) {
		t.Fatal("CmdRan(nil) should return true")
	}
}

func TestCmdRanNotFound(t *testing.T) {
	_, err := Exec(nil, nil, nil, "thiswontwork")
	if CmdRan(err) {
		t.Fatal("CmdRan should return false for not-found command")
	}
}

func TestExitStatusNil(t *testing.T) {
	code := ExitStatus(nil)
	if code != 0 {
		t.Fatalf("expected 0 for nil error, got %d", code)
	}
}

func TestExitStatusNonExecError(t *testing.T) {
	code := ExitStatus(errors.New("generic error"))
	if code != 1 {
		t.Fatalf("expected 1 for generic error, got %d", code)
	}
}

func TestExitStatusFromExec(t *testing.T) {
	_, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "42")
	code := ExitStatus(err)
	if code != 42 {
		t.Fatalf("expected exit status 42, got %d", code)
	}
}

func TestRunCmd(t *testing.T) {
	echoHello := RunCmd("echo", "hello")
	err := echoHello("world")
	// RunWith directs output based on verbose, so just check no error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutputWith(t *testing.T) {
	out, err := OutputWith(map[string]string{"MY_TEST_VAR": "xyz"}, os.Args[0], "-printVar", "MY_TEST_VAR")
	if err != nil {
		t.Fatal(err)
	}
	if out != "xyz" {
		t.Fatalf("expected 'xyz', got %q", out)
	}
}

func TestAutoExpandPrecedent(t *testing.T) {
	// Environment variables passed to OutputWith should take precedence
	// over any variables set in the actual environment.
	t.Setenv("MAGE_FOO", "wrong")
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

func TestExpand(t *testing.T) {
	// If "wrong" ever appears in the output, then Expand has incorrectly
	// selected one of values here that shouldn't be detected as a variable
	// name.
	replacements := map[string]string{
		"MAGE_BAR": "bar",
		"1":        "quux",
		"10":       "wrong",
		"A1":       "foo",
		"1A":       "wrong",
		"a\nb":     "newline",
		"a=b":      "equals",
		"a$b":      "dollar",
		"a b":      "space",
		"a*b":      "asterisk",
		"a.b":      "period",
		"a}b":      "wrong",
		"a{b":      "brace",
		"*":        "asterisk",
		"*a":       "wrong",
		"a":        "A",
	}
	for _, test := range []struct {
		label    string
		arg      string
		expected string // If empty, then match os.Expand.
	}{
		{label: "Braced identifier", arg: `foo${MAGE_BAR}baz`},
		// Identifier is MAGE_BARbaz, and there's no match for that.
		{label: "Unbraced identifier", arg: `foo$MAGE_BARbaz`},
		{label: "Escape function", arg: Escape(`foo${MAGE_BAR}baz`), expected: `foo${MAGE_BAR}baz`},
		{label: "Manual backslashes", arg: `foo\$${MAGE_BAR}\\baz`, expected: `foo$bar\baz`},
		{label: "Non-escape sequences", arg: `C:\tools\mage`},

		// To match os.Expand, ours should accept _any_ character
		// inside braces (except for a closing brace). We'll test
		// specifically for characters that POSIX says are valid in
		// environment-variable names (i.e., the portable character
		// set).
		// https://pubs.opengroup.org/onlinepubs/000095399/basedefs/xbd_chap06.html
		{label: "Period accepted in braces", arg: "X${a.b}X"},
		{label: "Newline accepted in braces", arg: "X${a\nb}X"},
		{label: "Dollar accepted in braces", arg: "X${a$b}X"},
		{label: "Space accepted in braces", arg: "X${a b}X"},
		{label: "Asterisk accepted in braces", arg: "X${a*b}X"},
		{label: "Left brace accepted in braces", arg: "X${a{b}X"},
		// There's no syntax for distinguishing a mid-variable right
		// brace from a terminal right brace.
		{label: "Right brace rejected in braces", arg: "X${a}b}X"},
		// Equals and nul aren't allowed in a true environment-variable
		// name, but os.Expand accepts them anyway, so we will, too.
		{label: "Equals accepted in braces", arg: "X${a=b}X"},
		{label: "Nul accepted in braces", arg: "X${a\000b}X"},

		// Bare variable names without braces. List of allowed
		// characters is just \w, plus some standalone symbols.
		{label: "Bare symbol", arg: "$*A"},
		// Without braces, $10 should expand as though it were ${1}0.
		{label: "Unbraced digits", arg: `$10`},
		{label: "First digit", arg: "$1A"},
		{label: "Later digit", arg: "$A1"},
		{label: "Bare period terminates", arg: "$a.b"},
		{label: "Bare newline terminates", arg: "$a\nb"},
		{label: "Bare space terminates", arg: "$a b"},
		{label: "Bare asterisk terminates", arg: "$a*b"},
		{label: "Bare left brace terminates", arg: "$a{b"},
		{label: "Bare right brace terminates", arg: "$a}b"},
		{label: "Bare equals terminates", arg: "$a=b"},
	} {
		t.Run(test.label, func(t *testing.T) {
			if test.expected == "" {
				// The current subtest doesn't involve any
				// backslash-escaping, so it amounts to a
				// regression test with respect to os.Expand.
				test.expected = os.Expand(test.arg, func(s string) string {
					return replacements[s]
				})
			}
			s, err := OutputWith(replacements, os.Args[0], "-printArgs", test.arg)
			if err != nil {
				t.Fatal(err)
			}
			expected := "[" + test.expected + "]"
			if s != expected {
				t.Fatalf(`Expected %q but got %q.`, expected, s)
			}
		})
	}
}
