package mage

import (
	"bytes"
	"testing"
)

func TestArgs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"status", "say", "hi", "bob", "count", "5", "status", "wait", "5ms", "cough", "false", "doubleIt", "3.1"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log(stderr.String())
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected := `status
saying hi bob
01234
status
waiting 5ms
not coughing
3.1 * 2 = 6.2
`
	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestBadIntArg(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"count", "abc123"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "can't convert argument \"abc123\" to int\n"

	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestBadBoolArg(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"cough", "abc123"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "can't convert argument \"abc123\" to bool\n"

	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestBadDurationArg(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"wait", "abc123"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "can't convert argument \"abc123\" to time.Duration\n"

	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestBadFloat64Arg(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"doubleIt", "abc123"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "can't convert argument \"abc123\" to float64\n"

	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestMissingArgs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "hi"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "not enough arguments for target \"Say\", expected 2, got 1\n"

	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestDocs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Help:   true,
		Args:   []string{"say"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := `Say says something. It's pretty cool. I think you should try it.

Usage:

	mage say <msg> <name>

Aliases: speak

`
	if actual != expected {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"greet", "World", "-greeting=Hi"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "Hi, World!\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsOmitted(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"greet", "World"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "Hello, World!\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsAllTypes(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args: []string{
			"add", "10", "-b=5",
			"scale", "2.0", "-factor=3.0",
			"run", "-verbose=true",
			"delay", "1s", "-extra=500ms",
		},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "15\n6.0\nrunning verbose\ndelay 1s + 500ms\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsAllOmitted(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args: []string{
			"add", "10",
			"run",
			"allOptional",
		},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "10\nrunning quiet\na=<nil>\nb=<nil>\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsBadInt(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"add", "10", "-b=notanumber"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "can't convert option \"b\" value \"notanumber\" to int\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsBadBool(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"run", "-verbose=notabool"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "can't convert option \"verbose\" value \"notabool\" to bool\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsBadFloat64(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"scale", "1.0", "-factor=notafloat"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "can't convert option \"factor\" value \"notafloat\" to float64\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsBadDuration(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"delay", "1s", "-extra=notaduration"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "can't convert option \"extra\" value \"notaduration\" to time.Duration\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestUnknownOptionalArg(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"greet", "World", "-unknown=foo"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "unknown option \"unknown\" for target \"Greet\"\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsMissingEquals(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"greet", "World", "-greeting"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "invalid option \"-greeting\" for target \"Greet\", expected -name=value format\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsMissingRequired(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"add"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 2, but got %v", code)
	}
	actual := stderr.String()
	expected2 := "not enough arguments for target \"Add\", expected 1, got 0\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsDocs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Help:   true,
		Args:   []string{"greet"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := `Greet greets someone with an optional greeting.

Usage:

	mage greet <name> [-greeting=<string>]

`
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsFlagDocs(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Help:   true,
		Args:   []string{"flagdocs"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := `FlagDocs tests that docs on flags are properly displayed when you run mage -h FlagDocs.

Usage:

	mage flagdocs <name> [<flags>]

Flags:

	-greeting=<string>  the message to append to the name
	-repeat=<int>       the number of times to repeat

`
	if actual != expected {
		t.Fatalf("output is not expected:\ngot:  %q\nwant: %q", actual, expected)
	}
}

func TestOptionalArgsSingleFlagDoc(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Help:   true,
		Args:   []string{"singleflagdoc"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := `SingleFlagDoc tests that a single documented flag shows the Flags section.

Usage:

	mage singleflagdoc <name> [-greeting=<string>]

Flags:

	-greeting=<string>  the greeting to use

`
	if actual != expected {
		t.Fatalf("output is not expected:\ngot:  %q\nwant: %q", actual, expected)
	}
}

func TestOptionalArgsCaseInsensitive(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"greet", "World", "-Greeting=Hey"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "Hey, World!\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalBoolBareFlag(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"run", "-verbose"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "running verbose\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalBoolBareFlagChained(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "hello", "-cap", "announce", "world"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "HELLO\nAnnouncement: world\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalBoolBareFlagMultiple(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "hi", "-cap", "-count=3"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "HI\nHI\nHI\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalAllOptionalWithValues(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"allOptional", "-a=hello", "-b=42"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "a=hello\nb=42\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalMixedRequiredAndOptional(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"mixed", "World", "2", "-greeting=Hey"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "Hey, World!\nHey, World!\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsChainedTargets(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "oops", "-count=2", "-cap=true", "announce", "done"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "OOPS\nOOPS\nAnnouncement: done\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsChainedNoOptionals(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "hello", "announce", "world"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "hello\nAnnouncement: world\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsChainedMultipleOptionalTargets(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"say", "hi", "-cap=true", "say", "bye", "-count=3"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "HI\nbye\nbye\nbye\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestOptionalArgsChainedAllOptionalThenRequired(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/optargs",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"allOptional", "-a=test", "announce", "news"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected2 := "a=test\nb=<nil>\nAnnouncement: news\n"
	if actual != expected2 {
		t.Fatalf("output is not expected:\n%q", actual)
	}
}

func TestMgF(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/args",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"HasDep"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log("stderr:", stderr)
		t.Log("stdout:", stdout)
		t.Fatalf("expected code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "saying hi Susan\n"
	if actual != expected {
		t.Fatalf("output is not expected: %q", actual)
	}
}
