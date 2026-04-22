package mage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/macho"
	"debug/pe"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io"
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
		_, _ = fmt.Fprint(os.Stdout, s)
		return
	}
	err := runmain(m)
	if err != nil {
		log.Println(err)
	}
}

func runmain(m *testing.M) error {
	// ensure we write our temporary binaries to a directory that we'll delete
	// after running tests.
	dir, err := os.MkdirTemp("", "tempmage")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := os.Setenv(mg.CacheEnv, dir); err != nil {
		return err
	}
	if err := os.Unsetenv(mg.VerboseEnv); err != nil {
		return err
	}
	if err := os.Unsetenv(mg.DebugEnv); err != nil {
		return err
	}
	if err := os.Unsetenv(mg.IgnoreDefaultEnv); err != nil {
		return err
	}
	if err := os.Setenv(mg.CacheEnv, dir); err != nil {
		return err
	}
	if err := os.Unsetenv(mg.EnableColorEnv); err != nil {
		return err
	}
	if err := os.Unsetenv(mg.TargetColorEnv); err != nil {
		return err
	}
	if err := resetTerm(); err != nil {
		return err
	}
	m.Run()
	return nil
}

func resetTerm() error {
	if term, exists := os.LookupEnv("TERM"); exists {
		log.Printf("Current terminal: %s", term)
		// unset TERM env var in order to disable color output to make the tests simpler
		// there is a specific test for colorized output, so all the other tests can use non-colorized one
		if err := os.Unsetenv("TERM"); err != nil {
			return fmt.Errorf("failed to unset TERM: %w", err)
		}
	}
	if err := os.Setenv(mg.EnableColorEnv, "false"); err != nil {
		return fmt.Errorf("failed to set %s: %w", mg.EnableColorEnv, err)
	}
	return nil
}

func TestTransitiveDepCache(t *testing.T) {
	cache, err := internal.OutputDebug("go", "env", "GOCACHE")
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
	defer func() { _ = os.Rename("testdata/transitiveDeps/dep/dog.notgo", "testdata/transitiveDeps/dep/dog.go") }()
	if err := os.Rename("testdata/transitiveDeps/dep/cat.notgo", "testdata/transitiveDeps/dep/cat.go"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Rename("testdata/transitiveDeps/dep/cat.go", "testdata/transitiveDeps/dep/cat.notgo") }()
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

func TestTransitiveHashFast(t *testing.T) {
	cache, err := internal.OutputDebug("go", "env", "GOCACHE")
	if err != nil {
		t.Fatal(err)
	}
	if cache == "" {
		t.Skip("skipping hashfast tests on go version without cache")
	}

	// Test that if we change a transitive dep, that we don't recompile.
	// We intentionally run the first time without hashfast to ensure that
	// we recompile the binary with the current code.
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
	defer func() { _ = os.Rename("testdata/transitiveDeps/dep/dog.notgo", "testdata/transitiveDeps/dep/dog.go") }()
	if err := os.Rename("testdata/transitiveDeps/dep/cat.notgo", "testdata/transitiveDeps/dep/cat.go"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Rename("testdata/transitiveDeps/dep/cat.go", "testdata/transitiveDeps/dep/cat.notgo") }()
	stderr.Reset()
	stdout.Reset()
	inv.HashFast = true
	code = Invoke(inv)
	if code != 0 {
		t.Fatalf("got code %v, err: %s", code, stderr)
	}
	// we should still get woof, even though the dependency was changed to
	// return "meow", because we're only hashing the top level magefiles, not
	// dependencies.
	if actual := stdout.String(); actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
}

func TestListMagefilesMain(t *testing.T) {
	files, err := Magefiles("testdata/mixed_main_files", "", "", false)
	if err != nil {
		t.Errorf("error from magefile list: %v", err)
	}
	expected := []string{
		filepath.FromSlash("testdata/mixed_main_files/mage_helpers.go"),
		filepath.FromSlash("testdata/mixed_main_files/magefile.go"),
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestListMagefilesIgnoresGOOS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("GOOS", "linux")
	} else {
		t.Setenv("GOOS", "windows")
	}
	files, err := Magefiles("testdata/goos_magefiles", "", "", false)
	if err != nil {
		t.Errorf("error from magefile list: %v", err)
	}
	var expected []string
	if runtime.GOOS == "windows" {
		expected = []string{filepath.FromSlash("testdata/goos_magefiles/magefile_windows.go")}
	} else {
		expected = []string{filepath.FromSlash("testdata/goos_magefiles/magefile_nonwindows.go")}
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestListMagefilesIgnoresRespectsGOOSArg(t *testing.T) {
	var goos string
	if runtime.GOOS == "windows" {
		goos = "linux"
	} else {
		goos = "windows"
	}
	// Set GOARCH as amd64 because windows is not on all non-x86 architectures.
	files, err := Magefiles("testdata/goos_magefiles", goos, "amd64", false)
	if err != nil {
		t.Errorf("error from magefile list: %v", err)
	}
	var expected []string
	if goos == "windows" {
		expected = []string{filepath.FromSlash("testdata/goos_magefiles/magefile_windows.go")}
	} else {
		expected = []string{filepath.FromSlash("testdata/goos_magefiles/magefile_nonwindows.go")}
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestCompileDiffGoosGoarch(t *testing.T) {
	target := t.TempDir()

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

	outputFile := filepath.Join(target, "output")
	inv := Invocation{
		Stderr:     stderr,
		Stdout:     stdout,
		Debug:      true,
		Dir:        "testdata",
		CompileOut: outputFile,
		GOOS:       goos,
		GOARCH:     goarch,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("got code %v, err: %s", code, stderr)
	}
	exeOS, arch, err := fileData(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if goos == "windows" {
		if exeOS != winExe {
			t.Error("ran with GOOS=windows but did not produce a windows exe")
		}
	} else {
		if exeOS != macExe {
			t.Error("ran with GOOS=darwin but did not produce a mac exe")
		}
	}
	if goarch == "amd64" {
		if arch != arch64 {
			t.Error("ran with GOARCH=amd64 but did not produce a 64 bit exe")
		}
	} else {
		if arch != arch32 {
			t.Error("ran with GOARCH=386 but did not produce a 32 bit exe")
		}
	}
}

func TestListMagefilesLib(t *testing.T) {
	files, err := Magefiles("testdata/mixed_lib_files", "", "", false)
	if err != nil {
		t.Fatalf("error from magefile list: %v", err)
	}
	expected := []string{
		filepath.FromSlash("testdata/mixed_lib_files/mage_helpers.go"),
		filepath.FromSlash("testdata/mixed_lib_files/magefile.go"),
	}
	if !reflect.DeepEqual(files, expected) {
		t.Fatalf("expected %q but got %q", expected, files)
	}
}

func TestMixedMageImports(t *testing.T) {
	_ = resetTerm()
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

func TestMagefilesFolder(t *testing.T) {
	_ = resetTerm()
	wd, err := os.Getwd()
	t.Log(wd)
	if err != nil {
		t.Fatalf("finding current working directory: %v", err)
	}
	if err := os.Chdir("testdata/with_magefiles_folder"); err != nil {
		t.Fatalf("changing to magefolders tests data: %v", err)
	}
	// restore previous state
	defer func() { _ = os.Chdir(wd) }()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "",
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

func TestMagefilesFolderMixedWithMagefiles(t *testing.T) {
	_ = resetTerm()
	wd, err := os.Getwd()
	t.Log(wd)
	if err != nil {
		t.Fatalf("finding current working directory: %v", err)
	}
	if err := os.Chdir("testdata/with_magefiles_folder_and_mage_files_in_dot"); err != nil {
		t.Fatalf("changing to magefolders tests data: %v", err)
	}
	// restore previous state
	defer func() { _ = os.Chdir(wd) }()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "",
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

	expectedErr := "[WARNING] You have both a magefiles directory and mage files in the current directory, in future versions the files will be ignored in favor of the directory\n"
	actualErr := stderr.String()
	if actualErr != expectedErr {
		t.Fatalf("expected Warning %q but got %q", expectedErr, actualErr)
	}
}

func TestUntaggedMagefilesFolder(t *testing.T) {
	_ = resetTerm()
	wd, err := os.Getwd()
	t.Log(wd)
	if err != nil {
		t.Fatalf("finding current working directory: %v", err)
	}
	if err := os.Chdir("testdata/with_untagged_magefiles_folder"); err != nil {
		t.Fatalf("changing to magefolders tests data: %v", err)
	}
	// restore previous state
	defer func() { _ = os.Chdir(wd) }()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "",
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

func TestMixedTaggingMagefilesFolder(t *testing.T) {
	_ = resetTerm()
	wd, err := os.Getwd()
	t.Log(wd)
	if err != nil {
		t.Fatalf("finding current working directory: %v", err)
	}
	if err := os.Chdir("testdata/with_mixtagged_magefiles_folder"); err != nil {
		t.Fatalf("changing to magefolders tests data: %v", err)
	}
	// restore previous state
	defer func() { _ = os.Chdir(wd) }()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}
	expected := "Targets:\n  build            \n  untaggedBuild    \n"
	actual := stdout.String()
	if actual != expected {
		t.Fatalf("expected %q but got %q", expected, actual)
	}
}

func TestSetDirWithMagefilesFolder(t *testing.T) {
	_ = resetTerm()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "testdata/setdir_with_magefiles_folder",
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
	c := exec.CommandContext(context.Background(), "go", "run", "main.go")
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
	t.Setenv("MAGEFILE_VERBOSE", "true")
	inv, _, err := Parse(io.Discard, io.Discard, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := true

	if !inv.Verbose {
		t.Fatalf("expected %t, but got %t ", expected, inv.Verbose)
	}
}

func TestVerboseFalseEnv(t *testing.T) {
	t.Setenv("MAGEFILE_VERBOSE", "0")
	stderr := &bytes.Buffer{}
	code := ParseAndRun(io.Discard, stderr, nil, []string{"-d", "testdata", "testverbose"})
	if code != 0 {
		t.Fatal("unexpected code", code)
	}

	if stderr.String() != "" {
		t.Fatalf("expected no output, but got %s", stderr.String())
	}
}

func TestMultilineEnv(t *testing.T) {
	t.Setenv(mg.MultilineEnv, "true")
	inv, _, err := Parse(io.Discard, io.Discard, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := true

	if !inv.Multiline {
		t.Fatalf("expected %t, but got %t ", expected, inv.Multiline)
	}
}

func TestMultilineEnvFalse(t *testing.T) {
	t.Setenv(mg.MultilineEnv, "0")
	inv, _, err := Parse(io.Discard, io.Discard, []string{})
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := false

	if inv.Multiline {
		t.Fatalf("expected %t, but got %t ", expected, inv.Multiline)
	}
}

func TestMultiline(t *testing.T) {
	t.Setenv(mg.MultilineEnv, "true")

	for _, tc := range []struct {
		name   string
		args   []string
		output string
	}{
		{
			name: "list",
			args: []string{"-d", "testdata/multiline", "-l"},
			// line returns at the end here are important to show that we don't add an extra line return.
			output: "This is a global comment for the mage output.\nIt should retain line returns.\n\nTargets:",
		},
		{
			name:   "help-func",
			args:   []string{"-d", "testdata/multiline", "-h", "doit"},
			output: "DoIt is a dummy function with a multiline comment.\nThat should show up with multiple lines.\n\nUsage:",
		},
		{
			name:   "help-func",
			args:   []string{"-d", "testdata/multiline", "-h", "sub:doittoo"},
			output: "DoItToo is a dummy function with a multiline comment.\nHere's the second line.\n\nUsage:",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			code := ParseAndRun(stdout, stderr, nil, tc.args)
			if code != 0 {
				t.Log("stderr:", stderr)
				t.Log("stdout:", stdout)
				t.Fatal("unexpected code", code)
			}

			if !strings.Contains(stdout.String(), tc.output) {
				t.Errorf("Could not find %q in output:\n%s", tc.output, stdout.String())
			}
		})
	}
}

func TestMultilineTag(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		output string
	}{
		{
			name: "list",
			args: []string{"-d", "testdata/multiline/tag", "-l"},
			// line returns at the end here are important to show that we don't add an extra line return.
			output: "This is a global comment for the mage output.\nIt should retain line returns.\n\nTargets:",
		},
		{
			name:   "help-func",
			args:   []string{"-d", "testdata/multiline/tag", "-h", "doit"},
			output: "DoIt is a dummy function with a multiline comment.\nThat should show up with multiple lines.\n\nUsage:",
		},
		{
			name:   "help-func",
			args:   []string{"-d", "testdata/multiline/tag", "-h", "sub:doittoo"},
			output: "DoItToo is a dummy function with a multiline comment.\nHere's the second line.\n\nUsage:",
		},
	} {
		// The tests should be the same regardless of the environment variable, because the tag should override.
		testFunc := func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			code := ParseAndRun(stdout, stderr, nil, tc.args)
			if code != 0 {
				t.Log("stderr:", stderr)
				t.Log("stdout:", stdout)
				t.Fatal("unexpected code", code)
			}

			if !strings.Contains(stdout.String(), tc.output) {
				t.Errorf("Could not find %q in output:\n%s", tc.output, stdout.String())
			}
		}
		t.Setenv(mg.MultilineEnv, "false")
		t.Run(tc.name+"TagFalse", testFunc)
		t.Setenv(mg.MultilineEnv, "true")
		t.Run(tc.name+"TagTrue", testFunc)
	}
}

func TestList(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/list",
		Stdout: stdout,
		Stderr: io.Discard,
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

func TestAutocomplete(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:          "./testdata/list",
		Stdout:       stdout,
		Stderr:       io.Discard,
		Autocomplete: true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "somepig\ntestverbose\n"
	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestAutocompleteNamespaces(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:          "./testdata/namespaces",
		Stdout:       stdout,
		Stderr:       io.Discard,
		Autocomplete: true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "ns:bare\nns:barectx\nns:ctxerr\nns:error\ntestnamespacedep\n"
	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestAutocompleteAliases(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:          "./testdata/alias",
		Stdout:       stdout,
		Stderr:       io.Discard,
		Autocomplete: true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	// aliases (co, st, stat) plus the actual targets (checkout, status)
	expected := "checkout\nco\nst\nstat\nstatus\n"
	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestParseAutocomplete(t *testing.T) {
	inv, _, err := Parse(io.Discard, io.Discard, []string{"-autocomplete"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if !inv.Autocomplete {
		t.Error("autocomplete should be true but was false")
	}
}

func TestParseAutocompleteConflictsWithOtherCommands(t *testing.T) {
	_, _, err := Parse(io.Discard, io.Discard, []string{"-autocomplete", "-version"})
	if err == nil {
		t.Fatal("expected error when using -autocomplete with -version")
	}
}

var terminals = []struct {
	code          string
	supportsColor bool
}{
	{"", true},
	{"vt100", false},
	{"cygwin", false},
	{"xterm-mono", false},
	{"xterm", true},
	{"xterm-vt220", true},
	{"xterm-16color", true},
	{"xterm-256color", true},
	{"screen-256color", true},
}

func TestListWithColor(t *testing.T) {
	t.Setenv(mg.EnableColorEnv, "true")
	t.Setenv(mg.TargetColorEnv, mg.Cyan.String())

	expectedPlainText := `
This is a comment on the package which should get turned into output with the list of targets.

Targets:
  somePig*       This is the synopsis for SomePig.
  testVerbose    

* default target
`[1:]

	// NOTE: using the literal string would be complicated because I would need to break it
	// in the middle and join with a normal string for the target names,
	// otherwise the single backslash would be taken literally and encoded as \\
	expectedColorizedText := "" +
		"This is a comment on the package which should get turned into output with the list of targets.\n" +
		"\n" +
		"Targets:\n" +
		"  \x1b[36msomePig*\x1b[0m       This is the synopsis for SomePig.\n" +
		"  \x1b[36mtestVerbose\x1b[0m    \n" +
		"\n" +
		"* default target\n"

	for _, terminal := range terminals {
		t.Run(terminal.code, func(t *testing.T) {
			t.Setenv("TERM", terminal.code)

			stdout := &bytes.Buffer{}
			inv := Invocation{
				Dir:    "./testdata/list",
				Stdout: stdout,
				Stderr: io.Discard,
				List:   true,
			}

			code := Invoke(inv)
			if code != 0 {
				t.Errorf("expected to exit with code 0, but got %v", code)
			}
			actual := stdout.String()
			var expected string
			if terminal.supportsColor {
				expected = expectedColorizedText
			} else {
				expected = expectedPlainText
			}

			if actual != expected {
				t.Logf("expected: %q", expected)
				t.Logf("  actual: %q", actual)
				t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
			}
		})
	}
}

func TestNoArgNoDefaultList(t *testing.T) {
	_ = resetTerm()
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
	t.Setenv(mg.IgnoreDefaultEnv, "1")
	_ = resetTerm()

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
		Stdout: io.Discard,
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
		Stderr: io.Discard,
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
		Stdout: io.Discard,
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
		Stdout: io.Discard,
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

// Test if the -keep flag does keep the mainfile around after running.
func TestKeepFlag(t *testing.T) {
	buildFile := fmt.Sprintf("./testdata/keep_flag/%s", mainfile)
	_ = os.Remove(buildFile)
	t.Cleanup(func() { _ = os.Remove(buildFile) })
	w := tLogWriter{t}

	inv := Invocation{
		Dir:    "./testdata/keep_flag",
		Stdout: w,
		Stderr: w,
		Args:   []string{"noop"},
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

// TestOnlyStdLib tests if generated mainfile references anything other than the stdlib.
func TestOnlyStdLib(t *testing.T) {
	buildFile := fmt.Sprintf("./testdata/onlyStdLib/%s", mainfile)
	_ = os.Remove(buildFile)
	t.Cleanup(func() { _ = os.Remove(buildFile) })

	w := tLogWriter{t}

	inv := Invocation{
		Dir:     "./testdata/onlyStdLib",
		Stdout:  w,
		Stderr:  w,
		Args:    []string{"noop"},
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
		if !strings.HasPrefix(pkg.Dir, build.Default.GOROOT) {
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
		t.Errorf("expected 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "Unknown target specified: \"NotGonnaWork\"\n"
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
	inv, cmd, err := Parse(io.Discard, buf, []string{"-v", "-debug", "-gocmd=foo", "-d", "dir", "build", "deploy"})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if cmd == Init {
		t.Error("init should be false but was true")
	}
	if cmd == Version {
		t.Error("showVersion should be false but was true")
	}
	if !inv.Debug {
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

func TestSetWorkingDir(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := Invoke(Invocation{
		Dir:     "testdata/setworkdir",
		WorkDir: "testdata/setworkdir/data",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"TestWorkingDir"},
	})

	if code != 0 {
		t.Errorf(
			"expected code 0, but got %d. Stdout:\n%s\nStderr:\n%s",
			code, stdout, stderr,
		)
	}

	expected := "file1.txt, file2.txt\n"
	if out := stdout.String(); out != expected {
		t.Fatalf("expected list of files to be %q, but was %q", expected, out)
	}
}

// Test the timeout option.
func TestTimeout(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:     "testdata/context",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"timeout"},
		Timeout: 100 * time.Millisecond,
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
	_, _, err := Parse(io.Discard, buf, []string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatal("unexpected error", err)
	}
	buf2 := &bytes.Buffer{}
	_, _, err = Parse(io.Discard, buf2, []string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
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
		Stderr: io.Discard,
		Args:   []string{"panics"},
		Help:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "Function that panics.\n\nUsage:\n\n\tmage panics\n\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestHelpAlias(t *testing.T) {
	stdout := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: io.Discard,
		Args:   []string{"status"},
		Help:   true,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual := stdout.String()
	expected := "Prints status.\n\nUsage:\n\n\tmage status\n\nAliases: st, stat\n\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
	inv = Invocation{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: io.Discard,
		Args:   []string{"checkout"},
		Help:   true,
	}
	stdout.Reset()
	code = Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v", code)
	}
	actual = stdout.String()
	expected = "Usage:\n\n\tmage checkout\n\nAliases: co\n\n"
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
		Stderr: io.Discard,
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
	log.SetOutput(io.Discard)
	inv := Invocation{
		Dir:    "./testdata/invalid_alias",
		Stdout: io.Discard,
		Stderr: stderr,
		Args:   []string{"co"},
	}
	code := Invoke(inv)
	if code != 2 {
		t.Errorf("expected to exit with code 2, but got %v", code)
	}
	actual := stderr.String()
	expected := "Unknown target specified: \"co\"\n"
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
	compileDir := t.TempDir()
	name := filepath.Join(compileDir, "mage_out")
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: name,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(stdout, stderr *bytes.Buffer, filename string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.CommandContext(context.Background(), filename, args...)
		cmd.Env = os.Environ()
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	// get help to target with flag -h target
	if err := run(stdout, stderr, name, "-h", "deploy"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(stdout.String())
	want := "This is the synopsis for Deploy. This part shouldn't show up.\n\nUsage:\n\n\t" + filepath.Base(name) + " deploy"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// run target with verbose flag -v
	if err := run(stdout, stderr, name, "-v", "testverbose"); err != nil {
		t.Fatal(err)
	}
	got = stderr.String()
	want = "hi!"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	// pass list flag -l
	if err := run(stdout, stderr, name, "-l"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "This is the synopsis for Deploy"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	want = "This is very verbose"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	// pass flag -t 1ms
	err := run(stdout, stderr, name, "-t", "1ms", "sleep")
	if err == nil {
		t.Fatal("expected an error because of timeout")
	}
	got = stderr.String()
	want = "context deadline exceeded"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
}

func TestCompiledEnvironmentVars(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/compiled"
	compileDir := t.TempDir()
	name := filepath.Join(compileDir, "mage_out")
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: name,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(stdout, stderr *bytes.Buffer, filename string, envval string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.CommandContext(context.Background(), filename, args...)
		cmd.Env = []string{envval}
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	if err := run(stdout, stderr, name, "MAGEFILE_HELP=1", "deploy"); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	want := "This is the synopsis for Deploy. This part shouldn't show up.\n\nUsage:\n\n\t" + filepath.Base(name) + " deploy\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if err := run(stdout, stderr, name, mg.VerboseEnv+"=1", "testverbose"); err != nil {
		t.Fatal(err)
	}
	got = stderr.String()
	want = "hi!"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, "MAGEFILE_LIST=1"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "This is the synopsis for Deploy"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
	want = "This is very verbose"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	if err := run(stdout, stderr, name, mg.IgnoreDefaultEnv+"=1"); err != nil {
		t.Fatal(err)
	}
	got = stdout.String()
	want = "Compiled package description."
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}

	err := run(stdout, stderr, name, "MAGEFILE_TIMEOUT=1ms", "sleep")
	if err == nil {
		t.Fatal("expected an error because of timeout")
	}
	got = stderr.String()
	want = "context deadline exceeded"
	if !strings.Contains(got, want) {
		t.Errorf("got %q, does not contain %q", got, want)
	}
}

func TestCompiledVerboseFlag(t *testing.T) {
	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/compiled"
	compileDir := t.TempDir()

	filename := filepath.Join(compileDir, "mage_out")
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}
	inv := Invocation{
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: filename,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Errorf("expected to exit with code 0, but got %v, stderr: %s", code, stderr)
	}

	run := func(verboseEnv string, args ...string) string {
		var stdout, stderr bytes.Buffer
		args = append(args, "printverboseflag")
		cmd := exec.CommandContext(context.Background(), filename, args...)
		cmd.Env = []string{verboseEnv}
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		if err := cmd.Run(); err != nil {
			t.Fatalf("running '%s %s' with env %s failed with: %v\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), verboseEnv, err, stdout.String(), stderr.String())
		}
		return strings.TrimSpace(stdout.String())
	}

	got := run("MAGEFILE_VERBOSE=false")
	want := "mg.Verbose()==false"
	if got != want {
		t.Errorf("got %q, expected %q", got, want)
	}

	got = run("MAGEFILE_VERBOSE=false", "-v")
	want = "mg.Verbose()==true"
	if got != want {
		t.Errorf("got %q, expected %q", got, want)
	}

	got = run("MAGEFILE_VERBOSE=true")
	want = "mg.Verbose()==true"
	if got != want {
		t.Errorf("got %q, expected %q", got, want)
	}

	got = run("MAGEFILE_VERBOSE=true", "-v=false")
	want = "mg.Verbose()==false"
	if got != want {
		t.Errorf("got %q, expected %q", got, want)
	}
}

func TestCompiledDeterministic(t *testing.T) {
	dir := "./testdata/compiled"
	compileDir := t.TempDir()

	var exp string
	outFile := filepath.Join(dir, mainfile)

	// compile a couple times to be sure
	for i, run := range []string{"one", "two", "three", "four"} {
		run := run
		t.Run(run, func(t *testing.T) {
			// probably don't run this parallel
			filename := filepath.Join(compileDir, "mage_out")
			if runtime.GOOS == "windows" {
				filename += ".exe"
			}

			inv := Invocation{
				Stderr:     os.Stderr,
				Stdout:     os.Stdout,
				Verbose:    true,
				Keep:       true,
				Dir:        dir,
				CompileOut: filename,
			}

			code := Invoke(inv)
			if code != 0 {
				t.Errorf("expected to exit with code 0, but got %v", code)
			}

			f, err := os.Open(outFile)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = f.Close() }()

			hasher := sha256.New()
			if _, err := io.Copy(hasher, f); err != nil {
				t.Fatal(err)
			}

			got := hex.EncodeToString(hasher.Sum(nil))
			// set exp on first iteration, subsequent iterations prove the compiled file is identical
			if i == 0 {
				exp = got
			}

			if i > 0 && got != exp {
				t.Errorf("unexpected sha256 hash of %s; wanted %s, got %s", outFile, exp, got)
			}
		})
	}
}

func TestClean(t *testing.T) {
	if err := os.RemoveAll(mg.CacheDir()); err != nil {
		t.Error("error removing cache dir:", err)
	}
	code := ParseAndRun(io.Discard, io.Discard, &bytes.Buffer{}, []string{"-clean"})
	if code != 0 {
		t.Errorf("expected 0, but got %v", code)
	}

	TestAlias(t) // make sure we've got something in the CACHE_DIR
	files, err := os.ReadDir(mg.CacheDir())
	if err != nil {
		t.Error("issue reading file:", err)
	}

	if len(files) < 1 {
		t.Error("Need at least 1 cached binaries to test --clean")
	}

	_, cmd, err := Parse(io.Discard, io.Discard, []string{"-clean"})
	if err != nil {
		t.Fatal(err)
	}
	if cmd != Clean {
		t.Errorf("Expected 'clean' command but got %v", cmd)
	}
	buf := &bytes.Buffer{}
	code = ParseAndRun(io.Discard, buf, &bytes.Buffer{}, []string{"-clean"})
	if code != 0 {
		t.Fatalf("expected 0, but got %v: %s", code, buf)
	}

	infos, err := os.ReadDir(mg.CacheDir())
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
	t.Setenv(testExeEnv, textOutput)
	dir := t.TempDir()
	// fake out the compiled file, since the code checks for it.
	f, err := os.CreateTemp(dir, "mage_out")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()

	buf := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if err := Compile("", "", "", dir, os.Args[0], name, []string{}, false, stderr, buf); err != nil {
		t.Log("stderr: ", stderr.String())
		t.Fatal(err)
	}
	if buf.String() != textOutput {
		t.Fatalf("We didn't run the custom go cmd. Expected output %q, but got %q", textOutput, buf)
	}
}

var runtimeVer = regexp.MustCompile(`go1\.(\d+)`)

func TestGoModules(t *testing.T) {
	_ = resetTerm()
	matches := runtimeVer.FindStringSubmatch(runtime.Version())
	if len(matches) < 2 || minorVer(t, matches[1]) < 11 {
		t.Skipf("Skipping Go modules test because go version %q is less than go1.11", runtime.Version())
	}
	dir := t.TempDir()
	// beware, mage builds in go versions older than 1.17 so both build tag formats need to be present
	err := os.WriteFile(filepath.Join(dir, "magefile.go"), []byte(`//go:build mage
// +build mage

package main

func Test() {
	print("nothing is imported here for >1.17 compatibility")
}
`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(context.Background(), "go", "mod", "init", "app")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running go mod init: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	stderr.Reset()
	stdout.Reset()

	// we need to run go mod tidy, since go build will no longer auto-add dependencies.
	cmd = exec.CommandContext(context.Background(), "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error running go mod tidy: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
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
		Stderr: io.Discard,
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
		Stderr: io.Discard,
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

func TestAliasToImport(_ *testing.T) {
}

func TestWrongDependency(t *testing.T) {
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/wrong_dep",
		Stderr: stderr,
		Stdout: io.Discard,
	}
	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected 1, but got %v", code)
	}
	expected := "Error: argument 0 (complex128), is not a supported argument type\n"
	actual := stderr.String()
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

// Regression tests, add tests to ensure we do not regress on known issues.

// TestBug508 is a regression test for: Bug: using Default with imports selects first matching func by name.
func TestBug508(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/bug508",
		Stderr: stderr,
		Stdout: stdout,
	}
	code := Invoke(inv)
	if code != 0 {
		t.Log(stderr.String())
		t.Fatalf("expected 0, but got %v", code)
	}
	expected := "test\n"
	if stdout.String() != expected {
		t.Fatalf("expected %q, but got %q", expected, stdout.String())
	}
}

// / This code liberally borrowed from https://github.com/rsc/goversion/blob/master/version/exe.go

type (
	exeType  int
	archSize int
)

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
	defer func() { _ = f.Close() }()
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
	return -1, -1, errors.New("unrecognized executable format")
}
