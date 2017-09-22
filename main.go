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

func main() {
	log.SetFlags(0)
	force := false
	flag.BoolVar(&force, "f", false, "force recreation of compiled magefile")
	flag.Parse()

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
			run(out, os.Args[1:]...)
			return
		}
	}

	fns, err := parse.Package(".", files)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := compile(out, fns, files); err != nil {
		log.Fatal(err)
	}

	os.Exit(run(out, flag.Args()...))
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
	if err := output.Execute(f, funcs); err != nil {
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

var tpl = `
// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) == 1 {
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 0, 0, 0)
		log.Println("Targets: ")
		{{- range .}}
		fmt.Fprintln(w, "  {{.Name}}\t{{.Comment}}")
		{{- end}}
		if err := w.Flush(); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch strings.ToLower(os.Args[1]) {
		{{range .}}case "{{lower .Name}}":
			{{if .IsError}}if err := {{.Name}}(); err != nil {
				log.Fatal(err)
			}
			{{else -}}
			{{.Name}}()
			{{- end}}
		{{end}}
		default:
			log.Fatalf("unknown target %q", os.Args[1])
	}
}
`

func run(cmd string, args ...string) int {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err := c.Run()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return e.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		}
		return 1
	}
	return 0
}
