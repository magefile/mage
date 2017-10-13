package mage

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/magefile/mage/mg"
)

func TestMain(m *testing.M) {
	os.Exit(testmain(m))
}

func testmain(m *testing.M) int {
	// ensure we write our temporary binaries to a directory that we'll delete
	// after running tests.
	dir := "./testing"
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Setenv(mg.CacheEnv, abs); err != nil {
		log.Fatal(err)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	return m.Run()
}

func TestGoRun(t *testing.T) {
	c := exec.Command("go", "run", "main.go")
	c.Dir = "./testdata"
	c.Env = os.Environ()
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

func TestVerbose(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"testverbose"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := ""
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
	stderr.Reset()
	stdout.Reset()
	inv.Verbose = true
	code = Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}

	actual = stderr.String()
	expected = "hi!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestVerboseEnv(t *testing.T) {
	os.Setenv("MAGE_VERBOSE", "true")

	stdout := &bytes.Buffer{}
	inv, _, _, err := Parse(stdout, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := true

	if inv.Verbose != true {
		t.Fatalf("expected %t, but got %t", expected, inv.Verbose)
	}

	os.Unsetenv("MAGE_VERBOSE")
}

func TestList(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: stdout,
		Stderr: ioutil.Discard,
		List:   true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := `
Targets:
  panics                Function that panics.
  panicsErr             Error function that panics.
  returnsError*         Synopsis for returns error.
  returnsNonNilError    Returns a non-nil error.
  returnsVoid           
  testVerbose           

* default target
`[1:]
	if actual != expected {
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actual)
	}
}

func TestNoArgNoDefaultList(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "testdata/no_default",
		Stdout: stdout,
		Stderr: stderr,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	if err := stderr.String(); err != "" {
		t.Errorf("unexpected stderr output:\n%s", err)
	}
	actual := stdout.String()
	expected := `
Targets:
  bazBuz    Prints out 'BazBuz'.
  fooBar    Prints out 'FooBar'.
`[1:]
	if actual != expected {
		t.Fatalf("expected:\n%q\n\ngot:\n%q", expected, actual)
	}
}

func TestTargetError(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"returnsnonnilerror"},
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Error: bang!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestTargetPanics(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"panics"},
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Error: boom!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestPanicsErr(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"panicserr"},
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Error: kaboom!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

// ensure we include the hash of the mainfile template in determining the
// executable name to run, so we automatically create a new exe if the template
// changes.
func TestHashTemplate(t *testing.T) {
	templ := tpl
	defer func() { tpl = templ }()
	name, err := ExeName([]string{"./testdata/func.go", "./testdata/command.go"})
	if err != nil {
		t.Fatal(err)
	}
	tpl = "some other template"
	changed, err := ExeName([]string{"./testdata/func.go", "./testdata/command.go"})
	if changed == name {
		t.Fatal("expected executable name to chage if template changed")
	}
}

// Test if the -keep flag does keep the mainfile around after running
func TestKeepFlag(t *testing.T) {
	buildFile := fmt.Sprintf("./testdata/keep_flag/%s", mainfile)
	os.Remove(buildFile)
	defer os.Remove(buildFile)
	inv := Invocation{
		Dir:    "./testdata/keep_flag",
		Stdout: ioutil.Discard,
		Stderr: ioutil.Discard,
		Args:   []string{"noop"},
		Keep:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected code 0, but got %v", code)
	}

	if _, err := os.Stat(buildFile); err != nil {
		t.Fatalf("expected file %q to exist but got err, %v", buildFile, err)
	}
}

func TestStopMultipleTargets(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"panicserr", "testVerbose"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Fatalf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Error: args after the target (panicserr) are not allowed: testVerbose\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}

}

func TestParse(t *testing.T) {
	buf := &bytes.Buffer{}
	inv, init, showVer, err := Parse(buf, []string{"-v", "build"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if init {
		t.Fatal("init should be false but was true")
	}
	if showVer {
		t.Fatal("showVersion should be false but was true")
	}
	if len(inv.Args) != 1 && inv.Args[0] != "build" {
		t.Fatalf("expected args to be %q but got %q", []string{"build"}, inv.Args)
	}
	if s := buf.String(); s != "" {
		t.Fatalf("expected no stdout output but got %q", s)
	}

}

func TestParseHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	_, _, _, err := Parse(buf, []string{"-h"})
	if err != flag.ErrHelp {
		t.Fatal("unexpected error", err)
	}
	buf2 := &bytes.Buffer{}
	_, _, _, err = Parse(buf2, []string{"--help"})
	if err != flag.ErrHelp {
		t.Fatal("unexpected error", err)
	}
	s := buf.String()
	s2 := buf2.String()
	if s != s2 {
		t.Fatalf("expected -h and --help to produce same output, but got different.\n\n-h:\n%s\n\n--help:\n%s", s, s2)
	}

}
