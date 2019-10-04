package toplevel

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Arguments is parsed arguments of generated Mage binary
type Arguments struct {
	Verbose bool          // print out log statements
	List    bool          // print out a list of targets
	Help    bool          // print out help for a specific target
	Timeout time.Duration // set a timeout to running the targets
	Args    []string      // args contain the non-flag command-line arguments
}

func parseBool(env string) bool {
	val := os.Getenv(env)
	if val == "" {
		return false
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("warning: environment variable %s is not a valid bool value: %v", env, val)
		return false
	}
	return b
}

func parseDuration(env string) time.Duration {
	val := os.Getenv(env)
	if val == "" {
		return 0
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("warning: environment variable %s is not a valid duration value: %v", env, val)
		return 0
	}
	return d
}

// Main is the main function for generated Mage binary
func Main() Arguments {
	args := Arguments{}
	fs := flag.FlagSet{}
	fs.SetOutput(os.Stdout)

	// default flag set with ExitOnError and auto generated PrintDefaults should be sufficient
	fs.BoolVar(&args.Verbose, "v", parseBool("MAGEFILE_VERBOSE"), "show verbose output when running targets")
	fs.BoolVar(&args.List, "l", parseBool("MAGEFILE_LIST"), "list targets for this binary")
	fs.BoolVar(&args.Help, "h", parseBool("MAGEFILE_HELP"), "print out help for a specific target")
	fs.DurationVar(&args.Timeout, "t", parseDuration("MAGEFILE_TIMEOUT"), "timeout in duration parsable format (e.g. 5m30s)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stdout, `
%s [options] [target]

Commands:
  -l    list targets in this binary
  -h    show this help

Options:
  -h    show description of a target
  -t <string>
        timeout in duration parsable format (e.g. 5m30s)
  -v    show verbose output when running targets
 `[1:], filepath.Base(os.Args[0]))
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		// flag will have printed out an error already.
		os.Exit(0)
	}
	args.Args = fs.Args()

	if args.Help && len(args.Args) == 0 {
		fs.Usage()
		os.Exit(0)
	}

	return args
}

// RunTarget runs one Mage target
func RunTarget(ctx context.Context, fn func(context.Context) error) interface{} {
	var err interface{}
	d := make(chan interface{})
	go func() {
		defer func() {
			err := recover()
			d <- err
		}()
		err := fn(ctx)
		d <- err
	}()
	select {
	case <-ctx.Done():
		e := ctx.Err()
		fmt.Printf("ctx err: %v\n", e)
		return e
	case err = <-d:
		return err
	}
}
