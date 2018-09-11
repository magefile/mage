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
	"regexp"
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

var debug = log.New(ioutil.Discard, "DEBUG: ", 0)

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
	return ParseAndRun(".", os.Stdout, os.Stderr, os.Stdin, os.Args[1:])
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
}

// ParseAndRun parses the command line, and then compiles and runs the mage
// files in the given directory with the given args (do not include the command
// name in the args).
func ParseAndRun(dir string, stdout, stderr io.Writer, stdin io.Reader, args []string) int {
	log := log.New(stderr, "", 0)
	inv, cmd, err := Parse(stderr, stdout, args)
	inv.Dir = dir
	inv.Stderr = stderr
	inv.Stdin = stdin
	if err == flag.ErrHelp {
		return 0
	}
	if err != nil {
		log.Println("Error:", err)
		return 2
	}

	switch cmd {
	case Version:
		log.Println("Mage Build Tool", gitTag)
		log.Println("Build Date:", timestamp)
		log.Println("Commit:", commitHash)
		log.Println("built with:", runtime.Version())
		return 0
	case Init:
		if err := generateInit(dir); err != nil {
			log.Println("Error:", err)
			return 1
		}
		log.Println(initFile, "created")
		return 0
	case Clean:
		dir := mg.CacheDir()
		if err := removeContents(dir); err != nil {
			log.Println("Error:", err)
			return 1
		}
		log.Println(dir, "cleaned")
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
	fs.BoolVar(&inv.Debug, "debug", false, "turn on debug messages (implies -keep)")
	fs.BoolVar(&inv.Force, "f", false, "force recreation of compiled magefile")
	fs.BoolVar(&inv.Verbose, "v", false, "show verbose output when running mage targets")
	fs.BoolVar(&inv.List, "l", false, "list mage targets in this directory")
	fs.BoolVar(&inv.Help, "h", false, "show this help")
	fs.DurationVar(&inv.Timeout, "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	fs.BoolVar(&inv.Keep, "keep", false, "keep intermediate mage files around after running")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "show version info for the mage binary")

	var mageInit bool
	fs.BoolVar(&mageInit, "init", false, "create a starting template if no mage files exist")
	var clean bool
	fs.BoolVar(&clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	var compileOutPath string
	fs.StringVar(&compileOutPath, "compile", "", "output a static binary to the given path")
	var goCmd string
	fs.StringVar(&goCmd, "gocmd", "", "use the given go binary to compile the output")

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
  -debug  turn on debug messages (implies -keep)
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

	if !inv.Debug {
		inv.Debug = mg.Debug()
	} else {
		os.Setenv(mg.DebugEnv, "1")
	}
	if inv.Debug {
		debug.SetOutput(stderr)
		inv.Keep = true
	}

	// If verbose is still false, we're going to peek at the environment variable to see if
	// MAGE_VERBOSE has been set. If so, we're going to use it for the value of MAGE_VERBOSE.
	if !inv.Verbose {
		inv.Verbose = mg.Verbose()
	}

	if goCmd != "" {
		if err := os.Setenv(mg.GoCmdEnv, goCmd); err != nil {
			return inv, cmd, fmt.Errorf("failed to set gocmd: %v", err)
		}
	}

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
	log := log.New(inv.Stderr, "", 0)

	files, err := Magefiles(inv)
	if err != nil {
		log.Println("Error determining list of magefiles:", err)
		return 1
	}

	if len(files) == 0 {
		log.Println("No .go files marked with the mage build tag in this directory.")
		return 1
	}
	debug.Printf("found magefiles: %s", strings.Join(files, ", "))
	exePath, err := ExeName(files)
	if err != nil {
		log.Println("Error getting exe name:", err)
		return 1
	}
	if inv.CompileOut != "" {
		exePath = inv.CompileOut
	}
	debug.Println("output exe is ", exePath)

	_, err = os.Stat(exePath)
	switch {
	case err == nil:
		if inv.Force {
			debug.Println("ignoring existing executable")
		} else {
			debug.Println("Running existing exe")
			return RunCompiled(inv, exePath)
		}
	case os.IsNotExist(err):
		debug.Println("no existing exe, creating new")
	default:
		debug.Printf("error reading existing exe at %v: %v", exePath, err)
		debug.Println("creating new exe")
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
		log.Println("Error parsing magefiles:", err)
		return 1
	}

	main := filepath.Join(inv.Dir, mainfile)
	if err := GenerateMainfile(main, info); err != nil {
		log.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		defer func() {
			debug.Println("removing main file")
			if err := os.Remove(main); err != nil && !os.IsNotExist(err) {
				debug.Println("error removing main file: ", err)
			}
		}()
	}
	files = append(files, main)
	if err := Compile(inv, exePath, files); err != nil {
		log.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		// remove this file before we run the compiled version, in case the
		// compiled file screws things up.  Yes this doubles up with the above
		// defer, that's ok.
		if err := os.Remove(main); err != nil && !os.IsNotExist(err) {
			debug.Println("error removing main file: ", err)
		}
	} else {
		debug.Print("keeping mainfile")
	}

	if inv.CompileOut != "" {
		return 0
	}

	return RunCompiled(inv, exePath)
}

type data struct {
	Description  string
	Funcs        []parse.Function
	DefaultError bool
	Default      string
	DefaultFunc  parse.Function
	Aliases      map[string]string
}

// Yeah, if we get to go2, this will need to be revisited. I think that's ok.
var goVerReg = regexp.MustCompile(`1\.[0-9]+`)

func goVersion() (string, error) {
	cmd := exec.Command(mg.GoCmd(), "version")
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
func Magefiles(inv Invocation) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return nil, err
	}

	debug.Println("getting all non-mage files in", inv.Dir)
	// // first, grab all the files with no build tags specified.. this is actually
	// // our exclude list of things without the mage build tag.
	cmd := exec.Command(mg.GoCmd(), "list", "-e", "-f", `{{join .GoFiles "||"}}`)

	if inv.Debug {
		cmd.Stderr = inv.Stderr
	}
	cmd.Dir = inv.Dir
	b, err := cmd.Output()
	if err != nil {
		fail(fmt.Errorf("failed to list non-mage gofiles: %v", err))
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
	cmd = exec.Command(mg.GoCmd(), "list", "-tags=mage", "-e", "-f", `{{join .GoFiles "||"}}`)
	if inv.Debug {
		cmd.Stderr = inv.Stderr
	}
	cmd.Dir = inv.Dir
	b, err = cmd.Output()
	list = strings.TrimSpace(string(b))

	if err != nil {
		fail(fmt.Errorf("failed to list mage gofiles: %v", err))
	}
	files := []string{}
	for _, f := range strings.Split(list, "||") {
		if f != "" && !exclude[f] {
			files = append(files, f)
		}
	}
	for i := range files {
		files[i] = filepath.Join(inv.Dir, files[i])
	}
	return files, nil
}

// Compile uses the go tool to compile the files into an executable at path.
func Compile(inv Invocation, path string, gofiles []string) error {
	debug.Println("compiling to", path)
	debug.Println("compiling using gocmd:", mg.GoCmd())
	if inv.Debug {
		runDebug(mg.GoCmd(), "version")
		runDebug(mg.GoCmd(), "env")
	}

	// strip off the path since we're setting the path in the build command
	for i := range gofiles {
		gofiles[i] = filepath.Base(gofiles[i])
	}
	debug.Printf("running %s build -o %s %s", mg.GoCmd(), path, strings.Join(gofiles, ", "))
	c := exec.Command(mg.GoCmd(), append([]string{"build", "-o", path}, gofiles...)...)
	c.Env = os.Environ()
	c.Stderr = inv.Stderr
	c.Stdout = inv.Stdout
	c.Dir = inv.Dir
	err := c.Run()
	if err != nil {
		return errors.New("error compiling magefiles")
	}
	if _, err := os.Stat(path); err != nil {
		return errors.New("failed to find compiled magefile")
	}
	return nil
}

func runDebug(cmd string, args ...string) {
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	debug.Println("running", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	c.Stderr = errbuf
	c.Stdout = buf
	if err := c.Run(); err != nil {
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errbuf)
	}
	debug.Println(buf)
}

// GenerateMainfile creates the mainfile at path with the info from
func GenerateMainfile(path string, info *parse.PkgInfo) error {
	debug.Println("Creating mainfile at", path)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("can't create mainfile: %v", err)
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

	if err := output.Execute(f, data); err != nil {
		return fmt.Errorf("can't execute mainfile template: %v", err)
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
	ver, err := goVersion()
	if err != nil {
		return "", err
	}
	hash := sha1.Sum([]byte(strings.Join(hashes, "") + magicRebuildKey + ver))
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
func RunCompiled(inv Invocation, exePath string) int {
	debug.Println("running binary", exePath)
	c := exec.Command(exePath, inv.Args...)
	c.Stderr = inv.Stderr
	c.Stdout = inv.Stdout
	c.Stdin = inv.Stdin
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
	if inv.Timeout > 0 {
		c.Env = append(c.Env, fmt.Sprintf("MAGEFILE_TIMEOUT=%s", inv.Timeout.String()))
	}
	debug.Print("running magefile with mage vars:\n", strings.Join(filter(c.Env, "MAGEFILE"), "\n"))
	return sh.ExitStatus(c.Run())
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
