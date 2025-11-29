package sh

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	helperCmd    bool
	printArgs    bool
	stderr       string
	stdout       string
	exitCode     int
	printVar     string
	dryRunOutput bool
)

func init() {
	flag.BoolVar(&helperCmd, "helper", false, "")
	flag.BoolVar(&printArgs, "printArgs", false, "")
	flag.StringVar(&stderr, "stderr", "", "")
	flag.StringVar(&stdout, "stdout", "", "")
	flag.IntVar(&exitCode, "exit", 0, "")
	flag.StringVar(&printVar, "printVar", "", "")
	flag.BoolVar(&dryRunOutput, "dryRunOutput", false, "")
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

	if dryRunOutput {
		// Simulate dry-run mode and print the output of a command that would have been run.
		// We use a non-echo command to make the "DRYRUN: " prefix deterministic.
		_ = os.Setenv("MAGEFILE_DRYRUN_POSSIBLE", "1")
		_ = os.Setenv("MAGEFILE_DRYRUN", "1")
		s, err := Output("somecmd", "arg1", "arg two")
		if err != nil {
			fmt.Println("ERR:", err)
			return
		}
		fmt.Println(s)
		return
	}

	if helperCmd {
		fmt.Fprintln(os.Stderr, stderr)
		fmt.Fprintln(os.Stdout, stdout)
		os.Exit(exitCode)
	}
	os.Exit(m.Run())
}
