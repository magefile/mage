package mage

// var only for tests
var tpl = `// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/magefile/mage/mg"
)

func main() {
	log.SetFlags(0)
	if os.Getenv("MAGEFILE_LIST") != "" {
		if err := list(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if os.Getenv("MAGEFILE_HELP") != "" {
		if len(os.Args) < 2 {
			fmt.Println("no target specified")
			os.Exit(1)
		}
		switch strings.ToLower(os.Args[1]) {
			{{range .Funcs}}case "{{lower .Name}}":
				fmt.Print("mage {{lower .Name}}:\n\n")
                fmt.Println(` + "`{{.Comment}}`" + `)
				return
			{{end}}
			default:
				fmt.Printf("Unknown target: %q\n", os.Args[1])
				os.Exit(1)
		}	}

	if os.Getenv("MAGEFILE_VERBOSE") == "" {
		log.SetOutput(ioutil.Discard)
	}
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			type code interface { ExitStatus() int }
			if c, ok := err.(code); ok {
				os.Exit(c.ExitStatus())
			}
			os.Exit(1)
		}
	}()
	if len(os.Args) < 2 {
	{{- if .Default}}
		{{.DefaultFunc.TemplateString}}
		return
	{{- else}}
		if err := list(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	{{- end}}
	}
	switch strings.ToLower(os.Args[1]) {
	{{range .Funcs -}}
	case "{{lower .Name}}":
		{{.TemplateString}}
	{{end}}
	default:
		fmt.Printf("Unknown target: %q\n", os.Args[1])
		os.Exit(1)
	}
}

func list() error {
	{{$default := .Default}}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	log.Println("Targets: ")
	{{- range .Funcs}}
	fmt.Fprintln(w, "  {{lower .Name}}{{if eq .Name $default}}*{{end}}\t{{.Synopsis}}")
	{{- end}}
	err := w.Flush()
	{{- if .Default}}
	if err == nil {
		fmt.Println("\n* default target")
	}
	{{- end}}
	return err
}

`
