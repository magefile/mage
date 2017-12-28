package mage

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/magefile/mage/build"
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
		if os.IsExist(err) {
			os.RemoveAll(dir)
		} else {
			log.Fatal(err)
		}
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
	expected = "Running target: TestVerbose\nhi!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestVerboseEnv(t *testing.T) {
	os.Setenv("MAGE_VERBOSE", "true")

	stdout := &bytes.Buffer{}
	inv, _, err := Parse(stdout, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := true

	if inv.Verbose != true {
		t.Fatalf("expected %t, but got %t ", expected, inv.Verbose)
	}

	os.Unsetenv("MAGE_VERBOSE")
}

func TestList(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/list",
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

func TestStdinCopy(t *testing.T) {
	stdout := &bytes.Buffer{}
	stdin := strings.NewReader("hi!")
	inv := Invocation{
		Dir:    "./testdata",
		Stderr: ioutil.Discard,
		Stdout: stdout,
		Stdin:  stdin,
		Args:   []string{"CopyStdin"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "hi!"
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
	w := tLogWriter{t}

	inv := Invocation{
		Dir:    "./testdata/keep_flag",
		Stdout: w,
		Stderr: w,
		List:   true,
		Keep:   true,
		Force:  true, // need force so we always regenerate
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected code 0, but got %v", code)
	}

	if _, err := os.Stat(buildFile); err != nil {
		t.Fatalf("expected file %q to exist but got err, %v", buildFile, err)
	}
}

type tLogWriter struct {
	*testing.T
}

func (t tLogWriter) Write(b []byte) (n int, err error) {
	t.Log(string(b))
	return len(b), nil
}

// Test if generated mainfile references anything other than the stdlib
func TestOnlyStdLib(t *testing.T) {
	buildFile := fmt.Sprintf("./testdata/onlyStdLib/%s", mainfile)
	os.Remove(buildFile)
	defer os.Remove(buildFile)

	w := tLogWriter{t}

	inv := Invocation{
		Dir:     "./testdata/onlyStdLib",
		Stdout:  w,
		Stderr:  w,
		List:    true,
		Keep:    true,
		Force:   true, // need force so we always regenerate
		Verbose: true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected code 0, but got %v", code)
	}

	if _, err := os.Stat(buildFile); err != nil {
		t.Fatalf("expected file %q to exist but got err, %v", buildFile, err)
	}

	fset := &token.FileSet{}
	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, buildFile, nil, parser.ImportsOnly)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the imports from the file's AST.
	for _, s := range f.Imports {
		// the path value comes in as a quoted string, i.e. literally \"context\"
		path := strings.Trim(s.Path.Value, "\"")
		pkg, err := build.Default.Import(path, "./testdata/keep_flag", build.FindOnly)
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.HasPrefix(pkg.Dir, build.Default.GOROOT) {
			t.Errorf("import of non-stdlib package: %s", s.Path.Value)
		}
	}
}

func TestMultipleTargets(t *testing.T) {
	var stderr, stdout bytes.Buffer
	inv := Invocation{
		Dir:     "./testdata",
		Stdout:  &stdout,
		Stderr:  &stderr,
		Args:    []string{"TestVerbose", "ReturnsNilError"},
		Verbose: true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected 0, but got %v", code)
	}
	actual := stderr.String()
	expected := "Running target: TestVerbose\nhi!\nRunning target: ReturnsNilError\n"
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
	actual = stdout.String()
	expected = "stuff\n"
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
}

func TestFirstTargetFails(t *testing.T) {
	var stderr, stdout bytes.Buffer
	inv := Invocation{
		Dir:     "./testdata",
		Stdout:  &stdout,
		Stderr:  &stderr,
		Args:    []string{"ReturnsNonNilError", "ReturnsNilError"},
		Verbose: true,
	}
	code := Invoke(inv)
	if code != 1 {
		t.Errorf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Running target: ReturnsNonNilError\nError: bang!\n"
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
	actual = stdout.String()
	expected = ""
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
}

func TestBadSecondTargets(t *testing.T) {
	var stderr, stdout bytes.Buffer
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: &stdout,
		Stderr: &stderr,
		Args:   []string{"TestVerbose", "NotGonnaWork"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Errorf("expected 0, but got %v", code)
	}
	actual := stderr.String()
	expected := "Unknown target specified: NotGonnaWork\n"
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
	actual = stdout.String()
	expected = ""
	if actual != expected {
		t.Errorf("expected %q, but got %q", expected, actual)
	}
}

func TestParse(t *testing.T) {
	buf := &bytes.Buffer{}
	inv, cmd, err := Parse(buf, []string{"-v", "build"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if cmd == Init {
		t.Fatal("init should be false but was true")
	}
	if cmd == Version {
		t.Fatal("showVersion should be false but was true")
	}
	if len(inv.Args) != 1 && inv.Args[0] != "build" {
		t.Fatalf("expected args to be %q but got %q", []string{"build"}, inv.Args)
	}
	if s := buf.String(); s != "" {
		t.Fatalf("expected no stdout output but got %q", s)
	}

}

// Test the timeout option
func TestTimeout(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:     "./testdata/context",
		Stdout:  ioutil.Discard,
		Stderr:  stderr,
		Args:    []string{"timeout"},
		Timeout: time.Duration(100 * time.Millisecond),
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Error: context deadline exceeded\n"

	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}
func TestParseHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	_, _, err := Parse(buf, []string{"-h"})
	if err != flag.ErrHelp {
		t.Fatal("unexpected error", err)
	}
	buf2 := &bytes.Buffer{}
	_, _, err = Parse(buf2, []string{"--help"})
	if err != flag.ErrHelp {
		t.Fatal("unexpected error", err)
	}
	s := buf.String()
	s2 := buf2.String()
	if s != s2 {
		t.Fatalf("expected -h and --help to produce same output, but got different.\n\n-h:\n%s\n\n--help:\n%s", s, s2)
	}
}

func TestHelpTarget(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata",
		Stdout: stdout,
		Stderr: os.Stderr,
		Args:   []string{"panics"},
		Help:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "mage panics:\n\nFunction that panics.\n\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestHelpAlias(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: os.Stderr,
		Args:   []string{"status"},
		Help:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "mage status:\n\nPrints status.\n\nAliases: st, stat\n\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
	inv = Invocation{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: os.Stderr,
		Args:   []string{"checkout"},
		Help:   true,
	}
	stdout.Reset()
	code = Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual = stdout.String()
	expected = "mage checkout:\n\nAliases: co\n\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestAlias(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: ioutil.Discard,
		Args:   []string{"status"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "alias!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
	stdout.Reset()
	inv.Args = []string{"st"}
	code = Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual = stdout.String()
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestInvalidAlias(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/invalid_alias",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"co"},
	}
	code := Invoke(inv)
	if code != 1 {
		t.Errorf("expected to exit with code 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Unknown target: \"co\"\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestClean(t *testing.T) {
	TestGoRun(t) // make sure we've got something in the CACHE_DIR
	dir := "./testing"
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(abs)
	if err != nil {
		t.Error("issue reading file:", err)
	}

	if len(files) < 1 {
		t.Error("Need at least 1 cached binaries to test --clean")
	}

	buf := &bytes.Buffer{}
	_, cmd, err := Parse(buf, []string{"-clean"})
	if cmd != Clean {
		t.Errorf("Expected 'clean' command but got %v", cmd)
	}
	code := ParseAndRun(dir, os.Stdin, os.Stderr, buf, []string{"-clean"})
	if code != 0 {
		t.Errorf("expected 0, but got %v", code)
	}

	files, err = ioutil.ReadDir(abs)
	if err != nil {
		t.Error("issue reading file:", err)
	}

	if len(files) != 0 {
		t.Errorf("expected '-clean' to remove files from CACHE_DIR, but still have %v", files)
	}
}
