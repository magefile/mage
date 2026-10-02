package shctx

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	helperCmd        bool
	printArgs        bool
	stderr           string
	stdout           string
	exitCode         int
	sleepTimeSeconds int
	printVar         string
)

func init() { //nolint:gochecknoinits // required for test flag setup
	flag.BoolVar(&helperCmd, "helper", false, "")
	flag.BoolVar(&printArgs, "printArgs", false, "")
	flag.StringVar(&stderr, "stderr", "", "")
	flag.StringVar(&stdout, "stdout", "", "")
	flag.IntVar(&exitCode, "exit", 0, "")
	flag.IntVar(&sleepTimeSeconds, "sleep", 0, "")
	flag.StringVar(&printVar, "printVar", "", "")
}

func TestMain(m *testing.M) {
	flag.Parse()

	if printArgs {
		fmt.Println(flag.Args())
		return
	}
	if printVar != "" {
		fmt.Println(os.Getenv(printVar))
		return
	}

	if sleepTimeSeconds > 0 {
		time.Sleep(time.Duration(sleepTimeSeconds) * time.Second)
	}

	if helperCmd {
		_, _ = fmt.Fprintln(os.Stderr, stderr)
		_, _ = fmt.Fprintln(os.Stdout, stdout)
		os.Exit(exitCode)
	}
	os.Exit(m.Run())
}
