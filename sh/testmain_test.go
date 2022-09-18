package sh

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	helperCmd bool
	printArgs bool
	stderr    string
	stdout    string
	exitCode  int
	printVar  string
	sleep     time.Duration
)

func init() {
	flag.BoolVar(&helperCmd, "helper", false, "")
	flag.BoolVar(&printArgs, "printArgs", false, "")
	flag.StringVar(&stderr, "stderr", "", "")
	flag.StringVar(&stdout, "stdout", "", "")
	flag.IntVar(&exitCode, "exit", 0, "")
	flag.StringVar(&printVar, "printVar", "", "")
	flag.DurationVar(&sleep, "sleep", 0, "")
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

	if sleep != 0 {
		time.Sleep(sleep)
		return
	}

	if helperCmd {
		fmt.Fprintln(os.Stderr, stderr)
		fmt.Fprintln(os.Stdout, stdout)
		os.Exit(exitCode)
	}
	os.Exit(m.Run())
}
