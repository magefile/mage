package mg

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestDepsOfDeps(t *testing.T) {
	os.Setenv("MAGEFILE_VERBOSE", "1")
	defer os.Unsetenv("MAGEFILE_VERBOSE")
	buf := bytes.NewBuffer(make([]byte, 4096))

	defaultLogger := logger
	logger = log.New(buf, "", 0)
	defer func() { logger = defaultLogger }()

	ch := make(chan string, 1)
	f := func() {
		Deps(func() {})
		ch <- "f"
	}
	Deps(f)

	<-ch
	if !strings.Contains(buf.String(), "Running dependency") {
		t.Fatal("expected dependencies to be logged")
	}
}
