package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"text/template"

	"github.com/pkg/errors"

	"github.com/magefile/mage/build"
	"github.com/magefile/mage/parse"
)

var output = template.Must(template.New("").Funcs(map[string]interface{}{
	"lower": strings.ToLower,
}).Parse(tpl))

const mainfile = "mage_output_file.go"

var (
	force, verbose, list, help bool
	tags                       string
)

func init() {
	flag.BoolVar(&force, "f", false, "force recreation of compiled magefile")
	flag.BoolVar(&verbose, "v", false, "show verbose output when running mage targets")
	flag.BoolVar(&list, "l", false, "list mage targets in this directory")
	flag.BoolVar(&help, "h", false, "show this help")
	flag.Usage = func() {
		fmt.Println("mage [options] [target]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	if help && len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	files := magefiles()

	filename, err := hashStrings(files)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	dir := confDir()
	out := filepath.Join(dir, filename)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	if !force {
		if _, err := os.Stat(out); err == nil {
			run(out, flag.Args()...)
			return
		}
	}

	fns, err := parse.Package(".", files)
	if err != nil {
		log.Fatalf("%v", err)
	}

	names := map[string][]string{}
	lowers := map[string]bool{}
	hasDupes := false
	for _, f := range fns {
		low := strings.ToLower(f.Name)
		if lowers[low] {
			hasDupes = true
		}
		lowers[low] = true
		names[low] = append(names[low], f.Name)
	}
	if hasDupes {
		fmt.Println("Build targets must be case insensitive, thus the follow targets conflict:")
		for _, v := range names {
			if len(v) > 1 {
				fmt.Println("  " + strings.Join(v, ", "))
			}
		}
		os.Exit(1)
	}

	if err := compile(out, fns, files); err != nil {
		log.Fatal(err)
	}

	os.Exit(run(out, flag.Args()...))
}

type data struct {
	Funcs      []parse.Function
	Default    parse.Function
	HasDefault bool
}

func magefiles() []string {
	ctx := build.Default
	ctx.RequiredTags = []string{"mage"}
	ctx.BuildTags = []string{"mage"}
	p, err := ctx.ImportDir(".", 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			log.Fatal("No files marked with the mage build tag in this directory.")
		}
		log.Fatalf("%+v", err)
	}
	return p.GoFiles
}

func compile(out string, funcs []parse.Function, gofiles []string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0700); err != nil {
		return errors.WithMessage(err, "can't create cachedir")
	}
	f, err := os.Create(mainfile)
	if err != nil {
		return errors.WithMessage(err, "can't create mainfile")
	}
	defer os.Remove(mainfile)
	defer f.Close()

	data := data{
		Funcs: funcs,
	}

	for _, f := range funcs {
		if strings.ToLower(f.Name) == "build" {
			data.Default = f
			data.HasDefault = true
		}
	}

	if err := output.Execute(f, data); err != nil {
		return errors.WithMessage(err, "can't execute mainfile template")
	}
	// close is idenmpotent, so this is ok
	f.Close()
	if 0 != run("go", append([]string{"build", "-o", out, mainfile}, gofiles...)...) {
		return errors.New("error compiling magefiles")
	}
	if _, err := os.Stat(out); err != nil {
		return errors.New("failed to find compiled magefile")
	}
	return nil
}

func hashStrings(list []string) (string, error) {
	var hashes []string
	for _, s := range list {
		h, err := hash(s)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	sort.Strings(hashes)
	h := sha1.Sum([]byte(strings.Join(hashes, "")))
	return fmt.Sprintf("%x", h), nil
}

func hash(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", errors.WithMessage(err, "can't open input file")
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.WithMessage(err, "can't write data to hash")
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CacheEnv is the environment variable that users may set to change the
// location where mage stores its compiled binaries.
const CacheEnv = "MAGEFILE_CACHE"

// confDir returns the default gorram configuration directory.
func confDir() string {
	d := os.Getenv(CacheEnv)
	if d != "" {
		return d
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), "magoo")
	default:
		return filepath.Join(os.Getenv("HOME"), ".magefile")
	}
}

func run(cmd string, args ...string) int {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	c.Env = os.Environ()
	if verbose {
		c.Env = append(c.Env, "MAGEFILE_VERBOSE=1")
	}
	if list {
		c.Env = append(c.Env, "MAGEFILE_LIST=1")
	}
	if help {
		c.Env = append(c.Env, "MAGEFILE_HELP=1")
	}
	err := c.Run()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return e.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		}
		return 1
	}
	return 0
}

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
	fmt.Println(os.Args)
	switch strings.ToLower(os.Args[1]) {
		{{range .Funcs}}case "{{lower .Name}}":
			{{if .IsError}}if err := {{.Name}}(); err != nil {
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
