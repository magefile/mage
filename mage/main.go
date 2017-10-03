package mage

import (
	"crypto/sha1"
	"errors"
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

	"github.com/magefile/mage/build"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/parse"
)

// mageVer is used when hashing the output binary to ensure that we get a new
// binary if we use a differernt version of mage.
const mageVer = "v0.3"

var output = template.Must(template.New("").Funcs(map[string]interface{}{
	"lower": strings.ToLower,
}).Parse(tpl))
var initOutput = template.Must(template.New("").Parse(mageTpl))

const mainfile = "mage_output_file.go"
const initFile = "magefile.go"

var (
	force, verbose, list, help, mageInit, keep, showVersion bool

	timestamp, commitHash, gitTag string
)

func init() {
	flag.BoolVar(&force, "f", false, "force recreation of compiled magefile")
	flag.BoolVar(&verbose, "v", false, "show verbose output when running mage targets")
	flag.BoolVar(&list, "l", false, "list mage targets in this directory")
	flag.BoolVar(&help, "h", false, "show this help")
	flag.BoolVar(&mageInit, "init", false, "create a starting template if no mage files exist")
	flag.BoolVar(&keep, "keep", false, "keep intermediate mage files around after running")
	flag.BoolVar(&showVersion, "version", false, "show version info for the mage binary")
	flag.Usage = func() {
		fmt.Println("mage [options] [target]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

// Main is the entrypoint for running mage.  It exists external to mage's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's mage's Main.
func Main() int {
	log.SetFlags(0)
	flag.Parse()
	if help && len(flag.Args()) == 0 {
		flag.Usage()
		return 0
	}
	if showVersion {
		fmt.Println("Mage Build Tool", gitTag)
		fmt.Println("Build Date:", timestamp)
		fmt.Println("Commit:", commitHash)
		return 0
	}
	files, err := magefiles()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if len(files) == 0 && mageInit {
		if err := generateInit(); err != nil {
			log.Printf("%+v", err)
			return 1
		}
		log.Println(initFile, "created")
		return 0
	} else if len(files) == 0 {
		log.Print("No .go files marked with the mage build tag in this directory.")
		return 1
	}

	out, err := ExeName(files)

	if err != nil {
		log.Printf("%+v", err)
		return 1
	}

	if !force {
		if _, err := os.Stat(out); err == nil {
			return run(out, flag.Args()...)
		}
	}

	info, err := parse.Package(".", files)
	if err != nil {
		log.Printf("%v", err)
		return 1
	}

	names := map[string][]string{}
	lowers := map[string]bool{}
	hasDupes := false
	for _, f := range info.Funcs {
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

	if err := compile(out, info, files); err != nil {
		log.Print(err)
		return 1
	}
	return run(out, flag.Args()...)
}

type data struct {
	Funcs        []parse.Function
	DefaultError bool
	Default      string
	DefaultFunc  parse.Function
}

func magefiles() ([]string, error) {
	ctx := build.Default
	ctx.RequiredTags = []string{"mage"}
	ctx.BuildTags = []string{"mage"}
	p, err := ctx.ImportDir(".", 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			return []string{}, nil
		}
		return nil, err
	}
	return p.GoFiles, nil
}

func compile(out string, info *parse.PkgInfo, gofiles []string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0700); err != nil {
		return fmt.Errorf("can't create cachedir: %v", err)
	}
	f, err := os.Create(mainfile)
	if err != nil {
		return fmt.Errorf("can't create mainfile: %v", err)
	}
	if !keep {
		defer os.Remove(mainfile)
	}
	defer f.Close()

	data := data{
		Funcs:   info.Funcs,
		Default: info.DefaultName,
		DefaultFunc: info.DefaultFunc,
	}

	data.DefaultError = info.DefaultIsError

	if err := output.Execute(f, data); err != nil {
		return fmt.Errorf("can't execute mainfile template: %v", err)
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

// ExeName reports the executable filename that this version of Mage would
// create for the given magefiles.
func ExeName(files []string) (string, error) {
	var hashes []string
	for _, s := range files {
		h, err := hashFile(s)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	// hash the mainfile template to ensure if it gets updated, we make a new
	// binary.
	hashes = append(hashes, fmt.Sprintf("%x", sha1.Sum([]byte(tpl))))
	sort.Strings(hashes)
	hash := sha1.Sum([]byte(strings.Join(hashes, "") + mageVer))
	filename := fmt.Sprintf("%x", hash)

	out := filepath.Join(mg.CacheDir(), filename)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	return out, nil
}

func hashFile(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("can't open input file: %v", err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("can't write data to hash: %v", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generateInit() error {
	f, err := os.Create(initFile)
	if err != nil {
		return fmt.Errorf("could not create mage template: %v", err)
	}
	defer f.Close()

	if err := initOutput.Execute(f, nil); err != nil {
		return fmt.Errorf("can't execute magefile template: %v", err)
	}

	return nil
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
