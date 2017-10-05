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
)

func main() {
	log.SetFlags(0)
	if os.Getenv("MAGEFILE_VERBOSE") == "" {
		log.SetOutput(ioutil.Discard)
	}
	logger := log.New(os.Stderr, "", 0)
	if os.Getenv("MAGEFILE_LIST") != "" {
		if err := list(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	if os.Getenv("MAGEFILE_HELP") != "" {
		if len(os.Args) < 2 {
			logger.Println("no target specified")
			os.Exit(1)
		}
		switch strings.ToLower(os.Args[1]) {
			{{range .Funcs}}case "{{lower .Name}}":
				fmt.Print("mage {{lower .Name}}:\n\n")
                fmt.Println(` + "`{{.Comment}}`" + `)
				return
			{{end}}
			default:
				logger.Printf("Unknown target: %q\n", os.Args[1])
				os.Exit(1)
		}	}


	defer func() {
		err := recover()
		if err != nil {
			logger.Printf("Error: %v\n", err)
			type code interface { ExitStatus() int }
			if c, ok := err.(code); ok {
				os.Exit(c.ExitStatus())
			}
			os.Exit(1)
		}
	}()
	if len(os.Args) < 2 {
	{{- if .Default}}
		{{- if .DefaultError}}
		if err := Default(); err != nil {
			logger.Println(err)
			os.Exit(1)
		}
		return
		{{- else}}
		Default()
		{{- end}}
		return
	{{- else}}
		if err := list(); err != nil {
			logger.Println("Error:", err)
			os.Exit(1)
		}
		return
	{{- end}}
	}
	switch strings.ToLower(os.Args[1]) {
	{{range .Funcs -}}
	case "{{lower .Name}}":
		{{if .IsError -}}
			if err := {{.Name}}(); err != nil {
				logger.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		{{else -}}
			{{.Name}}()
		{{- end}}
	{{end}}
	default:
		logger.Printf("Unknown target: %q\n", os.Args[1])
		os.Exit(1)
	}
}

func list() error {
	{{$default := .Default}}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Println("Targets:")
	{{- range .Funcs}}
	fmt.Fprintln(w, "  {{lowerfirst .Name}}{{if eq .Name $default}}*{{end}}\t{{.Synopsis}}")
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

var mageTpl = `// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
`
