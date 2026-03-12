package sh

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	helperCmd bool
	printArgs bool
	stderr    string
	stdout    string
	exitCode  int
	printVar  string
)

func init() { //nolint:gochecknoinits // required for test flag setup
	flag.BoolVar(&helperCmd, "helper", false, "")
	flag.BoolVar(&printArgs, "printArgs", false, "")
	flag.StringVar(&stderr, "stderr", "", "")
	flag.StringVar(&stdout, "stdout", "", "")
	flag.IntVar(&exitCode, "exit", 0, "")
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

	if helperCmd {
		_, _ = fmt.Fprintln(os.Stderr, stderr)
		_, _ = fmt.Fprintln(os.Stdout, stdout)
		os.Exit(exitCode) //nolint:revive // sub-process test helper requires explicit exit
	}
	os.Exit(m.Run()) //nolint:revive // TestMain with branching logic needs explicit exit
}
