package mage

const tpl = `// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

func main() {
	log.SetFlags(0)
	if os.Getenv("MAGEFILE_LIST") != "" {
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
		log.Println("Targets: ")
		{{- range .Funcs}}
		fmt.Fprintln(w, "  {{lower .Name}}\t{{.Synopsis}}")
		{{- end}}
		if err := w.Flush(); err != nil {
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
			fmt.Println(err)
			type code interface { ExitCode() int }
			if c, ok := err.(code); ok {
				os.Exit(c.ExitCode())
			}
		}
	}()
	if len(os.Args) < 2 {
	{{- if .HasDefault}}{{with .Default}}
		{{- if .IsError}}
		if err := {{.Name}}(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
		{{- else}}
		{{.Name}}()
		{{- end}}{{end}}
		return
	{{- else}}
		fmt.Println("No default (Build) target exists")
		os.Exit(1)
	{{- end}}
	}
	switch strings.ToLower(os.Args[1]) {
	{{range .Funcs -}}
	case "{{lower .Name}}":
		{{if .IsError -}}
			if err := {{.Name}}(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		{{else -}}
			{{.Name}}()
		{{- end}}
	{{end}}
	default:
		fmt.Printf("Unknown target: %q\n", os.Args[1])
		os.Exit(1)
	}
}`
