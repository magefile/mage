package mage

import (
	"bytes"
	"debug/macho"
	"debug/pe"
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/magefile/mage/internal"
	"github.com/magefile/mage/mg"
)

const testExeEnv = "MAGE_TEST_STRING"

func TestMain(m *testing.M) {
	if s := os.Getenv(testExeEnv); s != "" {
		fmt.Fprint(os.Stdout, s)
		os.Exit(0)
	}
	os.Exit(testmain(m))
}

func testmain(m *testing.M) int {
	// ensure we write our temporary binaries to a directory that we'll delete
	// after running tests.
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := os.Setenv(mg.CacheEnv, dir); err != nil {
		log.Fatal(err)
	}
	if err := os.Unsetenv(mg.VerboseEnv); err != nil {
		log.Fatal(err)
	}
	if err := os.Unsetenv(mg.DebugEnv); err != nil {
		log.Fatal(err)
	}
	if err := os.Unsetenv(mg.IgnoreDefaultEnv); err != nil {
		log.Fatal(err)
	}
	return m.Run()
}

func TestTransitiveDepCache(t *testing.T) {
	cache, err := internal.OutputDebug("go", "env", "gocache")
	if err != nil {
		t.Fatal(err)
	}
	if cache == "" {
		t.Skip("skipping gocache tests on go version without cache")
	}
	// Test that if we change a transitive dep, that we recompile
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Stderr: stderr,
		Stdout: stdout,
		Dir:    "testdata/transitiveDeps",
		Args:   []string{"Run"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("got code %v, err: %s", code, stderr)
	}
	expected := "woof\n"
	if actual := stdout.String(); actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
	// ok, so baseline, the generated and cached binary should do "woof"
	// now change out the transitive dependency that does the output
	// so that it produces different output.
	if err := os.Rename("testdata/transitiveDeps/dep/dog.go", "testdata/transitiveDeps/dep/dog.notgo"); err != nil {
		t.Fatal(err)
	}
	defer os.Rename("testdata/transitiveDeps/dep/dog.notgo", "testdata/transitiveDeps/dep/dog.go")
	if err := os.Rename("testdata/transitiveDeps/dep/cat.notgo", "testdata/transitiveDeps/dep/cat.go"); err != nil {
		t.Fatal(err)
	}
	defer os.Rename("testdata/transitiveDeps/dep/cat.go", "testdata/transitiveDeps/dep/cat.notgo")
	stderr.Reset()
	stdout.Reset()
	code = Invoke(inv)
	if code != 0 {
		t.Fatalf("got code %v, err: %s", code, stderr)
	}
	expected = "meow\n"
	if actual := stdout.String(); actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
}

func TestListMagefilesMain(t *testing.T) {
	buf := &bytes.Buffer{}
	files, err := Magefiles("testdata/mixed_main_files", "", "", "go", buf, false)
	if err != nil {
		t.Errorf("error from magefile list: %v: %s", err, buf)
	}
	expected := []string{"testdata/mixed_main_files/mage_helpers.go", "testdata/mixed_main_files/magefile.go"}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestListMagefilesIgnoresGOOS(t *testing.T) {
	buf := &bytes.Buffer{}
	if runtime.GOOS == "windows" {
		os.Setenv("GOOS", "linux")
	} else {
		os.Setenv("GOOS", "windows")
	}
	defer os.Setenv("GOOS", runtime.GOOS)
	files, err := Magefiles("testdata/goos_magefiles", "", "", "go", buf, false)
	if err != nil {
		t.Errorf("error from magefile list: %v: %s", err, buf)
	}
	var expected []string
	if runtime.GOOS == "windows" {
		expected = []string{"testdata/goos_magefiles/magefile_windows.go"}
	} else {
		expected = []string{"testdata/goos_magefiles/magefile_nonwindows.go"}
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestListMagefilesIgnoresRespectsGOOSArg(t *testing.T) {
	buf := &bytes.Buffer{}
	var goos string
	if runtime.GOOS == "windows" {
		goos = "linux"
	} else {
		goos = "windows"
	}
	files, err := Magefiles("testdata/goos_magefiles", goos, "", "go", buf, false)
	if err != nil {
		t.Errorf("error from magefile list: %v: %s", err, buf)
	}
	var expected []string
	if goos == "windows" {
		expected = []string{"testdata/goos_magefiles/magefile_windows.go"}
	} else {
		expected = []string{"testdata/goos_magefiles/magefile_nonwindows.go"}
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestCompileDiffGoosGoarch(t *testing.T) {
	target, err := ioutil.TempDir("./testdata", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(target)

	// intentionally choose an arch and os to build that are not our current one.

	goos := "windows"
	if runtime.GOOS == "windows" {
		goos = "darwin"
	}
	goarch := "amd64"
	if runtime.GOARCH == "amd64" {
		goarch = "386"
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Stderr: stderr,
		Stdout: stdout,
		Debug:  true,
		Dir:    "testdata",
		// this is relative to the Dir above
		CompileOut: filepath.Join(".", filepath.Base(target), "output"),
		GOOS:       goos,
		GOARCH:     goarch,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("got code %v, err: %s", code, stderr)
	}
	os, arch, err := fileData(filepath.Join(target, "output"))
	if err != nil {
		t.Fatal(err)
	}
	if goos == "windows" {
		if os != winExe {
			t.Error("ran with GOOS=windows but did not produce a windows exe")
		}
	} else {
		if os != macExe {
			t.Error("ran with GOOS=darwin but did not a mac exe")
		}
	}
	if goarch == "amd64" {
		if arch != arch64 {
			t.Error("ran with GOARCH=amd64 but did not produce a 64 bit exe")
		}
	} else {
		if arch != arch32 {
			t.Error("rand with GOARCH=386 but did not produce a 32 bit exe")
		}
	}
}

func TestListMagefilesLib(t *testing.T) {
	buf := &bytes.Buffer{}
	files, err := Magefiles("testdata/mixed_lib_files", "", "", "go", buf, false)
	if err != nil {
		t.Errorf("error from magefile list: %v: %s", err, buf)
	}
	expected := []string{"testdata/mixed_lib_files/mage_helpers.go", "testdata/mixed_lib_files/magefile.go"}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestMixedMageImports(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/mixed_lib_files",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}
	expected := "Targets:\n  build    \n"
	actual := stdout.String()
	if actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
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
	os.Setenv("MAGEFILE_VERBOSE", "true")
	defer os.Unsetenv("MAGEFILE_VERBOSE")
	stdout := &bytes.Buffer{}
	inv, _, err := Parse(ioutil.Discard, stdout, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := true

	if inv.Verbose != true {
		t.Fatalf("expected %t, but got %t ", expected, inv.Verbose)
	}
}
func TestVerboseFalseEnv(t *testing.T) {
	os.Setenv("MAGEFILE_VERBOSE", "0")
	defer os.Unsetenv("MAGEFILE_VERBOSE")
	stdout := &bytes.Buffer{}
	code := ParseAndRun(ioutil.Discard, stdout, nil, []string{"-d", "testdata", "testverbose"})
	if code != 0 {
		t.Fatal("unexpected code", code)
	}

	if stdout.String() != "" {
		t.Fatalf("expected no output, but got %s", stdout.String())
	}
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

func TestIgnoreDefault(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/list",
		Stdout: stdout,
		Stderr: stderr,
	}
	defer os.Unsetenv(mg.IgnoreDefaultEnv)
	if err := os.Setenv(mg.IgnoreDefaultEnv, "1"); err != nil {
		t.Fatal(err)
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
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
	templ := mageMainfileTplString
	defer func() { mageMainfileTplString = templ }()
	name, err := ExeName("go", mg.CacheDir(), []string{"testdata/func.go", "testdata/command.go"})
	if err != nil {
		t.Fatal(err)
	}
	mageMainfileTplString = "some other template"
	changed, err := ExeName("go", mg.CacheDir(), []string{"testdata/func.go", "testdata/command.go"})
	if err != nil {
		t.Fatal(err)
	}
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
	inv, cmd, err := Parse(ioutil.Discard, buf, []string{"-v", "-debug", "-gocmd=foo", "-d", "dir", "build", "deploy"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if cmd == Init {
		t.Error("init should be false but was true")
	}
	if cmd == Version {
		t.Error("showVersion should be false but was true")
	}
	if inv.Debug != true {
		t.Error("debug should be true")
	}
	if inv.Dir != "dir" {
		t.Errorf("Expected dir to be \"dir\" but was %q", inv.Dir)
	}
	if inv.GoCmd != "foo" {
		t.Errorf("Expected gocmd to be \"foo\" but was %q", inv.GoCmd)
	}
	expected := []string{"build", "deploy"}
	if !reflect.DeepEqual(inv.Args, expected) {
		t.Fatalf("expected args to be %q but got %q", expected, inv.Args)
	}
	if s := buf.String(); s != "" {
		t.Fatalf("expected no stdout output but got %q", s)
	}

}

func TestSetDir(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Invoke(Invocation{
		Dir:    "testdata/setdir",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"TestCurrentDir"},
	})
	if code != 0 {
		t.Errorf("expected code 0, but got %d. Stdout:\n%s\nStderr:\n%s", code, stdout, stderr)
	}
	expected := "setdir.go\n"
	if out := stdout.String(); out != expected {
		t.Fatalf("expected list of files to be %q, but was %q", expected, out)
	}
}

// Test the timeout option
func TestTimeout(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:     "testdata/context",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"timeout"},
		Timeout: time.Duration(100 * time.Millisecond),
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v, stderr: %q, stdout: %q", code, stderr, stdout)
	}
	actual := stderr.String()
	expected := "Error: context deadline exceeded\n"

	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}
func TestParseHelp(t *testing.T) {
	buf := &bytes.Buffer{}
	_, _, err := Parse(ioutil.Discard, buf, []string{"-h"})
	if err != flag.ErrHelp {
		t.Fatal("unexpected error", err)
	}
	buf2 := &bytes.Buffer{}
	_, _, err = Parse(ioutil.Discard, buf2, []string{"--help"})
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
		Stderr: ioutil.Discard,
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
		Stderr: ioutil.Discard,
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
		Stderr: ioutil.Discard,
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
	stderr := &bytes.Buffer{}
	debug.SetOutput(stderr)
	inv := Invocation{
		Dir:    "testdata/alias",
		Stdout: stdout,
		Stderr: ioutil.Discard,
		Args:   []string{"status"},
		Debug:  true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v\noutput:\n%s\nstderr:\n%s", code, stdout, stderr)
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
	log.SetOutput(ioutil.Discard)
	inv := Invocation{
		Dir:    "./testdata/invalid_alias",
		Stdout: ioutil.Discard,
		Stderr: stderr,
		Args:   []string{"co"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Errorf("expected to exit with code 1, but got %v", code)
	}
	actual := stderr.String()
	expected := "Unknown target specified: co\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestRunCompiledPrintsError(t *testing.T) {
	stderr := &bytes.Buffer{}
	logger := log.New(stderr, "", 0)
	code := RunCompiled(Invocation{}, "thiswon'texist", logger)
	if code != 1 {
		t.Errorf("expected code 1 but got %v", code)
	}

	if strings.TrimSpace(stderr.String()) == "" {
		t.Fatal("expected to get output to stderr when a run fails, but got nothing.")
	}
}

func TestCompiledFlags(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/compiled"
	compileDir, err := ioutil.TempDir(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(compileDir, "mage_out")
	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + name[len(dir)-1:]
	defer os.RemoveAll(compileDir)
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(stdout, stderr *bytes.Buffer, filename string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.Command(filename, args...)
		cmd.Env = os.Environ()
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %v\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	// get help to target with flag -h target
	if err := run(stdout, stderr, name, "-h", "deploy"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(stdout.String())
	want := filepath.Base(name) + " deploy:\n\nThis is the synopsis for Deploy. This part shouldn't show up."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// run target with verbose flag -v
	if err := run(stdout, stderr, name, "-v", "testverbose"); err != nil {
		t.Fatal(err)
	}
	got = stderr.String()
	want = "hi!"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	// pass list flag -l
	if err := run(stdout, stderr, name, "-l"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "This is the synopsis for Deploy"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	want = "This is very verbose"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	// pass flag -t 1ms
	err = run(stdout, stderr, name, "-t", "1ms", "sleep")
	if err == nil {
		t.Fatalf("expected an error because of timeout")
	}
	got = stdout.String()
	want = "context deadline exceeded"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}
}

func TestCompiledEnvironmentVars(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/compiled"
	compileDir, err := ioutil.TempDir(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(compileDir, "mage_out")
	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + name[len(dir)-1:]
	defer os.RemoveAll(compileDir)
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(stdout, stderr *bytes.Buffer, filename string, envval string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.Command(filename, args...)
		cmd.Env = []string{envval}
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %v\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	if err := run(stdout, stderr, name, "MAGEFILE_HELP=1", "deploy"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(stdout.String())
	want := filepath.Base(name) + " deploy:\n\nThis is the synopsis for Deploy. This part shouldn't show up."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if err := run(stdout, stderr, name, mg.VerboseEnv+"=1", "testverbose"); err != nil {
		t.Fatal(err)
	}
	got = stderr.String()
	want = "hi!"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "MAGEFILE_LIST=1"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "This is the synopsis for Deploy"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	want = "This is very verbose"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, mg.IgnoreDefaultEnv+"=1"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "Compiled package description."
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	err = run(stdout, stderr, name, "MAGEFILE_TIMEOUT=1ms", "sleep")
	if err == nil {
		t.Fatalf("expected an error because of timeout")
	}
	got = stdout.String()
	want = "context deadline exceeded"
	if strings.Contains(got, want) == false {
		t.Errorf("got %q, does not contain %q", got, want)
	}
}

func TestClean(t *testing.T) {
	if err := os.RemoveAll(mg.CacheDir()); err != nil {
		t.Error("error removing cache dir:", err)
	}
	code := ParseAndRun(ioutil.Discard, ioutil.Discard, &bytes.Buffer{}, []string{"-clean"})
	if code != 0 {
		t.Errorf("expected 0, but got %v", code)
	}

	TestAlias(t) // make sure we've got something in the CACHE_DIR
	files, err := ioutil.ReadDir(mg.CacheDir())
	if err != nil {
		t.Error("issue reading file:", err)
	}

	if len(files) < 1 {
		t.Error("Need at least 1 cached binaries to test --clean")
	}

	_, cmd, err := Parse(ioutil.Discard, ioutil.Discard, []string{"-clean"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != Clean {
		t.Errorf("Expected 'clean' command but got %v", cmd)
	}
	buf := &bytes.Buffer{}
	code = ParseAndRun(ioutil.Discard, buf, &bytes.Buffer{}, []string{"-clean"})
	if code != 0 {
		t.Fatalf("expected 0, but got %v: %s", code, buf)
	}

	infos, err := ioutil.ReadDir(mg.CacheDir())
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, i := range infos {
		if !i.IsDir() {
			names = append(names, i.Name())
		}
	}

	if len(names) != 0 {
		t.Errorf("expected '-clean' to remove files from CACHE_DIR, but still have %v", names)
	}
}

func TestGoCmd(t *testing.T) {
	textOutput := "TestGoCmd"
	defer os.Unsetenv(testExeEnv)
	if err := os.Setenv(testExeEnv, textOutput); err != nil {
		t.Fatal(err)
	}

	// fake out the compiled file, since the code checks for it.
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	dir := filepath.Dir(name)
	defer os.Remove(name)
	f.Close()

	buf := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if err := Compile("", "", dir, os.Args[0], name, []string{}, false, stderr, buf); err != nil {
		t.Log("stderr: ", stderr.String())
		t.Fatal(err)
	}
	if buf.String() != textOutput {
		t.Fatalf("We didn't run the custom go cmd. Expected output %q, but got %q", textOutput, buf)
	}
}

var runtimeVer = regexp.MustCompile(`go1\.([0-9]+)`)

func TestGoModules(t *testing.T) {
	matches := runtimeVer.FindStringSubmatch(runtime.Version())
	if len(matches) < 2 || minorVer(t, matches[1]) < 11 {
		t.Skipf("Skipping Go modules test because go version %q is less than go1.11", runtime.Version())
	}
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	err = ioutil.WriteFile(filepath.Join(dir, "magefile.go"), []byte(`//+build mage

package main

import "golang.org/x/text/unicode/norm"

func Test() {
	print("unicode version: " + norm.Version)
}
`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Command("go", "mod", "init", "app")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running go mod init: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	stderr.Reset()
	stdout.Reset()
	code := Invoke(Invocation{
		Dir:    dir,
		Stderr: stderr,
		Stdout: stdout,
	})
	if code != 0 {
		t.Fatalf("exited with code %d. \nStdout: %s\nStderr: %s", code, stdout, stderr)
	}
	expected := `
Targets:
  test    
`[1:]
	if output := stdout.String(); output != expected {
		t.Fatalf("expected output %q, but got %q", expected, output)
	}
}

func minorVer(t *testing.T, v string) int {
	a, err := strconv.Atoi(v)
	if err != nil {
		t.Fatal("unexpected non-numeric version", v)
	}
	return a
}

func TestNamespaceDep(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/namespaces",
		Stderr: stderr,
		Stdout: stdout,
		Args:   []string{"TestNamespaceDep"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected 0, but got %v, stderr:\n%s", code, stderr)
	}
	expected := "hi!\n"
	if stdout.String() != expected {
		t.Fatalf("expected %q, but got %q", expected, stdout.String())
	}
}

func TestNamespace(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/namespaces",
		Stderr: ioutil.Discard,
		Stdout: stdout,
		Args:   []string{"ns:error"},
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected 0, but got %v", code)
	}
	expected := "hi!\n"
	if stdout.String() != expected {
		t.Fatalf("expected %q, but got %q", expected, stdout.String())
	}
}

func TestNamespaceDefault(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/namespaces",
		Stderr: ioutil.Discard,
		Stdout: stdout,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected 0, but got %v", code)
	}
	expected := "hi!\n"
	if stdout.String() != expected {
		t.Fatalf("expected %q, but got %q", expected, stdout.String())
	}
}

func TestAliasToImport(t *testing.T) {

}

/// This code liberally borrowed from https://github.com/rsc/goversion/blob/master/version/exe.go

type exeType int
type archSize int

const (
	winExe exeType = iota
	macExe

	arch32 archSize = iota
	arch64
)

// fileData tells us if the given file is mac or windows and if they're 32bit or
// 64 bit.  Other exe versions are not supported.
func fileData(file string) (exeType, archSize, error) {
	f, err := os.Open(file)
	if err != nil {
		return -1, -1, err
	}
	defer f.Close()
	data := make([]byte, 16)
	if _, err := io.ReadFull(f, data); err != nil {
		return -1, -1, err
	}
	if bytes.HasPrefix(data, []byte("MZ")) {
		// hello windows exe!
		e, err := pe.NewFile(f)
		if err != nil {
			return -1, -1, err
		}
		if e.Machine == pe.IMAGE_FILE_MACHINE_AMD64 {
			return winExe, arch64, nil
		}
		return winExe, arch32, nil
	}

	if bytes.HasPrefix(data, []byte("\xFE\xED\xFA")) || bytes.HasPrefix(data[1:], []byte("\xFA\xED\xFE")) {
		// hello mac exe!
		fe, err := macho.NewFile(f)
		if err != nil {
			return -1, -1, err
		}
		if fe.Cpu&0x01000000 != 0 {
			return macExe, arch64, nil
		}
		return macExe, arch32, nil
	}
	return -1, -1, fmt.Errorf("unrecognized executable format")
}
