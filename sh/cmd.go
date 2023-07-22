package sh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
)

// runOptions is a set of options to be applied with ExecSh.
type runOptions struct {
	cmd            string
	args           []string
	dir            string
	env            map[string]string
	stderr, stdout io.Writer
}

// RunOpt applies an option to a runOptions set.
type RunOpt func(*runOptions)

// WithV sets stderr and stdout the standard streams
func WithV() RunOpt {
	return func(options *runOptions) {
		options.stdout = os.Stdout
		options.stderr = os.Stderr
	}
}

// WithEnv sets the env passed in env vars.
func WithEnv(env map[string]string) RunOpt {
	return func(options *runOptions) {
		if options.env == nil {
			options.env = make(map[string]string)
		}
		for k, v := range env {
			options.env[k] = v
		}
	}
}

// WithStderr sets the stderr stream.
func WithStderr(w io.Writer) RunOpt {
	return func(options *runOptions) {
		options.stderr = w
	}
}

// WithStdout sets the stdout stream.
func WithStdout(w io.Writer) RunOpt {
	return func(options *runOptions) {
		options.stdout = w
	}
}

// WithDir sets the working directory for the command.
func WithDir(dir string) RunOpt {
	return func(options *runOptions) {
		options.dir = dir
	}
}

// WithArgs appends command arguments.
func WithArgs(args ...string) RunOpt {
	return func(options *runOptions) {
		if options.args == nil {
			options.args = make([]string, 0, len(args))
		}
		options.args = append(options.args, args...)
	}
}

// RunSh returns a function that calls ExecSh, only returning errors.
func RunSh(cmd string, options ...RunOpt) func(args ...string) error {
	run := ExecSh(cmd, options...)
	return func(args ...string) error {
		_, err := run(args...)
		return err
	}
}

// ExecSh returns a function that executes the command, piping its stdout and
// stderr according to the config options. If the command fails, it will return
// an error that, if returned from a target or mg.Deps call, will cause mage to
// exit with the same code as the command failed with.
//
// ExecSh takes a variable list of RunOpt objects to configure how the command
// is executed. See RunOpt docs for more details.
//
// Env vars configured on the command override the current environment variables
// set (which are also passed to the command). The cmd and args may include
// references to environment variables in $FOO format, in which case these will be
// expanded before the command is run.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func ExecSh(cmd string, options ...RunOpt) func(args ...string) (bool, error) {
	opts := runOptions{
		cmd: cmd,
	}
	for _, o := range options {
		o(&opts)
	}

	if opts.stdout == nil && mg.Verbose() {
		opts.stdout = os.Stdout
	}

	return func(args ...string) (bool, error) {
		expand := func(s string) string {
			s2, ok := opts.env[s]
			if ok {
				return s2
			}
			return os.Getenv(s)
		}
		cmd = os.Expand(cmd, expand)
		finalArgs := append(opts.args, args...)
		for i := range finalArgs {
			finalArgs[i] = os.Expand(finalArgs[i], expand)
		}
		ran, code, err := run(opts.dir, opts.env, opts.stdout, opts.stderr, cmd, finalArgs...)

		if err == nil {
			return ran, nil
		}
		if ran {
			return ran, mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
		}
		return ran, fmt.Errorf(`failed to run "%s %s: %v"`, cmd, strings.Join(args, " "), err)
	}
}

// RunCmd returns a function that will call Run with the given command. This is
// useful for creating command aliases to make your scripts easier to read, like
// this:
//
//	 // in a helper file somewhere
//	 var g0 = sh.RunCmd("go")  // go is a keyword :(
//
//	 // somewhere in your main code
//		if err := g0("install", "github.com/gohugo/hugo"); err != nil {
//			return err
//	 }
//
// Args passed to command get baked in as args to the command when you run it.
// Any args passed in when you run the returned function will be appended to the
// original args.  For example, this is equivalent to the above:
//
//	var goInstall = sh.RunCmd("go", "install") goInstall("github.com/gohugo/hugo")
//
// RunCmd uses Exec underneath, so see those docs for more details.
func RunCmd(cmd string, args ...string) func(args ...string) error {
	return RunSh(cmd, WithArgs(args...))
}

// OutCmd is like RunCmd except the command returns the output of the
// command.
func OutCmd(cmd string, args ...string) func(args ...string) (string, error) {
	return func(args2 ...string) (string, error) {
		return Output(cmd, append(args, args2...)...)
	}
}

// Run is like RunWith, but doesn't specify any environment variables.
func Run(cmd string, args ...string) error {
	return RunSh(cmd, WithArgs(args...))()
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func RunV(cmd string, args ...string) error {
	return RunSh(cmd, WithV(), WithArgs(args...))()
}

// RunWith runs the given command, directing stderr to this program's stderr and
// printing stdout to stdout if mage was run with -v.  It adds adds env to the
// environment variables for the command being run. Environment variables should
// be in the format name=value.
func RunWith(env map[string]string, cmd string, args ...string) error {
	return RunSh(cmd, WithEnv(env), WithArgs(args...))()
}

// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	return RunSh(cmd, WithV(), WithEnv(env), WithArgs(args...))()
}

// Output runs the command and returns the text from stdout.
func Output(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	err := RunSh(cmd, WithStderr(os.Stderr), WithStdout(buf), WithArgs(args...))()
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// OutputWith is like RunWith, but returns what is written to stdout.
func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	err := RunSh(cmd, WithEnv(env), WithStderr(os.Stderr), WithStdout(buf), WithArgs(args...))()
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// Exec executes the command, piping its stdout and stderr to the given
// writers. If the command fails, it will return an error that, if returned
// from a target or mg.Deps call, will cause mage to exit with the same code as
// the command failed with. Env is a list of environment variables to set when
// running the command, these override the current environment variables set
// (which are also passed to the command). cmd and args may include references
// to environment variables in $FOO format, in which case these will be
// expanded before the command is run.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	return ExecSh(cmd, WithArgs(args...), WithStderr(stderr), WithStdout(stdout), WithEnv(env))()
}

func run(dir string, env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, code int, err error) {
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
	}
	c.Dir = dir
	c.Stderr = stderr
	c.Stdout = stdout
	c.Stdin = os.Stdin

	var quoted []string
	for i := range args {
		quoted = append(quoted, fmt.Sprintf("%q", args[i]))
	}
	// To protect against logging from doing exec in global variables
	if mg.Verbose() {
		log.Println("exec:", cmd, strings.Join(quoted, " "))
	}
	err = c.Run()
	return CmdRan(err), ExitStatus(err), err
}

// CmdRan examines the error to determine if it was generated as a result of a
// command running via os/exec.Command.  If the error is nil, or the command ran
// (even if it exited with a non-zero exit code), CmdRan reports true.  If the
// error is an unrecognized type, or it is an error from exec.Command that says
// the command failed to run (usually due to the command not existing or not
// being executable), it reports false.
func CmdRan(err error) bool {
	if err == nil {
		return true
	}
	ee, ok := err.(*exec.ExitError)
	if ok {
		return ee.Exited()
	}
	return false
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or 1 if it is a different error.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(exitStatus); ok {
		return e.ExitStatus()
	}
	if e, ok := err.(*exec.ExitError); ok {
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}
	return 1
}
