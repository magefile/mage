package mage

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

// mageVer is used when hashing the output binary to ensure that we get a new
// binary if we use a differernt version of mage.
const mageVer = "v0.3"

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

func Main() {
	log.SetFlags(0)
	flag.Parse()
	if help && len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	files := magefiles()
	filename, err := exeName(files, mageVer)
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
			os.Exit(run(out, flag.Args()...))
		}
	}

	info, err := parse.Package(".", files)
	if err != nil {
		log.Fatalf("%v", err)
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
		log.Fatal(err)
	}
	os.Exit(run(out, flag.Args()...))
}

type data struct {
	Funcs        []parse.Function
	DefaultError bool
	Default      string
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

func compile(out string, info *parse.PkgInfo, gofiles []string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0700); err != nil {
		return errors.WithMessage(err, "can't create cachedir")
	}
	f, err := os.Create(mainfile)
	if err != nil {
		return errors.WithMessage(err, "can't create mainfile")
	}
	//defer os.Remove(mainfile)
	defer f.Close()

	data := data{
		Funcs:   info.Funcs,
		Default: info.DefaultName,
	}

	data.DefaultError = info.DefaultIsError

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

func exeName(files []string, ver string) (string, error) {
	var hashes []string
	for _, s := range files {
		h, err := hash(s)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	sort.Strings(hashes)
	h := sha1.Sum([]byte(strings.Join(hashes, "") + ver))
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
