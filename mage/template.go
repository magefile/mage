package mage

// this template uses the "data"

// var only for tests
var mageMainfileTplString = `// +build ignore

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	{{- if .DefaultFunc.Name}}
	"strconv"
        {{end}}
	"strings"
	"text/tabwriter"
	{{range .Imports}}{{.UniqueName}} "{{.Path}}"
	{{end}}

        "github.com/magefile/mage/mg"
        "github.com/magefile/mage/toplevel"
)

func main() {
        mg.SetModule("{{.Module}}")

        args := toplevel.Main()

	list := func() error {
		{{with .Description}}fmt.Println(` + "`{{.}}\n`" + `)
		{{- end}}
		{{- $default := .DefaultFunc}}
		targets := map[string]string{
		{{- range .Funcs}}
			"{{lowerFirst .TargetName}}{{if and (eq .Name $default.Name) (eq .Receiver $default.Receiver)}}*{{end}}": {{printf "%q" .Synopsis}},
		{{- end}}
		{{- range .Imports}}{{$imp := .}}
			{{- range .Info.Funcs}}
			"{{lowerFirst .TargetName}}{{if and (eq .Name $default.Name) (eq .Receiver $default.Receiver)}}*{{end}}": {{printf "%q" .Synopsis}},
			{{- end}}
		{{- end}}
		}

		keys := make([]string, 0, len(targets))
		for name := range targets {
			keys = append(keys, name)
		}
		sort.Strings(keys)

		fmt.Println("Targets:")
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		for _, name := range keys {
			fmt.Fprintf(w, "  %v\t%v\n", name, targets[name])
		}
		err := w.Flush()
		{{- if .DefaultFunc.Name}}
			if err == nil {
				fmt.Println("\n* default target")
			}
		{{- end}}
		return err
	}

	ctx := context.Background()
        if args.Timeout != 0 {
            var cancel context.CancelFunc
            ctx, cancel = context.WithTimeout(ctx, args.Timeout)
            defer cancel()
        }

	runTarget := func(fn func(context.Context) error) interface{} {
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
	// This is necessary in case there aren't any targets, to avoid an unused
	// variable error.
	_ = runTarget

	handleError := func(logger *log.Logger, err interface{}) {
		if err != nil {
			logger.Printf("Error: %+v\n", err)
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
				fmt.Print("{{$.BinaryName}} {{lower .TargetName}}:\n\n")
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
		ignoreDefault, _ := strconv.ParseBool(os.Getenv("MAGEFILE_IGNOREDEFAULT"))
		if ignoreDefault {
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
