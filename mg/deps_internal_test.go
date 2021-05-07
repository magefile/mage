package mg

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestDepsLogging(t *testing.T) {
	os.Setenv("MAGEFILE_VERBOSE", "1")
	defer os.Unsetenv("MAGEFILE_VERBOSE")
	buf := &bytes.Buffer{}

	defaultLogger := logger
	logger = log.New(buf, "", 0)
	defer func() { logger = defaultLogger }()

	foo()

	if strings.Count(buf.String(), "Running dependency: github.com/magefile/mage/mg.baz") != 1 {
		t.Fatalf("expected one baz to be logged, but got\n%s", buf)
	}
}

func foo() {
	Deps(bar, baz)
}

func bar() {
	Deps(baz)
}

func baz() {}

func TestDepWasNotInvoked(t *testing.T) {
	fn1 := func() error {
		return nil
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		gotErr := fmt.Sprint(err)
		wantErr := "A dependency of the current target was defined improperly, with parenthesis"
		if !strings.Contains(gotErr, wantErr) {
			t.Fatalf(`expected to get "%s" but got "%s"`, wantErr, gotErr)
		}
	}()
	func(fns ...interface{}) {
		checkFns(fns)
	}(fn1())
}
