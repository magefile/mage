package mage

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		if os.IsExist(err) {
			os.RemoveAll(fmt.Sprintf("%s/*"))
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
	c := exec.Command("go", "run", "main.go", "testverbose")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	if err != nil {
		t.Error("error:", err)
	}
	actual := string(b)
	expected := ""
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
	c = exec.Command("go", "run", "main.go", "-v", "testverbose")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err = c.CombinedOutput()
	if err != nil {
		t.Error("error:", err)
	}
	actual = string(b)
	expected = "hi!\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestGoRunError(t *testing.T) {
	c := exec.Command("go", "run", "main.go", "returnsnonnilerror")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	actualErr := err.Error()
	expectedErr := "exit status 1"
	if actualErr != expectedErr {
		t.Fatalf("expected %q, but got %q", expectedErr, actualErr)
	}
	actual := string(b)
	expected := "Error: bang!\nexit status 1\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestGoRunPanics(t *testing.T) {
	c := exec.Command("go", "run", "main.go", "panics")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	actualErr := err.Error()
	expectedErr := "exit status 1"
	if actualErr != expectedErr {
		t.Fatalf("expected %q, but got %q", expectedErr, actualErr)
	}
	actual := string(b)
	expected := "Error: boom!\nexit status 1\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}

func TestGoRunPanicsErr(t *testing.T) {
	c := exec.Command("go", "run", "main.go", "panicserr")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	actualErr := err.Error()
	expectedErr := "exit status 1"
	if actualErr != expectedErr {
		t.Fatalf("expected %q, but got %q", expectedErr, actualErr)
	}
	actual := string(b)
	expected := "Error: kaboom!\nexit status 1\n"
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
	c := exec.Command("go", "run", "main.go", "-keep", "noop")
	c.Dir = "./testdata/keep_flag"
	c.Env = os.Environ()
	_, err := c.CombinedOutput()
	if err != nil && strings.Compare(err.Error(), "exit status 1") != 0 {
		t.Error("unexpected error:", err)
	}

	if _, err := os.Stat(buildFile); os.IsNotExist(err) {
		t.Fatalf("expected file %q to exist but it did not", buildFile)
	}
	os.Remove(buildFile)
}

// Test the timeout option
func TestTimeout(t *testing.T) {
	c := exec.Command("go", "run", "main.go", "-t", "1", "timeout")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	actualErr := err.Error()
	expectedErr := "exit status 1"
	if actualErr != expectedErr {
		t.Fatalf("expected %q, but got %q", expectedErr, actualErr)
	}
	actual := string(b)
	expected := "Error: context deadline exceeded\nexit status 1\n"
	if actual != expected {
		t.Fatalf("expected %q, but got %q", expected, actual)
	}
}
