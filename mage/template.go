package mage

// var only for tests
var tpl = `// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	{{range .Imports}}{{.UniqueName}} "{{.Path}}"
	{{end}}
)

func main() {
	// Use local types and functions in order to avoid name conflicts with additional magefiles.
	type arguments struct {
		Verbose       bool          // print out log statements
		List          bool          // print out a list of targets
		Help          bool          // print out help for a specific target
		IgnoreDefault bool          // ignore the default build target and print all build targets instead
		Timeout       time.Duration // set a timeout to running the targets
		Args          []string      // args contain the non-flag command-line arguments
	}
	initArgs := func() arguments {
		args := arguments{}
		// default flag set with ExitOnError and auto generated PrintDefaults should be sufficient
		flag.BoolVar(&args.Verbose, "v", false, "show verbose output when running mage targets")
		flag.BoolVar(&args.List, "l", false, "list mage targets in this directory")
		flag.BoolVar(&args.Help, "h", false, "print out help for a specific target")
		flag.BoolVar(&args.IgnoreDefault, "ignoredefault", false, "ignore the default build target")
		flag.DurationVar(&args.Timeout, "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
		flag.Parse()
		args.Args = flag.Args()
		// With golangs flag package we can't distinguish, whether a flag's default value has been:
		//  * provided by the user or
		//  * never passed by command-line and thus having been set to its default value
		// This feature would be nice because we could lookup an environment var ONLY if the
		// flag hasn't been provided by the user.
		// The answer at https://stackoverflow.com/a/51903637 provides a pretty clever workaround
		// without breaking flag.PrintDefaults, but is a bit tedious. Let's just allow
		// the override of any flag with its respective environment variable value.
		env := os.Getenv("MAGEFILE_VERBOSE")
		if val, err := strconv.ParseBool(env); err == nil {
			args.Verbose = val
		}
		env = os.Getenv("MAGEFILE_IGNOREDEFAULT")
		if val, err := strconv.ParseBool(env); err == nil {
			args.IgnoreDefault = val
		}
		env = os.Getenv("MAGEFILE_LIST")
		if val, err := strconv.ParseBool(env); err == nil {
			args.List = val
		}
		env = os.Getenv("MAGEFILE_HELP")
		if val, err := strconv.ParseBool(env); err == nil {
			args.Help = val
		}
		env = os.Getenv("MAGEFILE_TIMEOUT")
		if val, err := time.ParseDuration(env); err == nil {
			args.Timeout = val
		}
		return args
	}
	args := initArgs()

	list := func() error {
		{{with .Description}}fmt.Println(` + "`{{.}}\n`" + `)
		{{- end}}
		{{- $default := .DefaultFunc}}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		fmt.Println("Targets:")
		{{- range .Funcs}}
			fmt.Fprintln(w, "  {{lowerFirst .TargetName}}{{if and (eq .Name $default.Name) (eq .Receiver $default.Receiver)}}*{{end}}\t" + {{printf "%q" .Synopsis}})
		{{- end}}
		{{- range .Imports}}{{$imp := .}}
			{{- range .Info.Funcs}}
			fmt.Fprintln(w, "  {{lowerFirst .TargetName}}{{if and (eq .Name $default.Name) (eq .Receiver $default.Receiver)}}*{{end}}\t" + {{printf "%q" .Synopsis}})
			{{end}}
		{{- end}}
		err := w.Flush()
		{{- if .DefaultFunc.Name}}
			if err == nil {
				fmt.Println("\n* default target")
			}
		{{- end}}
		return err
	}

	var ctx context.Context
	var ctxCancel func()

	getContext := func() (context.Context, func()) {
		if ctx != nil {
			return ctx, ctxCancel
		}

		if args.Timeout != 0 {
			ctx, ctxCancel = context.WithTimeout(context.Background(), args.Timeout)
		} else {
			ctx = context.Background()
			ctxCancel = func() {}
		}
		return ctx, ctxCancel
	}

	runTarget := func(fn func(context.Context) error) interface{} {
		var err interface{}
		ctx, cancel := getContext()
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
			cancel()
			e := ctx.Err()
			fmt.Printf("ctx err: %v\n", e)
			return e
		case err = <-d:
			cancel()
			return err
		}
	}
	// This is necessary in case there aren't any targets, to avoid an unused
	// variable error.
	_ = runTarget

	handleError := func(logger *log.Logger, err interface{}) {
		if err != nil {
			logger.Printf("Error: %v\n", err)
			type code interface {
				ExitStatus() int
			}
			if c, ok := err.(code); ok {
				os.Exit(c.ExitStatus())
			}
			os.Exit(1)
		}
	}
	_ = handleError

	log.SetFlags(0)
	if !args.Verbose {
		log.SetOutput(ioutil.Discard)
	}
	logger := log.New(os.Stderr, "", 0)
	if args.List {
		if err := list(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	targets := map[string]bool {
		{{range $alias, $funci := .Aliases}}"{{lower $alias}}": true,
		{{end}}
		{{range .Funcs}}"{{lower .TargetName}}": true,
		{{end}}
		{{range .Imports}}
			{{$imp := .}}
			{{range $alias, $funci := .Info.Aliases}}"{{if ne $imp.Alias "."}}{{lower $imp.Alias}}:{{end}}{{lower $alias}}": true,
			{{end}}
			{{range .Info.Funcs}}"{{lower .TargetName}}": true,
			{{end}}
		{{end}}
	}

	var unknown []string
	for _, arg := range args.Args {
		if !targets[strings.ToLower(arg)] {
			unknown = append(unknown, arg)
		}
	}
	if len(unknown) == 1 {
		logger.Println("Unknown target specified:", unknown[0])
		os.Exit(2)
	}
	if len(unknown) > 1 {
		logger.Println("Unknown targets specified:", strings.Join(unknown, ", "))
		os.Exit(2)
	}

	if args.Help {
		if len(args.Args) < 1 {
			logger.Println("no target specified")
			os.Exit(1)
		}
		switch strings.ToLower(args.Args[0]) {
			{{range .Funcs}}case "{{lower .TargetName}}":
				fmt.Print("mage {{lower .TargetName}}:\n\n")
				{{if ne .Comment "" -}}
				fmt.Println({{printf "%q" .Comment}})
				fmt.Println()
				{{end}}
				var aliases []string
				{{- $name := .Name -}}
				{{- $recv := .Receiver -}}
				{{range $alias, $func := $.Aliases}}
				{{if and (eq $name $func.Name) (eq $recv $func.Receiver)}}aliases = append(aliases, "{{$alias}}"){{end -}}
				{{- end}}
				if len(aliases) > 0 {
					fmt.Printf("Aliases: %s\n\n", strings.Join(aliases, ", "))
				}
				return
			{{end}}
			default:
				logger.Printf("Unknown target: %q\n", args.Args[0])
				os.Exit(1)
		}
	}
	if len(args.Args) < 1 {
	{{- if .DefaultFunc.Name}}
		if args.IgnoreDefault {
			if err := list(); err != nil {
				logger.Println("Error:", err)
				os.Exit(1)
			}
			return
		}
		{{.DefaultFunc.ExecCode}}
		handleError(logger, err)
		return
	{{- else}}
		if err := list(); err != nil {
			logger.Println("Error:", err)
			os.Exit(1)
		}
		return
	{{- end}}
	}
	for _, target := range args.Args {
		switch strings.ToLower(target) {
		{{range $alias, $func := .Aliases}}
			case "{{lower $alias}}":
				target = "{{$func.TargetName}}"
		{{- end}}
		}
		switch strings.ToLower(target) {
		{{range .Funcs }}
			case "{{lower .TargetName}}":
				if args.Verbose {
					logger.Println("Running target:", "{{.TargetName}}")
				}
				{{.ExecCode}}
				handleError(logger, err)
		{{- end}}
		{{range .Imports}}
		{{$imp := .}}
			{{range .Info.Funcs }}
				case "{{lower .TargetName}}":
					if args.Verbose {
						logger.Println("Running target:", "{{.TargetName}}")
					}
					{{.ExecCode}}
					handleError(logger, err)
			{{- end}}
		{{- end}}
		default:
			// should be impossible since we check this above.
			logger.Printf("Unknown target: %q\n", args.Args[0])
			os.Exit(1)
		}
	}
}




`
