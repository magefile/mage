package mage

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/parse"
	"github.com/magefile/mage/sh"
)

// magicRebuildKey is used when hashing the output binary to ensure that we get
// a new binary even if nothing in the input files or generated mainfile has
// changed. This can be used when we change how we parse files, or otherwise
// change the inputs to the compiling process.
const magicRebuildKey = "v0.3"

var output = template.Must(template.New("").Funcs(map[string]interface{}{
	"lower": strings.ToLower,
	"lowerfirst": func(s string) string {
		parts := strings.Split(s, ":")
		for i, t := range parts {
			r := []rune(t)
			parts[i] = string(unicode.ToLower(r[0])) + string(r[1:])
		}
		return strings.Join(parts, ":")
	},
}).Parse(tpl))
var initOutput = template.Must(template.New("").Parse(mageTpl))

const mainfile = "mage_output_file.go"
const initFile = "magefile.go"

var debug = log.New(ioutil.Discard, "DEBUG: ", log.Ltime|log.Lmicroseconds)

// set by ldflags when you "mage build"
var (
	commitHash = "<not set>"
	timestamp  = "<not set>"
	gitTag     = "<not set>"
)

//go:generate stringer -type=Command

// Command tracks invocations of mage that run without targets or other flags.
type Command int

// The various command types
const (
	None          Command = iota
	Version               // report the current version of mage
	Init                  // create a starting template for mage
	Clean                 // clean out old compiled mage binaries from the cache
	CompileStatic         // compile a static binary of the current directory
)

// Main is the entrypoint for running mage.  It exists external to mage's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's mage's Main.
func Main() int {
	return ParseAndRun(os.Stdout, os.Stderr, os.Stdin, os.Args[1:])
}

// Invocation contains the args for invoking a run of Mage.
type Invocation struct {
	Debug      bool          // turn on debug messages
	Dir        string        // directory to read magefiles from
	Force      bool          // forces recreation of the compiled binary
	Verbose    bool          // tells the magefile to print out log statements
	List       bool          // tells the magefile to print out a list of targets
	Help       bool          // tells the magefile to print out help for a specific target
	Keep       bool          // tells mage to keep the generated main file after compiling
	Timeout    time.Duration // tells mage to set a timeout to running the targets
	CompileOut string        // tells mage to compile a static binary to this path, but not execute
	Stdout     io.Writer     // writer to write stdout messages to
	Stderr     io.Writer     // writer to write stderr messages to
	Stdin      io.Reader     // reader to read stdin from
	Args       []string      // args to pass to the compiled binary
	GoCmd      string        // the go binary command to run
	CacheDir   string        // the directory where we should store compiled binaries
}

// ParseAndRun parses the command line, and then compiles and runs the mage
// files in the given directory with the given args (do not include the command
// name in the args).
func ParseAndRun(stdout, stderr io.Writer, stdin io.Reader, args []string) int {
	errlog := log.New(stderr, "", 0)
	out := log.New(stdout, "", 0)
	inv, cmd, err := Parse(stderr, stdout, args)
	inv.Stderr = stderr
	inv.Stdin = stdin
	if err == flag.ErrHelp {
		return 0
	}
	if err != nil {
		errlog.Println("Error:", err)
		return 2
	}

	switch cmd {
	case Version:
		out.Println("Mage Build Tool", gitTag)
		out.Println("Build Date:", timestamp)
		out.Println("Commit:", commitHash)
		out.Println("built with:", runtime.Version())
		return 0
	case Init:
		if err := generateInit(inv.Dir); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		out.Println(initFile, "created")
		return 0
	case Clean:
		if err := removeContents(inv.CacheDir); err != nil {
			out.Println("Error:", err)
			return 1
		}
		if err := removeContents(filepath.Join(inv.CacheDir, mainfileSubdir)); err != nil {
			out.Println("Error:", err)
			return 1
		}

		out.Println(inv.CacheDir, "cleaned")
		return 0
	case CompileStatic:
		return Invoke(inv)
	case None:
		return Invoke(inv)
	default:
		panic(fmt.Errorf("Unknown command type: %v", cmd))
	}
}

// Parse parses the given args and returns structured data.  If parse returns
// flag.ErrHelp, the calling process should exit with code 0.
func Parse(stderr, stdout io.Writer, args []string) (inv Invocation, cmd Command, err error) {
	inv.Stdout = stdout
	fs := flag.FlagSet{}
	fs.SetOutput(stdout)
	fs.BoolVar(&inv.Force, "f", false, "force recreation of compiled magefile")
	fs.BoolVar(&inv.Debug, "debug", mg.Debug(), "turn on debug messages")
	fs.BoolVar(&inv.Verbose, "v", mg.Verbose(), "show verbose output when running mage targets")
	fs.BoolVar(&inv.List, "l", false, "list mage targets in this directory")
	fs.BoolVar(&inv.Help, "h", false, "show this help")
	fs.DurationVar(&inv.Timeout, "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	fs.BoolVar(&inv.Keep, "keep", false, "keep intermediate mage files around after running")
	fs.StringVar(&inv.Dir, "d", ".", "run magefiles in the given directory")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "show version info for the mage binary")

	var mageInit bool
	fs.BoolVar(&mageInit, "init", false, "create a starting template if no mage files exist")
	var clean bool
	fs.BoolVar(&clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	var compileOutPath string
	fs.StringVar(&compileOutPath, "compile", "", "output a static binary to the given path")
	fs.StringVar(&inv.GoCmd, "gocmd", mg.GoCmd(), "use the given go binary to compile the output")

	fs.Usage = func() {
		fmt.Fprint(stdout, `
mage [options] [target]

Mage is a make-like command runner.  See https://magefile.org for full docs.

Commands:
  -clean    clean out old generated binaries from CACHE_DIR
  -compile <string>
            output a static binary to the given path
  -init     create a starting template if no mage files exist
  -l        list mage targets in this directory
  -h        show this help
  -version  show version info for the mage binary

Options:
  -d <string> 
          run magefiles in the given directory (default ".")
  -debug  turn on debug messages
  -h      show description of a target
  -f      force recreation of compiled magefile
  -keep   keep intermediate mage files around after running
  -gocmd <string>
          use the given go binary to compile the output (default: "go")
  -t <string>
          timeout in duration parsable format (e.g. 5m30s)
  -v      show verbose output when running mage targets
`[1:])
	}
	err = fs.Parse(args)
	if err == flag.ErrHelp {
		// parse will have already called fs.Usage()
		return inv, cmd, err
	}
	if err == nil && inv.Help && len(fs.Args()) == 0 {
		fs.Usage()
		// tell upstream, to just exit
		return inv, cmd, flag.ErrHelp
	}

	numFlags := 0
	switch {
	case mageInit:
		numFlags++
		cmd = Init
	case compileOutPath != "":
		numFlags++
		cmd = CompileStatic
		inv.CompileOut = compileOutPath
		inv.Force = true
	case showVersion:
		numFlags++
		cmd = Version
	case clean:
		numFlags++
		cmd = Clean
		if fs.NArg() > 0 {
			// Temporary dupe of below check until we refactor the other commands to use this check
			return inv, cmd, errors.New("-h, -init, -clean, -compile and -version cannot be used simultaneously")

		}
	}
	if inv.Help {
		numFlags++
	}

	if inv.Debug {
		debug.SetOutput(stderr)
	}

	inv.CacheDir = mg.CacheDir()

	if numFlags > 1 {
		debug.Printf("%d commands defined", numFlags)
		return inv, cmd, errors.New("-h, -init, -clean, -compile and -version cannot be used simultaneously")
	}

	inv.Args = fs.Args()
	if inv.Help && len(inv.Args) > 1 {
		return inv, cmd, errors.New("-h can only show help for a single target")
	}

	return inv, cmd, err
}

// Invoke runs Mage with the given arguments.
func Invoke(inv Invocation) int {
	errlog := log.New(inv.Stderr, "", 0)
	if inv.GoCmd == "" {
		inv.GoCmd = "go"
	}
	if inv.Dir == "" {
		inv.Dir = "."
	}
	if inv.CacheDir == "" {
		inv.CacheDir = mg.CacheDir()
	}

	files, err := Magefiles(inv.Dir, inv.GoCmd, inv.Stderr, inv.Debug)
	if err != nil {
		errlog.Println("Error determining list of magefiles:", err)
		return 1
	}

	if len(files) == 0 {
		errlog.Println("No .go files marked with the mage build tag in this directory.")
		return 1
	}
	debug.Printf("found magefiles: %s", strings.Join(files, ", "))
	exePath, err := ExeName(inv.GoCmd, inv.CacheDir, files)
	if err != nil {
		errlog.Println("Error getting exe name:", err)
		return 1
	}
	if inv.CompileOut != "" {
		exePath = inv.CompileOut
	}
	debug.Println("output exe is ", exePath)

	useCache := false
	if s, err := outputDebug(inv.GoCmd, "env", "GOCACHE"); err == nil {
		// if GOCACHE exists, always rebuild, so we catch transitive
		// dependencies that have changed.
		if s != "" {
			debug.Println("build cache exists, will ignore any compiled binary")
			useCache = true
		}
	}

	if !useCache {
		_, err = os.Stat(exePath)
		switch {
		case err == nil:
			if inv.Force {
				debug.Println("ignoring existing executable")
			} else {
				debug.Println("Running existing exe")
				return RunCompiled(inv, exePath, errlog)
			}
		case os.IsNotExist(err):
			debug.Println("no existing exe, creating new")
		default:
			debug.Printf("error reading existing exe at %v: %v", exePath, err)
			debug.Println("creating new exe")
		}
	}

	// parse wants dir + filenames... arg
	fnames := make([]string, 0, len(files))
	for i := range files {
		fnames = append(fnames, filepath.Base(files[i]))
	}
	if inv.Debug {
		parse.EnableDebug()
	}
	debug.Println("parsing files")
	info, err := parse.Package(inv.Dir, fnames)
	if err != nil {
		errlog.Println("Error parsing magefiles:", err)
		return 1
	}

	main := filepath.Join(inv.Dir, mainfile)
	err = GenerateMainfile(main, inv.CacheDir, info)
	if err != nil {
		errlog.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		defer os.RemoveAll(main)
	}
	files = append(files, main)
	if err := Compile(inv.Dir, inv.GoCmd, exePath, files, inv.Debug, inv.Stderr, inv.Stdout); err != nil {
		errlog.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		// move aside this file before we run the compiled version, in case the
		// compiled file screws things up.  Yes this doubles up with the above
		// defer, that's ok.
		os.RemoveAll(main)
	} else {
		debug.Print("keeping mainfile")
	}

	if inv.CompileOut != "" {
		return 0
	}

	return RunCompiled(inv, exePath, errlog)
}

func moveMainToCache(cachedir, main, hash string) {
	debug.Println("moving main file to cache:", cachedMainfile(cachedir, hash))
	if err := os.Rename(main, cachedMainfile(cachedir, hash)); err != nil && !os.IsNotExist(err) {
		debug.Println("error caching main file: ", err)
	}
}

type data struct {
	Description  string
	Funcs        []parse.Function
	DefaultError bool
	Default      string
	DefaultFunc  parse.Function
	Aliases      map[string]string
}

func goVersion(gocmd string) (string, error) {
	cmd := exec.Command(gocmd, "version")
	out, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout = out
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if s := stderr.String(); s != "" {
			return "", fmt.Errorf("failed to run `go version`: %s", s)
		}
		return "", fmt.Errorf("failed to run `go version`: %v", err)
	}
	return out.String(), nil
}

// Magefiles returns the list of magefiles in dir.
func Magefiles(magePath, goCmd string, stderr io.Writer, isDebug bool) ([]string, error) {
	start := time.Now()
	defer func() {
		debug.Println("time to scan for Magefiles:", time.Since(start))
	}()
	fail := func(err error) ([]string, error) {
		return nil, err
	}

	debug.Println("getting all non-mage files in", magePath)
	// // first, grab all the files with no build tags specified.. this is actually
	// // our exclude list of things without the mage build tag.
	cmd := exec.Command(goCmd, "list", "-e", "-f", `{{join .GoFiles "||"}}`)

	if isDebug {
		cmd.Stderr = stderr
	}
	cmd.Dir = magePath
	b, err := cmd.Output()
	if err != nil {
		return fail(fmt.Errorf("failed to list non-mage gofiles: %v", err))
	}
	list := strings.TrimSpace(string(b))
	debug.Println("found non-mage files", list)
	exclude := map[string]bool{}
	for _, f := range strings.Split(list, "||") {
		if f != "" {
			debug.Printf("marked file as non-mage: %q", f)
			exclude[f] = true
		}
	}
	debug.Println("getting all files plus mage files")
	cmd = exec.Command(goCmd, "list", "-tags=mage", "-e", "-f", `{{join .GoFiles "||"}}`)
	if isDebug {
		cmd.Stderr = stderr
	}
	cmd.Dir = magePath
	b, err = cmd.Output()
	if err != nil {
		return fail(fmt.Errorf("failed to list mage gofiles: %v", err))
	}

	list = strings.TrimSpace(string(b))
	files := []string{}
	for _, f := range strings.Split(list, "||") {
		if f != "" && !exclude[f] {
			files = append(files, f)
		}
	}
	for i := range files {
		files[i] = filepath.Join(magePath, files[i])
	}
	return files, nil
}

// Compile uses the go tool to compile the files into an executable at path.
func Compile(magePath, goCmd, compileTo string, gofiles []string, isDebug bool, stderr, stdout io.Writer) error {
	debug.Println("compiling to", compileTo)
	debug.Println("compiling using gocmd:", goCmd)
	if isDebug {
		runDebug(goCmd, "version")
		runDebug(goCmd, "env")
	}

	// strip off the path since we're setting the path in the build command
	for i := range gofiles {
		gofiles[i] = filepath.Base(gofiles[i])
	}
	debug.Printf("running %s build -o %s %s", goCmd, compileTo, strings.Join(gofiles, " "))
	c := exec.Command(goCmd, append([]string{"build", "-o", compileTo}, gofiles...)...)
	c.Env = os.Environ()
	c.Stderr = stderr
	c.Stdout = stdout
	c.Dir = magePath
	start := time.Now()
	err := c.Run()
	debug.Println("time to compile Magefile:", time.Since(start))
	if err != nil {
		return errors.New("error compiling magefiles")
	}
	return nil
}

func runDebug(cmd string, args ...string) error {
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	debug.Println("running", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	c.Stderr = errbuf
	c.Stdout = buf
	if err := c.Run(); err != nil {
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errbuf)
		return err
	}
	debug.Println(buf)
	return nil
}

func outputDebug(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	debug.Println("running", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	c.Stderr = errbuf
	c.Stdout = buf
	if err := c.Run(); err != nil {
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errbuf)
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

const mainfileSubdir = "mainfiles"

func cachedMainfile(cachedir, hash string) string {
	return filepath.Join(cachedir, mainfileSubdir, hash)
}

// GenerateMainfile generates the mage mainfile at path.  Because go build's
// cache cares about modtimes, we squirrel away our mainfiles in the cachedir
// and move them back as needed.
func GenerateMainfile(path, cachedir string, info *parse.PkgInfo) error {
	debug.Println("Creating mainfile at", path)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating generated mainfile: %v", err)
	}
	defer f.Close()
	data := data{
		Description: info.Description,
		Funcs:       info.Funcs,
		Default:     info.DefaultName,
		DefaultFunc: info.DefaultFunc,
		Aliases:     info.Aliases,
	}

	data.DefaultError = info.DefaultIsError

	debug.Println("writing new file at", path)
	if err := output.Execute(f, data); err != nil {
		return fmt.Errorf("can't execute mainfile template: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing generated mainfile: %v", err)
	}
	// we set an old modtime on the generated mainfile so that the go tool
	// won't think it has changed more recently than the compiled binary.
	longAgo := time.Now().Add(-time.Hour * 24 * 365 * 10)
	if err := os.Chtimes(path, longAgo, longAgo); err != nil {
		return fmt.Errorf("error setting old modtime on generated mainfile: %v", err)
	}
	return nil
}

// ExeName reports the executable filename that this version of Mage would
// create for the given magefiles.
func ExeName(goCmd, cacheDir string, files []string) (string, error) {
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
	ver, err := goVersion(goCmd)
	if err != nil {
		return "", err
	}
	hash := sha1.Sum([]byte(strings.Join(hashes, "") + magicRebuildKey + ver))
	filename := fmt.Sprintf("%x", hash)

	out := filepath.Join(cacheDir, filename)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	return out, nil
}

func hashFile(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("can't open input file for hashing: %#v", err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("can't write data to hash: %v", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generateInit(dir string) error {
	debug.Println("generating default magefile in", dir)
	f, err := os.Create(filepath.Join(dir, initFile))
	if err != nil {
		return fmt.Errorf("could not create mage template: %v", err)
	}
	defer f.Close()

	if err := initOutput.Execute(f, nil); err != nil {
		return fmt.Errorf("can't execute magefile template: %v", err)
	}

	return nil
}

// RunCompiled runs an already-compiled mage command with the given args,
func RunCompiled(inv Invocation, exePath string, errlog *log.Logger) int {
	debug.Println("running binary", exePath)
	c := exec.Command(exePath, inv.Args...)
	c.Stderr = inv.Stderr
	c.Stdout = inv.Stdout
	c.Stdin = inv.Stdin
	c.Dir = inv.Dir
	c.Env = os.Environ()
	if inv.Verbose {
		c.Env = append(c.Env, "MAGEFILE_VERBOSE=1")
	}
	if inv.List {
		c.Env = append(c.Env, "MAGEFILE_LIST=1")
	}
	if inv.Help {
		c.Env = append(c.Env, "MAGEFILE_HELP=1")
	}
	if inv.Debug {
		c.Env = append(c.Env, "MAGEFILE_DEBUG=1")
	}
	if inv.Timeout > 0 {
		c.Env = append(c.Env, fmt.Sprintf("MAGEFILE_TIMEOUT=%s", inv.Timeout.String()))
	}
	debug.Print("running magefile with mage vars:\n", strings.Join(filter(c.Env, "MAGEFILE"), "\n"))
	err := c.Run()
	if !sh.CmdRan(err) {
		errlog.Printf("failed to run compiled magefile: %v", err)
	}
	return sh.ExitStatus(err)
}

func filter(list []string, prefix string) []string {
	var out []string
	for _, s := range list {
		if strings.HasPrefix(s, prefix) {
			out = append(out, s)
		}
	}
	return out
}

func removeContents(dir string) error {
	debug.Println("removing all files in", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
