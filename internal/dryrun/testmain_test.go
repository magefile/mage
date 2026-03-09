package dryrun

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	printIsDryRunRequested bool
	printIsDryRunPossible  bool
	printIsDryRun          bool
)

func init() {
	flag.BoolVar(&printIsDryRunRequested, "printIsDryRunRequested", false, "")
	flag.BoolVar(&printIsDryRunPossible, "printIsDryRunPossible", false, "")
	flag.BoolVar(&printIsDryRun, "printIsDryRun", false, "")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if printIsDryRunRequested {
		fmt.Println(IsRequested())
		return
	}
	if printIsDryRunPossible {
		fmt.Println(IsPossible())
		return
	}
	if printIsDryRun {
		fmt.Println(IsDryRun())
		return
	}
	os.Exit(m.Run())
}
