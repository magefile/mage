package sh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
)

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
	return func(args2 ...string) error {
		return Run(cmd, append(args, args2...)...)
	}
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
	return RunWith(nil, cmd, args...)
}

// RunV is like Run, but always sends the command's stdout to os.Stdout.
func RunV(cmd string, args ...string) error {
	_, err := Exec(nil, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// RunWith runs the given command, directing stderr to this program's stderr and
// printing stdout to stdout if mage was run with -v.  It adds adds env to the
// environment variables for the command being run. Environment variables should
// be in the format name=value.
func RunWith(env map[string]string, cmd string, args ...string) error {
	var output io.Writer
	if mg.Verbose() {
		output = os.Stdout
	}
	_, err := Exec(env, output, os.Stderr, cmd, args...)
	return err
}

// RunWithV is like RunWith, but always sends the command's stdout to os.Stdout.
func RunWithV(env map[string]string, cmd string, args ...string) error {
	_, err := Exec(env, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// Output runs the command and returns the text from stdout.
func Output(cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(nil, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// OutputWith is like RunWith, but returns what is written to stdout.
func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(env, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// Exec executes the command, piping its stdout and stderr to the given writers.
// If the command fails, it will return an error that, if returned from a target
// or mg.Deps call, will cause mage to exit with the same code as the command
// failed with.
// Env is a list of environment variables to set when running the command,
// these override the current environment variables set (which are also passed
// to the command).
// cmd and args may include references to environment variables in $FOO format,
// in which case these will be expanded before the command is run.
//
// Ran reports if the command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran. If err == nil, ran
// is always true and code is always 0.
func Exec(env map[string]string, stdout, stderr io.Writer, cmd string, args ...string) (ran bool, err error) {
	return Command{
		Cmd:    cmd,
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
		Env:    env,
	}.Exec(context.Background())
}

// Command is a command to be executed.
//
// Both Path and Args may include references to environment variables in $FOO
// format, // in which case these will be expanded before the command is run.
type Command struct {
	// Cmd is the path of the command to execute.
	// Relative paths are evaluated with respect to WorkingDir.
	//
	// Environment variable references of the form $FOO will be expanded before
	// the command is run.
	Cmd string
	// Args are the command line arguments to pass to the command.
	//
	// Environment variable references of the form $FOO will be expanded before
	// the command is run.
	Args []string

	// Env is a list of environment variables to set when running the command.
	// These override the current environment variables set (which are also
	// passed to the command).
	Env map[string]string

	// Stdout is the command's stdout stream.
	Stdout io.Writer
	// Stderr is the command's stderr stream.
	Stderr io.Writer

	// WorkingDir specifies the working directory this will command execute in.
	// An empty string indicates the command should run in the current working
	// directory.
	WorkingDir string
}

// Output and Exec use value receivers to avoid race conditions when modifying
// the Command's internal state during shell expansion.

// Output runs the [Command] and returns the text from stdout.
//
// See [Command.Exec] for more detailts.
func (cmd Command) Output(ctx context.Context) (string, error) {
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	_, err := cmd.Exec(ctx)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

// Exec executes the [Command] using the provided [context.Context] for
// cancellation, and piping the stdout and stderr to the given writers.
// If the command fails, it will return an error that, if returned from a target
// or [mg.Deps] call, will cause mage to exit with the same code as the command
// failed with.
//
// Ran reports if the Command ran (rather than was not found or not executable).
// Code reports the exit code the command returned if it ran, and can be
// retrieved from err with [mg.ExitStatus].
// If err == nil, ran is always true and code is always 0.
func (cmd Command) Exec(ctx context.Context) (ran bool, err error) {
	expand := func(s string) string {
		s2, ok := cmd.Env[s]
		if ok {
			return s2
		}
		return os.Getenv(s)
	}
	cmd.Cmd = os.Expand(cmd.Cmd, expand)
	for i := range cmd.Args {
		cmd.Args[i] = os.Expand(cmd.Args[i], expand)
	}
	ran, code, err := cmd.run(ctx)
	if err == nil {
		return true, nil
	}
	if ran {
		return ran, mg.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd.Cmd, strings.Join(cmd.Args, " "), code)
	}
	return ran, fmt.Errorf(`failed to run "%s %s: %v"`, cmd.Cmd, strings.Join(cmd.Args, " "), err)

}

func (cmd *Command) run(ctx context.Context) (ran bool, code int, err error) {
	c := exec.CommandContext(ctx, cmd.Cmd, cmd.Args...)
	env := os.Environ()
	for k, v := range cmd.Env {
		env = append(c.Env, k+"="+v)
	}
	c.Env = env
	c.Stderr = cmd.Stderr
	c.Stdout = cmd.Stdout
	c.Stdin = os.Stdin
	c.Dir = cmd.WorkingDir

	quoted := make([]string, 0, len(cmd.Args))
	for _, c := range cmd.Args {
		quoted = append(quoted, fmt.Sprintf("%q", c))
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
